package ws

import (
	"maps"
	"time"

	"game.crisis-control.com/internal/data"
	"game.crisis-control.com/internal/gamedata"
	"game.crisis-control.com/internal/wsmessages"
)

func (r *Room) setupGameForPlayer(player *Client, online bool, treasures []gamedata.TreasureItem, gameMap data.GameMap) {
	player.PrepareForNextRound()
	player.setTreasureSet(treasures)
	r.PlayerModel.UpdateTreasureSet(player.session, treasures)
	if online {
		err := player.Send(wsmessages.WSMessage{Type: wsmessages.GameDataTreasures, Data: map[string]any{
			"treasures": treasures,
			"map":       gameMap,
		}})
		if err != nil {
			r.logger.Error("error sending treasures to player", "error", err)
		}
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

func (r *Room) setTTLForGameAndPlayers(playerSnapshot map[*Client]bool) {
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

	for player := range playerSnapshot {
		err = r.PlayerModel.SetTTL(player.session)
		if err != nil {
			r.logger.Error("error setting TTL for player", "sessionToken", player.session)
			continue
		}
	}
}

func (r *Room) startRound(playerSnapshot map[*Client]bool, playerCount int, gameMap data.GameMap) {
	r.Broadcast <- &wsmessages.WSMessage{Type: wsmessages.ShowGameLoading}
	treasures := r.engine.GenerateTreasureSet(playerCount, gameMap)
	count := 0
	for player, online := range playerSnapshot {
		r.setupGameForPlayer(player, online, treasures[count], gameMap)
		count++
	}

	err := r.GameModel.SetGameAndMapState(r.Code, data.GameStateInProgress, gameMap)
	if err != nil {
		r.logger.Error("error updating game state", "error", err.Error())
		return
	}

	now := time.Now()

	r.Broadcast <- &wsmessages.WSMessage{
		Type: wsmessages.BeginGame,
		Data: map[string]any{
			"serverNow": now.UnixMilli(),
			"endsAt":    now.Add(RoundTime).UnixMilli(),
		},
	}

	r.resetTimer <- roundStartData{roundTime: now.Add(RoundTime).UnixMilli(), roundMap: gameMap}
}

func (r *Room) onlinePlayerCount() int {
	count := 0
	for _, online := range r.Players {
		if online {
			count++
		}
	}

	return count
}

func (r *Room) getPlayerSnapshot() map[*Client]bool {
	playerSnapshot := make(map[*Client]bool)
	maps.Copy(playerSnapshot, r.Players)
	return playerSnapshot
}

func (r *Room) newGame(client *Client) {
	r.lastMap = data.GameMapNone
	for player := range r.Players {
		player.Reset()
	}

	if r.onlinePlayerCount() != r.maxPlayers {
		client.Send(wsmessages.WSMessage{Type: wsmessages.NotEnoughPlayers})
		return
	}

	playerSnapshot := r.getPlayerSnapshot()

	playerCount := len(r.Players)

	go func() {
		r.startRound(playerSnapshot, playerCount, data.GameMapFirst)
	}()
}
