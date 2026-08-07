package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"sync"
	"time"

	"atlas.sageyutta.com/internal/gamedata"
	"atlas.sageyutta.com/internal/wsmessages"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
	inventorySize  = 25
)

var (
	ErrPlayerDoesNotHaveRequestedTreasure = errors.New("player does not have requested treasure")
	ErrInvalidTreasureCount               = errors.New("invalid treasure count")
	ErrInventoryFull                      = errors.New("inventory is full")
	ErrTreasureNotFound                   = errors.New("treasure not found")
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

	successfullyClaimed int
	failedToClaimCount  int

	mu            sync.Mutex
	PumpsWg       sync.WaitGroup
	points        int
	treasureCount map[string]int
}

type ClientSnapshot struct {
	TreasureSet         []gamedata.TreasureItem
	Inventory           []gamedata.InventoryItem
	SuccessfullyClaimed int
	FailedToClaimCount  int
	TreasureCount       map[string]int
	Points              int
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

func (c *Client) Snapshot() ClientSnapshot {
	c.mu.Lock()
	defer c.mu.Unlock()
	return ClientSnapshot{
		TreasureSet:         append([]gamedata.TreasureItem{}, c.treasureSet...),
		Inventory:           append([]gamedata.InventoryItem{}, c.inventory...),
		SuccessfullyClaimed: c.successfullyClaimed,
		FailedToClaimCount:  c.failedToClaimCount,
		TreasureCount:       maps.Clone(c.treasureCount),
		Points:              c.points,
	}
}

func (c *Client) DeleteTreasure(instanceID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, item := range c.inventory {
		if item.InstanceID.String() == instanceID {
			c.inventory = append(c.inventory[:i], c.inventory[i+1:]...)
			break
		}
	}
}

func (c *Client) TryClaimBonusPoints(treasureID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	count, ok := c.treasureCount[treasureID]
	if !ok {
		return ErrPlayerDoesNotHaveRequestedTreasure
	}

	if count <= 1 {
		return ErrInvalidTreasureCount
	}

	bonus, err := gamedata.GetBonusPoints(treasureID, count)
	if err != nil {
		return err
	}

	delete(c.treasureCount, treasureID)
	newInventory := []gamedata.InventoryItem{}
	for _, item := range c.inventory {
		if item.ID != treasureID {
			newInventory = append(newInventory, item)
		}
	}
	c.inventory = newInventory
	c.points += bonus

	c.room.NotifyBonusClaimed <- playerNotification{
		Client: c,
		Message: wsmessages.WSMessage{
			Type: wsmessages.BonusClaimSuccessful,
			Data: map[string]any{
				"points":      c.points,
				"treasure_id": treasureID,
			},
		},
	}

	return nil
}

func (c *Client) TryClaimTreasure(treasureInstanceID uuid.UUID, questionID, selectedAnswer string) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.inventory) >= inventorySize {
		return false, ErrInventoryFull
	}

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
		return false, ErrTreasureNotFound
	}

	correct, err := gamedata.CheckAnswer(questionID, selectedAnswer)
	if err != nil {
		return false, err
	}

	if correct {
		c.treasureCount[t.ID]++
		c.points += t.Points
		c.successfullyClaimed++
		c.treasureSet = newTreasureSet
		c.inventory = append(c.inventory, gamedata.InventoryItem{InventoryItemBase: t.InventoryItemBase})

		c.room.NotifyTreasureClaimed <- playerNotification{
			Client: c,
			Message: wsmessages.WSMessage{
				Type: wsmessages.TreasureClaimSuccessful,
				Data: map[string]any{
					"instance_id":    treasureInstanceID,
					"treasure_id":    t.ID,
					"treasure_count": c.treasureCount[t.ID],
					"points":         c.points,
				},
			},
		}
	} else {
		c.failedToClaimCount++
	}

	return correct, nil
}

func (c *Client) setTreasureSet(set []gamedata.TreasureItem) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.treasureSet = set
}

func (c *Client) SetInventory(inventory []gamedata.InventoryItem) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.inventory = inventory
}

func (c *Client) PrepareForNextRound() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.successfullyClaimed = 0
	c.failedToClaimCount = 0
	c.treasureSet = []gamedata.TreasureItem{}
}

func (c *Client) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.successfullyClaimed = 0
	c.failedToClaimCount = 0
	c.treasureSet = []gamedata.TreasureItem{}
	c.inventory = []gamedata.InventoryItem{}
	c.treasureCount = map[string]int{}
	c.points = 0
}

func (c *Client) Send(v wsmessages.WSMessage) error {
	select {
	case c.send <- v:
		return nil
	case <-c.done:
		return fmt.Errorf("client disconnected")
	}
}

func (c *Client) Instantiate(conn *websocket.Conn) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.conn = conn
	c.send = make(chan wsmessages.WSMessage, 8)
	c.done = make(chan struct{})
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
	case wsmessages.NextRound:
		c.room.NextRound <- c
	case wsmessages.NewGame:
		c.room.NewGame <- c
	default:
		c.logger.Warn("unrecognized message from client", "clientSession", c.session, "message", msg)
	}
}

func (c *Client) ReadPump() {
	defer func() {
		close(c.done)
		c.conn.Close()
		c.room.Unregister <- c
		c.PumpsWg.Done()
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
		c.PumpsWg.Done()
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
