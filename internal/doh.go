package internal

import (
	"context"
	"errors"
	"fmt"

	"time"

	"github.com/likexian/doh-go"
	"github.com/likexian/doh-go/dns"
)

var errProviderNotExist = errors.New("provider not exist")

func ErrProviderNotExist(provider string) error {
	return fmt.Errorf("ErrProviderNotExist: %w : %s", errProviderNotExist, provider)
}

func ErrProvider(err error) error {
	return fmt.Errorf("ProviderNotExist: %w", err)
}

// Cloudflare
// DNSPod
// Google
// Quad9
// auto

func Doh(provider string, url string) (map[string]string, error) {
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
		return nil, ErrProviderNotExist(provider)
	}

	if err != nil {
		return nil, ErrProvider(err)
	}

	answer := response.Answer

	result := make(map[string]string)

	for _, a := range answer {
		result[url] = a.Data
	}

	return result, nil
}
