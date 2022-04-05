/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package main

import (
	"context"
	"sync"
	"time"

	"atools/internal"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/sync/semaphore"
)

var (
	count     int
	interval  int
	coroutine int

	sem     *semaphore.Weighted
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

	sem = semaphore.NewWeighted(int64(coroutine))

	if err := rootCmd.Execute(); err != nil {
		logrus.Error(err)
	}
}

func doPing(cmd *cobra.Command, args []string) {
	wgWrap := func(Func func(string) error, url string) {
		if err := sem.Acquire(context.TODO(), 1); err != nil {
			logrus.Fatal(err)
		}

		defer sem.Release(1)
		defer waitDog.Done()

		if err := Func(url); err != nil {
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
