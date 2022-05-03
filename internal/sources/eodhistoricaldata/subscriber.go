package eodhistoricaldata

import (
	"context"
	"errors"
	"fmt"
	"github.com/droppoint/occam-fi-demo-task/internal"
	"go.uber.org/zap"
	"time"
)

var tickersMap = map[internal.Ticker]string{
	internal.Ticker("BTCUSD"): "BTC-USD",
}

type Subscriber struct {
	client *Client
	log    *zap.Logger
}

func NewSubscriber(host string, apiKey string, log *zap.Logger) *Subscriber {
	return &Subscriber{client: NewClient(host, apiKey, log), log: log}
}

func (e *Subscriber) SubscribePriceStream(ctx context.Context, ticker internal.Ticker) (chan internal.TickerPrice, chan error) {
	resultCh, errCh := make(chan internal.TickerPrice, 1), make(chan error, 1)

	err := e.client.Connect()
	if err != nil {
		errCh <- fmt.Errorf("e.client.Connect: %w", err)
		_ = e.client.Disconnect()
		close(resultCh)
		close(errCh)
		return resultCh, errCh
	}

	clientTicker, ok := tickersMap[ticker]
	if !ok {
		errCh <- errors.New("ticker is not supported")
		_ = e.client.Disconnect()
		close(resultCh)
		close(errCh)
		return resultCh, errCh
	}

	err = e.client.SubscribeToTicker(clientTicker)
	if err != nil {
		errCh <- fmt.Errorf("e.client.SubscribeToTicker: %w", err)
		_ = e.client.Disconnect()
		close(resultCh)
		close(errCh)
		return resultCh, errCh
	}

	go func() {
		defer close(resultCh)
		defer close(errCh)
		var err error
		for {
			select {
			case <-ctx.Done():
				e.log.Debug("Shutting down subscriber goroutine")
				err = e.client.Disconnect()
				if err != nil {
					errCh <- fmt.Errorf("e.shutdown: %w", err)
				}
				return
			default:
				message, err := e.client.ReadMessage()
				if err != nil {
					errCh <- fmt.Errorf("e.client.ReadMessage: %w", err)
					_ = e.client.Disconnect()
					return
				}
				resultCh <- internal.TickerPrice{
					Price:  message.Price,
					Time:   time.Unix(message.Timestamp/1000, message.Timestamp%1000*1000000),
					Ticker: ticker,
				}
			}
		}

	}()
	return resultCh, errCh
}
