package exmocom

import (
	"fmt"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"net/url"
)

type TickerMessage struct {
	BuyPrice  string `json:"buy_price"`
	SellPrice string `json:"sell_price"`
	LastTrade string `json:"last_trade"`
	High      string `json:"high"`
	Low       string `json:"low"`
	Avg       string `json:"avg"`
	Vol       string `json:"vol"`
	VolCurr   string `json:"vol_curr"`
	Updated   int    `json:"updated"`
}

type Message struct {
	ID        *int64         `json:"id,omitempty"`
	Timestamp int64          `json:"ts"`
	Event     string         `json:"event"`
	Topic     string         `json:"topic"`
	Data      *TickerMessage `json:"data,omitempty"`
}

type Client struct {
	host string
	conn *websocket.Conn
	log  *zap.Logger
}

func NewClient(host string, log *zap.Logger) *Client {
	return &Client{host: host, log: log}
}

func (c *Client) Connect() error {
	var err error

	if c.conn != nil {
		return nil
	}

	c.log.Debug("Establishing connection", zap.String("host", c.host))

	u := url.URL{Scheme: "wss", Host: c.host, Path: "v1/public"}
	c.conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("websocket.DefaultDialer.Dial: %w", err)
	}

	return nil
}

func (c *Client) SubscribeToTicker(ticker string) error {
	c.log.Debug("Subscribing to ticker", zap.String("ticker", ticker))

	err := c.conn.WriteJSON(map[string]interface{}{"method": "subscribe", "topics": []string{ticker}})
	if err != nil {
		return fmt.Errorf("c.conn.WriteJSON: %w", err)
	}
	return nil
}

func (c *Client) ReadMessage() (*Message, error) {
	var message Message
	err := c.conn.ReadJSON(&message)
	if err != nil {
		return nil, fmt.Errorf("c.conn.ReadJSON: %w", err)
	}
	return &message, nil
}

func (c *Client) Disconnect() error {
	c.log.Debug("Disconnecting", zap.String("host", c.host))

	err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		return fmt.Errorf("c.conn.WriteMessage: %w", err)
	}
	err = c.conn.Close()
	if err != nil {
		return fmt.Errorf("c.conn.Close: %w", err)
	}
	return nil
}
