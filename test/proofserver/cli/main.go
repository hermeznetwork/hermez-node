package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/hermeznetwork/hermez-node/test/proofserver"
)

func main() {
	var addr string
	flag.StringVar(&addr, "a", "localhost:3000", "listen address")
	var provingDuration time.Duration
	flag.DurationVar(&provingDuration, "d", 2*time.Second, "proving time duration") //nolint:gomnd
	flag.Parse()

	mock := proofserver.NewMock(addr, provingDuration)
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		if err := mock.Run(ctx); err != nil {
			log.Fatal(err)
		}
		wg.Done()
	}()

	stopCh := make(chan interface{})
	// catch ^C to send the stop signal
	ossig := make(chan os.Signal, 1)
	signal.Notify(ossig, os.Interrupt)
	go func() {
		for sig := range ossig {
			if sig == os.Interrupt {
				stopCh <- nil
			}
		}
	}()
	<-stopCh
	cancel()
	wg.Wait()
}
