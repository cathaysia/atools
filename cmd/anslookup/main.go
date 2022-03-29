package main

import (
	"atools/internal"
	"flag"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

var waitGroup sync.WaitGroup

func main() {
	useDoh := flag.Bool("d", false, "是否启用 DoH 支持")
	doh := flag.String("doh", "auto", "使用指定的 DoH 提供商\n可能的值为: auto, Cloudflare, DNSPod, Google, Quad9")
	flag.Parse()

	websites := flag.Args()

	waitGroup.Add(len(websites))

	for _, url := range websites {
		if *useDoh {
			go wgWarp(*doh, url)
		} else {
			go wgWarp(url)
		}
	}

	waitGroup.Wait()
}

func Output(res map[string]string) {
	for website, ip := range res {
		str := fmt.Sprintf("Name: %v Address: %v", website, ip)
		fmt.Println(str)
	}
}

func wgWarp(args ...interface{}) {
	defer waitGroup.Done()

	var (
		res map[string]string
		err error
	)

	// 判断调用哪个函数
	arg1, ok := args[0].(string)
	if !ok {
		return
	}

	if len(args) == 1 {
		res, err = internal.IPAddress(arg1)
	} else {
		arg2, ok := args[1].(string)
		if !ok {
			return
		}
		res, err = internal.Doh(arg1, arg2)
	}

	if err != nil {
		logrus.Error(err)

		return
	}

	Output(res)
}
