package ws

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 50 * time.Second
	maxMessageSize = 4096
)

// Client 代表单个 WebSocket 连接
type Client struct {
	userID int64
	conn   *websocket.Conn
	send   chan []byte
	hub    *Hub
	// onMessage 回调，由外部注入消息处理逻辑
	onMessage func(userID int64, raw []byte)
}

func NewClient(userID int64, conn *websocket.Conn, hub *Hub, onMessage func(int64, []byte)) *Client {
	return &Client{
		userID:    userID,
		conn:      conn,
		send:      make(chan []byte, 256),
		hub:       hub,
		onMessage: onMessage,
	}
}

func (c *Client) Run() {
	go c.writePump()
	c.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.Unregister(c.userID)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("ws read error userID=%d: %v", c.userID, err)
			}
			break
		}
		if c.onMessage != nil {
			c.onMessage(c.userID, raw)
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
