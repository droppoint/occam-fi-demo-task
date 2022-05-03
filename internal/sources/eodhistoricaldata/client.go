package eodhistoricaldata

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"net/url"
)

type TickerMessage struct {
	Ticker          string `json:"s"`
	Price           string `json:"p"`
	Quantity        string `json:"q"`
	DailyChange     string `json:"dc"`
	DailyDifference string `json:"dd"`
	Timestamp       int64  `json:"t"`
}

type Client struct {
	host   string
	apiKey string
	conn   *websocket.Conn
	log    *zap.Logger
}

type AuthMessage struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

func NewClient(host string, apiKey string, log *zap.Logger) *Client {
	return &Client{host: host, apiKey: apiKey, log: log}
}

func (c *Client) Connect() error {
	var err error

	if c.conn != nil {
		return nil
	}

	c.log.Debug("Establishing connection", zap.String("host", c.host))
	params := url.Values{}
	params.Add("api_token", c.apiKey)
	u := url.URL{Scheme: "wss", Host: c.host, Path: "ws/crypto", RawQuery: params.Encode()}
	c.conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("websocket.DefaultDialer.Dial: %w", err)
	}

	var message AuthMessage
	err = c.conn.ReadJSON(&message)
	if err != nil {
		return fmt.Errorf("reading auth message conn.ReadJSON: %w", err)
	}

	if message.StatusCode != 200 {
		return errors.New("authentication error")
	}
	return nil
}

func (c *Client) SubscribeToTicker(ticker string) error {
	c.log.Debug("Subscribing to ticker", zap.String("ticker", ticker))
	err := c.conn.WriteJSON(map[string]string{"action": "subscribe", "symbols": ticker})
	if err != nil {
		return fmt.Errorf("c.conn.WriteJSON: %w", err)
	}
	return nil
}

func (c *Client) ReadMessage() (*TickerMessage, error) {
	var message TickerMessage
	err := c.conn.ReadJSON(&message)
	if err != nil {
		return nil, fmt.Errorf("c.conn.ReadJSON: %w", err)
	}
	return &message, nil
}

func (c *Client) Disconnect() error {
	if c.conn == nil {
		return nil
	}

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
