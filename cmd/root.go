package cmd

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

var (
	waitGroup sync.WaitGroup

	count    int
	interval int
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
			for i := 0; i < count; i++ {
				waitGroup.Add(1)
				time.Sleep(time.Second * time.Duration(interval))

				go httpPing(args[0])
			}
		} else {
			for i := 0; i < count; i++ {
				waitGroup.Add(1)
				time.Sleep(time.Second * time.Duration(interval))

				go ICMPPing(args[0])
			}
		}
		waitGroup.Wait()

		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func init() {
	rootCmd.Flags().IntVarP(&count, "count", "c", 4, "stop after c<ount> replies")
	rootCmd.Flags().IntVarP(&interval, "interval", "i", 0, "seconds between sending each packet")
}

func httpPing(url string) {
	defer waitGroup.Done()

	request, err := http.NewRequestWithContext(context.Background(), "HEAD", url, nil)
	if err != nil {
		panic(err)
	}

	request.Header.Add("User-Agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36 Edg/87.0.664.66`)

	start := time.Now().UnixNano()
	response, err := http.DefaultClient.Do(request)
	duration := (time.Now().UnixNano() - start) / 1e6

	if err != nil {
		panic(err)
	}

	response.Body.Close()

	fmt.Printf("ping %v: %v ms\n", url, duration)
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

func ICMPPing(url string) {
    defer waitGroup.Done()
	raddr, err := net.ResolveIPAddr("ip", url)
	if err != nil {
		panic(err)
	}

	laddr := net.IPAddr{
		IP: net.ParseIP("0.0.0.0"),
	}

	conn, err := net.DialIP("ip4:icmp", &laddr, raddr)
	if err != nil {
		panic(err)
	}

	defer conn.Close()

	var (
		buffer      bytes.Buffer
		icmp        = ICMP{8, 0, 0, 0, 0}
		originBytes = make([]byte, 2000)
	)

	err = binary.Write(&buffer, binary.BigEndian, icmp)

	if err != nil {
		panic(err)
	}

	err = binary.Write(&buffer, binary.BigEndian, originBytes[0:64])
	if err != nil {
		panic(err)
	}

	b := buffer.Bytes()
	binary.BigEndian.PutUint16(b[2:], CheckSum(b))

	_, err = conn.Write(buffer.Bytes())

	if err != nil {
		panic(err)
	}

	start := time.Now().UnixNano()
	recv := make([]byte, 1024)

	if err = conn.SetReadDeadline(time.Now().Add(time.Second * 3)); err != nil {
		panic(err)
	}

	if _, err = conn.Read(recv); err != nil {
		panic(err)
	}

	duration := (time.Now().UnixNano() - start) / 1e6

	fmt.Printf("ping %v: %v ms\n", url, duration)
}
