package ws

import (
	"log/slog"

	"game.crisis-control.com/internal/data"
	"game.crisis-control.com/internal/wsmessages"
)

type Room struct {
	Players     map[*Client]bool
	Register    chan *Client
	Unregister  chan *Client
	Shutdown    chan struct{}
	GameModel   *data.GameModel
	PlayerModel *data.PlayerModel
	Code        string
	logger      *slog.Logger
	Broadcast   chan *wsmessages.WSMessage
}

func NewRoom(models *data.Models, code string, gm *data.GameModel, pm *data.PlayerModel, logger *slog.Logger) *Room {
	return &Room{
		Players:     make(map[*Client]bool),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		Shutdown:    make(chan struct{}),
		GameModel:   gm,
		PlayerModel: pm,
		Code:        code,
		logger:      logger,
		Broadcast:   make(chan *wsmessages.WSMessage),
	}
}

func (r *Room) Run() {
	for {
		select {
		case <-r.Shutdown:
			return
		case player := <-r.Register:
			r.Players[player] = true
			err := r.GameModel.AddPlayerToGame(r.Code)
			if err != nil {
				r.logger.Error("error incrementing player count in redis", "error", err)
			}

			err = player.Send(wsmessages.WSMessage{Type: wsmessages.WelcomeToTheGame})
			if err != nil {
				r.logger.Error("error sending welcome message to player", "error", err)
			}

			for p := range r.Players {
				p.Send(wsmessages.WSMessage{Type: wsmessages.PlayerJoined, Data: map[string]any{
					"nickname": player.nickname,
					"id":       player.id,
				}})
			}
		case player := <-r.Unregister:
			err := r.GameModel.RemovePlayerFromGame(r.Code)
			if err != nil {
				r.logger.Error("error decrementing player count in redis", "error", err)
			}
			delete(r.Players, player)
		case message := <-r.Broadcast:
			for player := range r.Players {
				player.Send(*message)
			}
		}
	}
}
