/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package main

import (
	"sync"
	"time"

	"atools/internal"
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
	Args:  cobra.ExactArgs(1),
	Run:   doPing,
}

func main() {
	rootCmd.Flags().IntVarP(&count, "count", "c", 4, "stop after c<ount> replies")
	rootCmd.Flags().IntVarP(&interval, "interval", "i", 0, "seconds between sending each packet")
	rootCmd.Flags().IntVar(&coroutine, "coroutine", 10, "coroutine that can be opend. Too many coroutine may cause inaccurate time")

	sem = internal.NewSemaphore(coroutine)

	if err := rootCmd.Execute(); err != nil {
		logrus.Error(err)
	}
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
