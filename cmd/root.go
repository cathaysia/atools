package cmd

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

var waitGroup sync.WaitGroup

var count int

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
		if len(args[0]) < 4 {
			return &ParameterError{
				message: fmt.Sprintf("url 格式不正确：%v", args[0]),
			}
		}
		if args[0][0:4] != "http" {
			return &ParameterError{
				message: fmt.Sprintf("url 不以 http 开头：%v", args[0]),
			}
		}
		for i := 0; i < count; i++ {
			waitGroup.Add(1)
			go runPing(args[0])
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
}

func runPing(url string) {
	defer waitGroup.Done()

	request, err := http.NewRequestWithContext(context.Background(), "HEAD", url, nil)
	if err != nil {
		panic(err)
	}

	request.Header.Add("User-Agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36 Edg/87.0.664.66`)

	startT := time.Now().Nanosecond()
	response, err := http.DefaultClient.Do(request)
	startE := time.Now().Nanosecond() - startT

	if startE < 0 {
		startE = -startE
	}

	if err != nil {
		panic(err)
	}

	response.Body.Close()

	fmt.Printf("ping %v: %v ms\n", url, startE/1000000)
}
