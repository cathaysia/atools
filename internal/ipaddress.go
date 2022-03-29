package internal

import (
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
)

func ErrHTTPS(err error) error {
	return fmt.Errorf("ErrHTTPS: %w", err)
}

func IPAddress(url string) (map[string]string, error) {
	request, err := http.NewRequestWithContext(context.Background(), "GET", "https://www.ipaddress.com/search/"+url, nil)
	if err != nil {
		return nil, ErrHTTPS(err)
	}

	request.Header.Add("User-Agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36 Edg/87.0.664.66`)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, ErrHTTPS(err)
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, ErrHTTPS(err)
	}

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil, ErrHTTPS(err)
	}

	result := make(map[string]string)

	doc.Find("tr>td>.comma-separated").First().Find("li").Each(func(i int, s *goquery.Selection) {
		result[url] = s.Text()
	})

	return result, nil
}
