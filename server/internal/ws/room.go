package ws

import (
	"log/slog"
	"sync"

	"game.crisis-control.com/internal/data"
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
}

func NewRoom(models *data.Models, code string, maxPlayers int, gm *data.GameModel, pm *data.PlayerModel, logger *slog.Logger) *Room {
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
