package internal

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"
)

func ErrRequestFailed(err error) error {
	return fmt.Errorf("RequestFailError: %w", err)
}

func HTTPPing(url string) error {
	var (
		request  *http.Request
		response *http.Response
		err      error
	)

	if request, err = http.NewRequestWithContext(context.Background(), "HEAD", url, nil); err != nil {
		return ErrRequestFailed(err)
	}

	request.Header.Add("User-Agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36 Edg/87.0.664.66`)

	start := time.Now()

	if response, err = http.DefaultClient.Do(request); err != nil {
		return ErrRequestFailed(err)
	}

	elapsed := time.Since(start).Milliseconds()

	runtime.Gosched()

	response.Body.Close()

	shortURL := strings.Replace(url, "https://", "", 1)
	shortURL = strings.Replace(shortURL, "http://", "", 1)

	raddr, err := net.ResolveIPAddr("ip", shortURL)
	if err != nil {
		return ErrRequestFailed(err)
	}

	fmt.Printf("ping %v (%v): %v ms\n", url, raddr, elapsed)

	return nil
}
