package ws

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"game.crisis-control.com/internal/wsmessages"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type Client struct {
	conn      *websocket.Conn
	send      chan wsmessages.WSMessage
	logger    *slog.Logger
	done      chan struct{}
	room      *Room
	nickname  string
	id        string
	session   string
	isCreator bool
}

func NewClient(conn *websocket.Conn, room *Room, nickname, id, session string, logger *slog.Logger, isCreator bool) *Client {
	return &Client{
		conn:      conn,
		send:      make(chan wsmessages.WSMessage, 1),
		logger:    logger,
		done:      make(chan struct{}),
		room:      room,
		nickname:  nickname,
		id:        id,
		session:   session,
		isCreator: isCreator,
	}
}

func (c *Client) Send(v wsmessages.WSMessage) error {
	select {
	case c.send <- v:
		return nil
	case <-c.done:
		return fmt.Errorf("client disconnected")
	}
}

func (c *Client) HandleMessage(message []byte) {
	var msg wsmessages.WSMessage
	err := json.Unmarshal(message, &msg)
	if err != nil {
		c.logger.Error("ws message unmarshal error", "error", err.Error())
		return
	}

	switch msg.Type {
	case wsmessages.StartGameRequest:
		// verify this client is the creator
		if c.isCreator {
			if c.room.IsReadyToPlay() {
				err = c.Send(wsmessages.WSMessage{Type: wsmessages.AcknowledgeStartRequest})
				if err != nil {
					c.logger.Error("error sending start game acknowledgement", "error", err.Error())
					return
				}
			}
		} else {
			c.logger.Warn("non-creator player sent start game request", "nickname", c.nickname, "game room", c.room.Code)
			return
		}
	default:
		c.logger.Warn("unknown ws request", "type", msg.Type)
	}
}

func (c *Client) ReadPump() {
	defer func() {
		close(c.done)
		close(c.send)
		c.conn.Close()
		c.room.Unregister <- c
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure) {
				c.logger.Error("unexpected websocket close", "error", err)
			}
			break
		}

		c.HandleMessage(message)
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	// notify the client that the websocket connection has been established
	c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	data, err := json.Marshal(wsmessages.WSMessage{Type: wsmessages.ConnectionEstablished})
	if err != nil {
		c.logger.Error("failed to marshal websocket connection established message")
		return
	}

	if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return
	}

	for {
		select {
		case <-c.done:
			return
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}

			data, err := json.Marshal(message)
			if err != nil {
				c.logger.Error("failed to marshal websocket message", "type", message.Type, "error", err)
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
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
