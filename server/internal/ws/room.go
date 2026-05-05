package ws

import (
	"encoding/json"
	"log/slog"
	"sync"

	"game.crisis-control.com/internal/data"
	"game.crisis-control.com/internal/engine"
	"game.crisis-control.com/internal/wsmessages"
)

type Room struct {
	Players            map[*Client]bool
	Register           chan *Client
	Unregister         chan *Client
	Shutdown           chan struct{}
	GameModel          *data.GameModel
	PlayerModel        *data.PlayerModel
	Code               string
	logger             *slog.Logger
	maxPlayers         int
	currentPlayerCount int
	mutex              sync.RWMutex
	Broadcast          chan *wsmessages.WSMessage
	engine             *engine.Engine
}

func NewRoom(models *data.Models, code string, maxPlayers int, gm *data.GameModel, pm *data.PlayerModel, logger *slog.Logger, engine *engine.Engine) *Room {
	return &Room{
		Players:            make(map[*Client]bool),
		Register:           make(chan *Client),
		Unregister:         make(chan *Client),
		Shutdown:           make(chan struct{}),
		GameModel:          gm,
		PlayerModel:        pm,
		Code:               code,
		maxPlayers:         maxPlayers,
		currentPlayerCount: 0,
		logger:             logger,
		Broadcast:          make(chan *wsmessages.WSMessage),
		engine:             engine,
	}
}

func (r *Room) HandleMessage(message []byte, sender *Client) {
	var msg wsmessages.WSMessage
	err := json.Unmarshal(message, &msg)
	if err != nil {
		r.logger.Error("ws message unmarshal error", "error", err.Error())
		return
	}

	switch msg.Type {
	case wsmessages.StartGameRequest:
		if sender.isCreator {
			if r.IsReadyToPlay() {
				r.Broadcast <- &wsmessages.WSMessage{Type: wsmessages.ShowGameLoading}
			} else {
				sender.Send(wsmessages.WSMessage{Type: wsmessages.NotEnoughPlayers})
				return
			}

			treasures := r.engine.GenerateTreasureSet(r.currentPlayerCount)
			i := 0
			for player := range r.Players {
				player.Send(wsmessages.WSMessage{Type: wsmessages.GameDataTreasures, Data: map[string]any{
					"treasures": treasures[i],
				}})
				i++
			}
		} else {
			r.logger.Warn("non-creator player sent start game request", "nickname", sender.nickname, "gameroom", r.Code)
			return
		}
	default:
		r.logger.Warn("unknown ws request", "type", msg.Type)
	}
}

func (r *Room) IsReadyToPlay() bool {
	r.mutex.RLock()
	notReady := r.maxPlayers != r.currentPlayerCount
	r.mutex.RUnlock()

	return !notReady
}

func (r *Room) Run() {
	for {
		select {
		case <-r.Shutdown:
			for player := range r.Players {
				player.conn.Close()
			}

			for len(r.Players) > 0 {
				player := <-r.Unregister
				delete(r.Players, player)
			}
			return
		case player := <-r.Register:
			r.Players[player] = true
			err := r.GameModel.IncrementPlayerCount(r.Code)
			if err != nil {
				r.logger.Error("error incrementing player count in redis", "error", err)
			}

			err = player.Send(wsmessages.WSMessage{Type: wsmessages.WelcomeToTheGame})
			if err != nil {
				r.logger.Error("error sending welcome message to player", "error", err)
				break
			}

			for p := range r.Players {
				err = p.Send(wsmessages.WSMessage{Type: wsmessages.PlayerJoined, Data: map[string]any{
					"nickname": player.nickname,
					"id":       player.id,
				}})
				if err != nil {
					r.logger.Error("error sending player joined message to player", "error", err, "playerId", p.id)
				}
			}

			r.mutex.Lock()
			r.currentPlayerCount += 1
			r.mutex.Unlock()
		case player := <-r.Unregister:
			err := r.GameModel.DecrementPlayerCount(r.Code)
			if err != nil {
				r.logger.Error("error decrementing player count in redis", "error", err)
			}

			err = r.GameModel.RemovePlayerFromSet(r.Code, player.session)
			if err != nil {
				r.logger.Error("error removing player from set", "error", err)
			}

			delete(r.Players, player)

			for p := range r.Players {
				err = p.Send(wsmessages.WSMessage{Type: wsmessages.PlayerLeft, Data: map[string]any{
					"id": player.id,
				}})
				if err != nil {
					r.logger.Error("error sending player left message to player", "error", err, "playerId", p.id)
				}
			}

			r.mutex.Lock()
			r.currentPlayerCount -= 1
			r.mutex.Unlock()
		case message := <-r.Broadcast:
			for player := range r.Players {
				err := player.Send(*message)
				if err != nil {
					r.logger.Error("error sending broadcast message to player", "error", err, "playerId", player.id)
				}
			}
		}
	}
}
