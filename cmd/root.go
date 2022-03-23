package cmd

import (
	"time"

	"aping/internal"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	count     int
	interval  int
	coroutine int

	sem       *internal.Semaphore
	ErrorChan chan error
	DoneChan  chan bool
)

var rootCmd = &cobra.Command{
	Use:   "aping [<url>]",
	Short: "a ping with ability of http/https",
	Long:  `a ping tool for http/https`,
	Args:  cobra.MaximumNArgs(1),
	Run:   doPing,
}

func Execute() {
	ErrorChan = make(chan error)
	DoneChan = make(chan bool)
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
	if len(args[0]) > 4 && args[0][0:4] == "http" {
		go func() {
			for i := 0; i < count; i++ {
				sem.Acquire()

				defer sem.Release()

				time.Sleep(time.Second * time.Duration(interval))

				if err := internal.HTTPPing(args[0]); err != nil {
					ErrorChan <- err
				}
			}
			DoneChan <- true
		}()
	} else {
		go func() {
			for i := 0; i < count; i++ {
				sem.Acquire()

				defer sem.Release()

				time.Sleep(time.Second * time.Duration(interval))

				if err := internal.ICMPPing(args[0]); err != nil {
					ErrorChan <- err
				}
			}
			DoneChan <- true
		}()
	}

	for {
		select {
		case err := <-ErrorChan:
			logrus.Error(err)
		case <-DoneChan:
			return
		}
	}
}
