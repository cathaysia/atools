package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/likexian/doh-go"
	"github.com/likexian/doh-go/dns"
)

var waitGroup sync.WaitGroup

func Output(website string, ip string) {
	str := fmt.Sprintf("Name: %v Address: %v", website, ip)
	fmt.Println(str)
}

// Cloudflare
// DNSPod
// Google
// Quad9
// auto

func Doh(provider string, url string) {
	defer waitGroup.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var response *dns.Response

	var err error

	switch provider {
	case "auto":
		c := doh.Use(doh.Quad9Provider, doh.DNSPodProvider, doh.CloudflareProvider, doh.GoogleProvider)
		response, err = c.Query(ctx, dns.Domain(url), dns.TypeA)
		c.Close()
	case "Cloudflare":
		c := doh.New(doh.CloudflareProvider)
		response, err = c.Query(ctx, dns.Domain(url), dns.TypeA)
	case "DNSPod":
		c := doh.New(doh.DNSPodProvider)
		response, err = c.Query(ctx, dns.Domain(url), dns.TypeA)
	case "Google":
		c := doh.New(doh.GoogleProvider)
		response, err = c.Query(ctx, dns.Domain(url), dns.TypeA)
	case "Quad9":
		c := doh.New(doh.Quad9Provider)
		response, err = c.Query(ctx, dns.Domain(url), dns.TypeA)
	default:
		panic(fmt.Sprintf("The dns provider %v is invaild\n", provider))
	}

	if err != nil {
		panic(err.Error())
	}

	answer := response.Answer

	for _, a := range answer {
		Output(url, a.Data)
	}
}

func IPAddress(url string) {
	defer waitGroup.Done()

	request, err := http.NewRequestWithContext(context.Background(), "GET", "https://www.ipaddress.com/search/"+url, nil)
	if err != nil {
		panic(err.Error())
	}

	request.Header.Add("User-Agent", `'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36 Edg/87.0.664.66'`)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		panic(err.Error())
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		panic(err.Error())
	}

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		panic(err.Error())
	}

	doc.Find("tr>td>.comma-separated").First().Find("li").Each(func(i int, s *goquery.Selection) {
		Output(url, s.Text())
	})
}

func main() {
	useDoh := flag.Bool("d", false, "是否启用 DoH 支持")
	doh := flag.String("doh", "auto", "使用指定的 DoH 提供商\n可能的值为: auto, Cloudflare, DNSPod, Google, Quad9")
	flag.Parse()

	websites := flag.Args()
	waitGroup.Add(len(websites))

	for _, url := range websites {
		if *useDoh {
			go Doh(*doh, url)
		} else {
			go IPAddress(url)
		}
	}

	waitGroup.Wait()
}
