package main

import (
	"context"
	"flag"
	"github.com/droppoint/occam-fi-demo-task/internal"
	"github.com/droppoint/occam-fi-demo-task/internal/sources/eodhistoricaldata"
	"github.com/droppoint/occam-fi-demo-task/internal/sources/exmocom"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
)

func main() {
	debug := flag.Bool("debug", false, "display debug logs")
	flag.Parse()

	ticker := flag.Arg(0)
	if ticker == "" {
		log.Fatal("You must supply ticker")
	}

	zapConf := zap.NewProductionConfig()
	if debug != nil && *debug == true {
		zapConf = zap.NewDevelopmentConfig()
	}
	logger, err := zapConf.Build()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer logger.Sync() // nolint: errcheck

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	aggregator := internal.NewTickersAggregator(
		internal.Ticker(ticker),
		logger,
		eodhistoricaldata.NewSubscriber("ws.eodhistoricaldata.com", "OeAFFmMliFG5orCUuwAKQ8l4WWFQ67YX", logger),
		exmocom.NewSubscriber("ws-api.exmo.com:443", logger),
	)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-interrupt
		cancel()
	}()
	aggregator.Start(ctx)
}
