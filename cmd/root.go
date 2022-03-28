package cmd

import (
	"sync"
	"time"

	"aping/internal"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	count     int
	interval  int
	coroutine int

	sem     *internal.Semaphore
	waitDog sync.WaitGroup
)

var rootCmd = &cobra.Command{
	Use:   "aping [<url>]",
	Short: "a ping with ability of http/https",
	Long:  `a ping tool for http/https`,
	Args:  cobra.MaximumNArgs(1),
	Run:   doPing,
}

func Execute() {
	sem = internal.NewSemaphore(coroutine)

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func init() {
	rootCmd.Flags().IntVarP(&count, "count", "c", 4, "stop after c<ount> replies")
	rootCmd.Flags().IntVarP(&interval, "interval", "i", 0, "seconds between sending each packet")
	rootCmd.Flags().IntVar(&coroutine, "coroutine", 10, "coroutine that can be opend. Too many coroutine may cause inaccurate time")
}

func doPing(cmd *cobra.Command, args []string) {
	wgWrap := func(f func(string) error, url string) {
		sem.Acquire()
		defer sem.Release()
		defer waitDog.Done()

		if err := f(url); err != nil {
			logrus.Error(err)

			return
		}
	}

    waitDog.Add(count)
	if len(args[0]) > 4 && args[0][0:4] == "http" {
		for i := 0; i < count; i++ {
			time.Sleep(time.Second * time.Duration(interval))

			go wgWrap(internal.HTTPPing, args[0])
		}
	} else {
		for i := 0; i < count; i++ {
			time.Sleep(time.Second * time.Duration(interval))

			go wgWrap(internal.ICMPPing, args[0])
		}
	}

	waitDog.Wait()
}
