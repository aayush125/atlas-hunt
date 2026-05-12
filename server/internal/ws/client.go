package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"game.crisis-control.com/internal/gamedata"
	"game.crisis-control.com/internal/wsmessages"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type Client struct {
	conn        *websocket.Conn
	send        chan wsmessages.WSMessage
	logger      *slog.Logger
	done        chan struct{}
	room        *Room
	nickname    string
	id          string
	session     string
	isCreator   bool
	treasureSet []gamedata.TreasureItem
	inventory   []gamedata.InventoryItem

	mu            sync.Mutex
	points        int
	treasureCount map[string]int
}

func NewClient(conn *websocket.Conn, room *Room, nickname, id, session string, logger *slog.Logger, isCreator bool) *Client {
	return &Client{
		conn:          conn,
		send:          make(chan wsmessages.WSMessage, 8),
		logger:        logger,
		done:          make(chan struct{}),
		room:          room,
		nickname:      nickname,
		id:            id,
		session:       session,
		isCreator:     isCreator,
		treasureSet:   make([]gamedata.TreasureItem, 0, 40),
		treasureCount: make(map[string]int),
	}
}

func (c *Client) TryClaimTreasure(treasureInstanceID uuid.UUID, questionID, selectedAnswer string) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var newTreasureSet []gamedata.TreasureItem
	found := false
	var t gamedata.TreasureItem
	for _, treasure := range c.treasureSet {
		if treasure.InstanceID == treasureInstanceID {
			found = true
			t = treasure
			continue
		}
		newTreasureSet = append(newTreasureSet, treasure)
	}

	if !found {
		return false, errors.New("treasure not assigned to player")
	}

	correct, err := gamedata.CheckAnswer(questionID, selectedAnswer)
	if err != nil {
		return false, err
	}

	if correct {
		c.treasureCount[t.ID]++
		c.points += t.Points
		c.treasureSet = newTreasureSet
		c.inventory = append(c.inventory, gamedata.InventoryItem{InventoryItemBase: t.InventoryItemBase})

		err := c.Send(wsmessages.WSMessage{Type: wsmessages.TreasureClaimSuccessful, Data: map[string]any{
			"instance_id":    treasureInstanceID,
			"treasure_id":    t.ID,
			"treasure_count": c.treasureCount[t.ID],
			"points":         c.points,
		}})

		if err != nil {
			c.logger.Error("error updating player stats", "playerID", c.id)
		}
	}

	return correct, nil
}

func (c *Client) setTreasureSet(set []gamedata.TreasureItem) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.treasureSet = set
}

func (c *Client) Send(v wsmessages.WSMessage) error {
	select {
	case c.send <- v:
		return nil
	case <-c.done:
		return fmt.Errorf("client disconnected")
	}
}

func (c *Client) handleMessage(message []byte) {
	var msg wsmessages.WSMessage
	err := json.Unmarshal(message, &msg)
	if err != nil {
		c.logger.Error("ws message unmarshal error", "error", err.Error(), "client", c.session)
		return
	}

	switch msg.Type {
	case wsmessages.StartGameRequest:
		c.room.StartGame <- c
	default:
		c.logger.Warn("unrecognized message from client", "client", c.session, "message", msg)
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

		c.handleMessage(message)
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
