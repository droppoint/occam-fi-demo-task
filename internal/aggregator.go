package internal

import (
	"context"
	"fmt"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"sync"
	"time"
)

type PriceStreamSubscriber interface {
	SubscribePriceStream(ctx context.Context, ticker Ticker) (chan TickerPrice, chan error)
}
type TickersAggregator struct {
	ticker      Ticker
	subscribers []PriceStreamSubscriber
	lastPrices  map[int64]decimal.Decimal
	log         *zap.Logger
	mutex       sync.RWMutex
}

func NewTickersAggregator(ticker Ticker, log *zap.Logger, subscribers ...PriceStreamSubscriber) *TickersAggregator {
	return &TickersAggregator{
		ticker:      ticker,
		subscribers: subscribers,
		lastPrices:  make(map[int64]decimal.Decimal),
		log:         log,
	}
}

func (t *TickersAggregator) Start(ctx context.Context) {
	t.log.Debug("Starting TickersAggregator")

	var wg sync.WaitGroup
	for i, subscriber := range t.subscribers {
		wg.Add(1)
		i, subscriber := i, subscriber
		go func(ctx context.Context, i int, subscriber PriceStreamSubscriber) {
			t.log.Debug("Starting subscriber", zap.Int("idx", i))
			resCh, errCh := subscriber.SubscribePriceStream(ctx, t.ticker)
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					t.log.Debug("Shutting down subscriber", zap.Int("idx", i))
					return
				case message := <-resCh:
					price, err := decimal.NewFromString(message.Price)
					if err != nil {
						t.log.Warn("Invalid value received", zap.Int("idx", i), zap.Error(err))
						t.mutex.Lock()
						delete(t.lastPrices, int64(i))
						t.mutex.Unlock()
						continue
					}

					t.mutex.Lock()
					t.lastPrices[int64(i)] = price
					t.mutex.Unlock()
				case err := <-errCh:
					t.log.Error("Subscriber failed", zap.Int("idx", i), zap.Error(err))
					t.mutex.Lock()
					delete(t.lastPrices, int64(i))
					t.mutex.Unlock()
					return
				}
			}
		}(ctx, i, subscriber)
	}
	go t.serveOutput(ctx)
	wg.Wait()
	// Wait for 5 seconds to complete all established connections
	time.Sleep(5 * time.Second)
	t.log.Debug("TickersAggregator shutdown completed")
}

func (t *TickersAggregator) serveOutput(ctx context.Context) {
	t.log.Debug("Starting serveOutput")

	// Initial sleep to align with the start of the minute
	time.Sleep(time.Duration(int64(time.Minute) - (time.Now().UnixNano() % int64(time.Minute))))
	ticker := time.NewTicker(time.Minute)
	for {
		select {
		case <-ctx.Done():
			t.log.Debug("Shutting down serveOutput")
			return
		case currentTime := <-ticker.C:
			t.mutex.RLock()
			if len(t.lastPrices) == 0 {
				continue
			}
			prices := make([]decimal.Decimal, len(t.lastPrices))
			i := 0
			for _, price := range t.lastPrices {
				prices[i] = price
				i++
			}
			t.mutex.RUnlock()
			avgPrice := decimal.Avg(prices[0], prices[1:]...)
			fmt.Println(currentTime.Unix(), avgPrice.String())
		}
	}
}
