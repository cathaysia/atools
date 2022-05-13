package internal

import (
	"context"
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

type IPAddressError struct {
	msg string
}

func NewIPAddressErr(err error, msg string) error {
	if err != nil {
		msg = fmt.Sprintf("%v", err.Error()) + msg
	}

	return &IPAddressError{
		msg: fmt.Sprintf("IPAddressError: %v", msg),
	}
}

func (ipaddressErr *IPAddressError) Error() string {
	return ipaddressErr.msg
}

func IPAddress(url string) (map[string]string, error) {
	request, err := http.NewRequestWithContext(context.Background(), "GET", "https://www.ipaddress.com/search/"+url, nil)
	if err != nil {
		return nil, NewIPAddressErr(err, "")
	}

	request.Header.Add("User-Agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36 Edg/87.0.664.66`)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, NewIPAddressErr(err, "")
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, NewIPAddressErr(nil, fmt.Sprintf("HTTP 状态码为 %v 而不是 200", response.StatusCode))
	}

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil, NewIPAddressErr(err, "")
	}

	result := make(map[string]string)

	doc.Find("tr>td>.comma-separated").First().Find("li").Each(func(i int, s *goquery.Selection) {
		result[url] = s.Text()
	})

	return result, nil
}
