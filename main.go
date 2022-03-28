package main

import (
	"anslookup/internal"
	"flag"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

var waitGroup sync.WaitGroup

func Output(res map[string]string) {
	for website, ip := range res {
		str := fmt.Sprintf("Name: %v Address: %v", website, ip)
		fmt.Println(str)
	}
}

func main() {
	useDoh := flag.Bool("d", false, "是否启用 DoH 支持")
	doh := flag.String("doh", "auto", "使用指定的 DoH 提供商\n可能的值为: auto, Cloudflare, DNSPod, Google, Quad9")
	flag.Parse()

	websites := flag.Args()

	dogDoh := func(provider, url string) {
		defer waitGroup.Done()

		res, err := internal.Doh(provider, url)
		if err != nil {
			logrus.Error(err)

			return
		}

		Output(res)
	}
	dogIPAddress := func(url string) {
		defer waitGroup.Done()

		res, err := internal.IPAddress(url)
		if err != nil {
			logrus.Error(err)

			return
		}

		Output(res)
	}

	waitGroup.Add(len(websites))

	for _, url := range websites {
		if *useDoh {
			go dogDoh(*doh, url)
		} else {
			go dogIPAddress(url)
		}
	}

	waitGroup.Wait()
}
