package realtime

import (
	"log"
	"sync"
	"time"
	"backend/models"
	"github.com/gorilla/websocket"
)

const (
	outboundQueueSize = 32
	maxInboundBytes = 16 * 1024 
	
	writeWait = 10 * time.Second
	pongWait = 60 * time.Second
	pingPeriod = 50 * time.Second
)

type InboundHandler func(client *Client, payload []byte)

type Client struct {
	ChatID string
	Outbound chan models.Message

	hub *Hub
	conn *websocket.Conn
	onMessage InboundHandler

	done chan struct{}
	closeOnce sync.Once
}

func NewClient(
	hub *Hub,
	chatID string,
	conn *websocket.Conn,
	onMessage InboundHandler,
)  *Client {
	return &Client{
		ChatID: chatID,
		Outbound: make(chan models.Message, outboundQueueSize),
		hub: hub,
		conn: conn,
		onMessage: onMessage,
		done: make(chan struct{}),
	}
}

func (c *Client) Run() {
	c.hub.Register(c)
	go c.writePump()

	defer c.hub.Unregister(c)
	c.readPump()
}

func (c *Client) readPump() {
	if c.conn == nil {
		return
	}

	c.conn.SetReadLimit(maxInboundBytes)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		messageType, payload, err := c.conn.ReadMessage()
		if err != nil {
			return
		}

		if messageType != websocket.TextMessage {
			continue
		}

		if c.onMessage != nil {
			c.onMessage(c, payload)
		}
	}
}

func (c *Client) writePump() {
	if c.conn == nil {
		return
	}

	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-c.done: return

		case message, ok := <-c.Outbound:
			if !ok {
				return
			}
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				return
			}
			if err := c.conn.WriteJSON(message); err != nil {
				return
			}

		case <-ticker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				return
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) shutdown() {
	c.closeOnce.Do(func() {
		close(c.done)
		close(c.Outbound)

		if c.conn != nil {
			if err := c.conn.Close(); err != nil {
				log.Printf("close chat WEBSocket: %v", err)
			}
		}
	})
}