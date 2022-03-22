package cmd

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"

	"aping/lib"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	count     int
	interval  int
	coroutine int

	sem       *lib.Semaphore
	ErrorChan chan error
	DoneChan  chan bool
)

type ParameterError struct {
	message string
}

func (p *ParameterError) Error() string {
	return fmt.Sprintf("%v", p.message)
}

var rootCmd = &cobra.Command{
	Use:   "aping [<url>]",
	Short: "a ping with ability of http/https",
	Long:  `a ping tool for http/https`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args[0]) > 4 && args[0][0:4] == "http" {
			go doHTTPPing(args[0], count)
		} else {
			go doICMPPing(args[0], count)
		}
		for {
			select {
			case err := <-ErrorChan:
				logrus.Error(err)
			case <-DoneChan:
				return nil
			}
		}
	},
}

func Execute() {
	ErrorChan = make(chan error)
	DoneChan = make(chan bool)
	sem = lib.NewSemaphore(coroutine)

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func init() {
	rootCmd.Flags().IntVarP(&count, "count", "c", 4, "stop after c<ount> replies")
	rootCmd.Flags().IntVarP(&interval, "interval", "i", 0, "seconds between sending each packet")
	rootCmd.Flags().IntVar(&coroutine, "coroutine", 10, "coroutine that can be opend. Too many coroutine may cause inaccurate time")
}

func doHTTPPing(url string, count int) {
	for i := 0; i < count; i++ {
		sem.Acquire()

		defer sem.Release()

		time.Sleep(time.Second * time.Duration(interval))

		if err := httpPing(url); err != nil {
			ErrorChan <- err
		}
	}
	DoneChan <- true
}

func doICMPPing(url string, count int) {
	for i := 0; i < count; i++ {
		sem.Acquire()

		defer sem.Release()

		time.Sleep(time.Second * time.Duration(interval))

		if err := ICMPPing(url); err != nil {
			ErrorChan <- err
		}
	}
	DoneChan <- true
}

func RequestFailError(err error) error {
	return fmt.Errorf("RequestFailError: %w", err)
}

func ICMPError(err error) error {
	return fmt.Errorf("ICMPError: %w", err)
}

func httpPing(url string) error {
	var (
		request  *http.Request
		response *http.Response
		err      error
	)

	if request, err = http.NewRequestWithContext(context.Background(), "HEAD", url, nil); err != nil {
		return RequestFailError(err)
	}

	request.Header.Add("User-Agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36 Edg/87.0.664.66`)

	start := time.Now()

	if response, err = http.DefaultClient.Do(request); err != nil {
		return RequestFailError(err)
	}

	elapsed := time.Since(start).Milliseconds()

	runtime.Gosched()

	response.Body.Close()

	shortURL := strings.Replace(url, "https://", "", 1)
	shortURL = strings.Replace(shortURL, "http://", "", 1)

	raddr, err := net.ResolveIPAddr("ip", shortURL)
	if err != nil {
		ErrorChan <- err
	}

	fmt.Printf("ping %v (%v): %v ms\n", url, raddr, elapsed)

	return nil
}

// https://www.cnblogs.com/wlw-x/p/14169607.html
type ICMP struct {
	Type        uint8
	Code        uint8
	CheckSum    uint16
	Identifier  uint16
	SequenceNum uint16
}

func CheckSum(data []byte) (rt uint16) {
	var (
		sum    uint32
		index  int
		length = len(data)
	)

	for length > 1 {
		sum += uint32(data[index])<<8 + uint32(data[index+1])
		index += 2
		length -= 2
	}

	if length > 0 {
		sum += uint32(data[index]) << 8
	}

	rt = uint16(sum) + uint16(sum>>16)

	return ^rt
}

func ICMPPing(url string) error {
	// 构建 ICMP 报文
	var (
		buffer bytes.Buffer
		err    error
		icmp   = ICMP{8, 0, 0, 0, 0}
	)

	if err = binary.Write(&buffer, binary.BigEndian, icmp); err != nil {
		return ICMPError(err)
	}

	icmp.CheckSum = CheckSum(buffer.Bytes())
	buffer.Reset()

	if err = binary.Write(&buffer, binary.BigEndian, icmp); err != nil {
		return ICMPError(err)
	}

	var (
		remoteAddr *net.IPAddr
		conn       *net.IPConn
	)

	if remoteAddr, err = net.ResolveIPAddr("ip", url); err != nil {
		return ICMPError(err)
	}

	start := time.Now()

	// 发送 ICMP 包
	if conn, err = net.DialIP("ip4:icmp", nil, remoteAddr); err != nil {
		return ICMPError(err)
	}

	defer conn.Close()

	if _, err = conn.Write(buffer.Bytes()); err != nil {
		return ICMPError(err)
	}

	// 读取返回的包
	recv := make([]byte, 1024)

	if err = conn.SetReadDeadline(time.Now().Add(time.Second * 3)); err != nil {
		return ICMPError(err)
	}

	if _, err = conn.Read(recv); err != nil {
		return ICMPError(err)
	}

	duration := time.Since(start).Milliseconds()

	conn.Close()
	fmt.Printf("ping %v (%v): %v ms\n", url, remoteAddr, duration)

	return nil
}
