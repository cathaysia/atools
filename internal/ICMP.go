package internal

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

func ErrICMPFailed(err error) error {
	return fmt.Errorf("ICMPError: %w", err)
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
		return ErrICMPFailed(err)
	}

	icmp.CheckSum = CheckSum(buffer.Bytes())
	buffer.Reset()

	if err = binary.Write(&buffer, binary.BigEndian, icmp); err != nil {
		return ErrICMPFailed(err)
	}

	var (
		remoteAddr *net.IPAddr
		conn       *net.IPConn
	)

	if remoteAddr, err = net.ResolveIPAddr("ip", url); err != nil {
		return ErrICMPFailed(err)
	}

	start := time.Now()

	// 发送 ICMP 包
	if conn, err = net.DialIP("ip4:icmp", nil, remoteAddr); err != nil {
		return ErrICMPFailed(err)
	}

	defer conn.Close()

	if _, err = conn.Write(buffer.Bytes()); err != nil {
		return ErrICMPFailed(err)
	}

	// 读取返回的包
	recv := make([]byte, 1024)

	if err = conn.SetReadDeadline(time.Now().Add(time.Second * 3)); err != nil {
		return ErrICMPFailed(err)
	}

	if _, err = conn.Read(recv); err != nil {
		return ErrICMPFailed(err)
	}

	duration := time.Since(start).Milliseconds()

	conn.Close()
	fmt.Printf("ping %v (%v): %v ms\n", url, remoteAddr, duration)

	return nil
}
