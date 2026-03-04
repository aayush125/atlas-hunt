package ws

import (
	"log/slog"

	"game.crisis-control.com/internal/data"
	"game.crisis-control.com/internal/wsmessages"
)

type Room struct {
	Players    map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
	Shutdown   chan struct{}
	Models     *data.Models
	Code       string
	logger     *slog.Logger
}

func NewRoom(models *data.Models, code string, logger *slog.Logger) *Room {
	return &Room{
		Players:    make(map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Shutdown:   make(chan struct{}),
		Models:     models,
		Code:       code,
		logger:     logger,
	}
}

func (r *Room) Run() {
	for {
		select {
		case <-r.Shutdown:
			return
		case player := <-r.Register:
			r.Players[player] = true
			err := r.Models.Game.AddPlayerToGame(r.Code)
			if err != nil {
				r.logger.Error("error incrementing player count in redis", "error", err)
			}

			err = player.Send(wsmessages.WSMessage{Type: wsmessages.WelcomeToTheGame})
			if err != nil {
				r.logger.Error("error sending welcome message to player", "error", err)
			}
		case player := <-r.Unregister:
			delete(r.Players, player)
		}
	}
}
