package ws

import (
	"log/slog"
	"time"

	"game.crisis-control.com/internal/data"
	"game.crisis-control.com/internal/engine"
	"game.crisis-control.com/internal/gamedata"
	"game.crisis-control.com/internal/wsmessages"
)

type playerLookupRequest struct {
	session string
	result  chan *Client
}

type Room struct {
	Players            map[*Client]bool
	Register           chan *Client
	Unregister         chan *Client
	StartGame          chan *Client
	Shutdown           chan struct{}
	GameModel          *data.GameModel
	PlayerModel        *data.PlayerModel
	Code               string
	State              data.GameState
	logger             *slog.Logger
	maxPlayers         int
	currentPlayerCount int
	Broadcast          chan *wsmessages.WSMessage
	engine             *engine.Engine
	PlayerLookup       chan playerLookupRequest
	timer              *time.Timer
	resetTimer         chan struct{}
}

func NewRoom(models *data.Models, code string, maxPlayers int, gm *data.GameModel, pm *data.PlayerModel, logger *slog.Logger, engine *engine.Engine) *Room {
	return &Room{
		Players:            make(map[*Client]bool),
		Register:           make(chan *Client),
		Unregister:         make(chan *Client),
		StartGame:          make(chan *Client),
		Shutdown:           make(chan struct{}),
		GameModel:          gm,
		PlayerModel:        pm,
		State:              data.GameStateLobby,
		Code:               code,
		maxPlayers:         maxPlayers,
		currentPlayerCount: 0,
		logger:             logger,
		Broadcast:          make(chan *wsmessages.WSMessage),
		engine:             engine,
		PlayerLookup:       make(chan playerLookupRequest),
		timer:              time.NewTimer(5 * time.Minute),
		resetTimer:         make(chan struct{}),
	}
}

func (r *Room) setupGameForPlayer(player *Client, treasures []gamedata.TreasureItem) {
	player.setTreasureSet(treasures)
	r.PlayerModel.UpdateTreasureSet(player.session, treasures)
	err := player.Send(wsmessages.WSMessage{Type: wsmessages.GameDataTreasures, Data: map[string]any{
		"treasures": treasures,
		"map":       "first",
	}})
	if err != nil {
		r.logger.Error("error sending treasures to player", "error", err)
	}
}

func (r *Room) GetPlayerClientWithSessionToken(token string) *Client {
	req := playerLookupRequest{
		session: token,
		result:  make(chan *Client, 1),
	}
	r.PlayerLookup <- req
	return <-req.result
}

func (r *Room) startGame(requester *Client, playerSnapshot []*Client, playerCount int) {
	if requester.isCreator {
		if playerCount != r.maxPlayers {
			requester.Send(wsmessages.WSMessage{Type: wsmessages.NotEnoughPlayers})
			return
		}

		r.Broadcast <- &wsmessages.WSMessage{Type: wsmessages.ShowGameLoading}
		treasures := r.engine.GenerateTreasureSet(playerCount)
		for i, player := range playerSnapshot {
			r.setupGameForPlayer(player, treasures[i])
		}

		// Set TTL for the game as well as the players
		err := r.GameModel.SetTTL(r.Code)
		if err != nil {
			r.logger.Error("error setting TTL for game room", "error", err.Error())
			return
		}

		err = r.GameModel.SetTTLForPlayerSet(r.Code, 20*time.Minute)
		if err != nil {
			r.logger.Error("error setting TTL for player set", "gameCode", r.Code, "error", err.Error())
			return
		}

		for _, player := range playerSnapshot {
			err = r.PlayerModel.SetTTL(player.session)
			if err != nil {
				r.logger.Error("error setting TTL for player", "sessionToken", player.session)
				return
			}
		}

		err = r.GameModel.SetGameState(r.Code, data.GameStateInProgress)
		if err != nil {
			r.logger.Error("error updating game state", "error", err.Error())
			return
		}

		r.Broadcast <- &wsmessages.WSMessage{Type: wsmessages.BeginGame}

		r.resetTimer <- struct{}{}
	} else {
		r.logger.Warn("non-creator player sent start game request", "player_session", requester.session, "gameroom", r.Code)
		return
	}
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
		case client := <-r.StartGame:
			playerSnapshot := make([]*Client, 0, len(r.Players))
			for p := range r.Players {
				playerSnapshot = append(playerSnapshot, p)
			}

			playerCount := r.currentPlayerCount

			go func() {
				r.startGame(client, playerSnapshot, playerCount)
			}()
		case req := <-r.PlayerLookup:
			var found *Client
			for player := range r.Players {
				if player.session == req.session {
					found = player
					break
				}
			}
			req.result <- found
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

			r.currentPlayerCount += 1
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

			r.currentPlayerCount -= 1
		case message := <-r.Broadcast:
			for player := range r.Players {
				err := player.Send(*message)
				if err != nil {
					r.logger.Error("error sending broadcast message to player", "error", err, "playerId", player.id)
				}
			}
		case <-r.resetTimer:
			r.State = data.GameStateInProgress
			r.timer.Reset(5 * time.Minute)
		case <-r.timer.C:
			if r.State != data.GameStateInProgress {
				return
			}

			// Send map round over notification here
			for p := range r.Players {
				err := p.Send(wsmessages.WSMessage{Type: wsmessages.MapRoundOver})
				if err != nil {
					r.logger.Error("error sending map round over message to player", "error", err, "playerId", p.id)
				}
			}
		}
	}
}
