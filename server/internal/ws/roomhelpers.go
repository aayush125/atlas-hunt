package ws

import (
	"fmt"
	"maps"
	"time"

	"atlas.sageyutta.com/internal/data"
	"atlas.sageyutta.com/internal/gamedata"
	"atlas.sageyutta.com/internal/wsmessages"
)

func (r *Room) registerPlayer(player *Client) {
	r.Players[player] = true
	err := r.GameModel.IncrementPlayerCount(r.Code)
	if err != nil {
		r.logger.Error("error incrementing player count in redis", "error", err)
	}

	err = player.Send(wsmessages.WSMessageOut{Type: wsmessages.WelcomeToTheGame})
	if err != nil {
		r.logger.Error("error sending welcome message to player", "error", err)
		return
	}

	playerJoinedMessage := wsmessages.WSMessageOut{
		Type: wsmessages.PlayerJoined,
		Data: map[string]any{
			"nickname": player.nickname,
			"id":       player.id,
		},
	}

	r.broadcastMessage(
		&playerJoinedMessage,
		func(online bool, p *Client) bool {
			return !online || p.session == player.session
		},
		"error sending player joined message to player",
	)
}

func (r *Room) unregisterPlayer(player *Client) bool {
	err := r.GameModel.DecrementPlayerCount(r.Code)
	if err != nil {
		r.logger.Error("error decrementing player count in redis", "error", err)
	}

	if r.State == data.GameStateLobby { // if game is still in lobby, delete the player
		playerCount := len(r.Players)
		delete(r.Players, player)
		if playerCount > 1 { // to prevent "record not found" error when the room's defer function tries to clean up the player set
			err := r.GameModel.RemovePlayerFromSet(r.Code, player.session)
			if err != nil {
				r.logger.Error("error removing player from set in redis", "playerSession", player.session, "error", err)
			}
		}

		err = r.PlayerModel.Delete(player.session)
		if err != nil {
			r.logger.Error("error deleting player from redis", "playerSession", player.session, "error", err)
		}
	} else { // otherwise, simply mark them as offline
		r.Players[player] = false
	}
	// close their send channel regardless
	close(player.send)

	if len(r.Players) <= 0 || r.onlinePlayerCount() <= 0 {
		return true
	}

	playerLeftMessage := wsmessages.WSMessageOut{
		Type: wsmessages.PlayerLeft,
		Data: map[string]any{
			"id": player.id,
		},
	}

	r.broadcastMessage(
		&playerLeftMessage,
		func(online bool, p *Client) bool {
			return !online || p.session == player.session
		},
		"error sending player left message to player",
	)

	return false
}

func (r *Room) setupGameForPlayer(player *Client, online bool, treasures []gamedata.TreasureItem, gameMap data.GameMap) {
	player.PrepareForNextRound()
	player.setTreasureSet(treasures)
	r.PlayerModel.UpdateTreasureSet(player.session, treasures)
	if online {
		err := player.Send(wsmessages.WSMessageOut{Type: wsmessages.GameDataTreasures, Data: map[string]any{
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
	r.Broadcast <- &wsmessages.WSMessageOut{Type: wsmessages.ShowGameLoading}
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

	r.Broadcast <- &wsmessages.WSMessageOut{
		Type: wsmessages.BeginGame,
		Data: map[string]any{
			"serverNow": now.UnixMilli(),
			"endsAt":    now.Add(r.roundTime).UnixMilli(),
		},
	}

	r.resetTimer <- roundStartData{roundTime: now.Add(r.roundTime).UnixMilli(), roundMap: gameMap}
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

func (r *Room) startNewGame(client *Client) {
	if r.onlinePlayerCount() != r.maxPlayers {
		client.Send(wsmessages.WSMessageOut{Type: wsmessages.NotEnoughPlayers})
		return
	}

	playerSnapshot := r.getPlayerSnapshot()

	playerCount := len(r.Players)

	go func() {
		r.startRound(playerSnapshot, playerCount, data.GameMapFirst)
	}()
}

func (r *Room) resetAndStartNew(client *Client) {
	if r.State != data.GameStateTransitioning {
		r.logger.Warn("received new game request when game state not in transitioning", "gameState", r.State, "requesterSessio", client.session)
		return
	}

	if r.lastMap != data.GameMapThird {
		r.logger.Warn("received new game request when last map wasn't third", "lastMap", r.lastMap, "requesterSession", client.session)
		return
	}

	resetMessage := wsmessages.WSMessageOut{Type: wsmessages.ResetGame}

	r.broadcastMessage(
		&resetMessage,
		func(online bool, player *Client) bool {
			return !online
		},
		"error sending reset game state message to player",
	)

	r.lastMap = data.GameMapNone
	for player := range r.Players {
		player.Reset()
	}

	r.startNewGame(client)
}

func (r *Room) broadcastMessage(message *wsmessages.WSMessageOut, shouldSkip func(bool, *Client) bool, customErrMsg string) {
	for player, online := range r.Players {
		if shouldSkip(online, player) {
			continue
		}

		err := player.Send(*message)
		if err != nil {
			errMsg := customErrMsg
			if errMsg == "" {
				errMsg = fmt.Sprintf("error sending broadcast message (%s) to player", message.Type)
			}
			r.logger.Error(errMsg, "error", err, "playerId", player.id)
		}
	}
}

func (r *Room) resumeAtTransitioning(requestingPlayer *Client) {
	snapshot := requestingPlayer.Snapshot()

	var existingPlayers []data.PlayerInfo

	for p, online := range r.Players {
		if requestingPlayer != p && online {
			existingPlayers = append(existingPlayers, data.PlayerInfo{ID: p.id, Nickname: p.nickname})
		}
	}

	playerStatsMap := make(map[string]ClientSnapshot)

	for p := range r.Players {
		snapshot := p.Snapshot()
		playerStatsMap[p.id] = snapshot
	}

	allPlayerPoints := make(map[string]int)
	for p := range r.Players {
		allPlayerPoints[p.id] = playerStatsMap[p.id].Points
	}

	err := requestingPlayer.Send(wsmessages.WSMessageOut{Type: wsmessages.ResumeGameTransitioning, Data: map[string]any{
		"map":               r.lastMap,
		"treasures_claimed": playerStatsMap[requestingPlayer.id].SuccessfullyClaimed,
		"treasures_failed":  playerStatsMap[requestingPlayer.id].FailedToClaimCount,
		"all_player_points": allPlayerPoints,
		"inventory":         snapshot.Inventory,
		"treasure_count":    snapshot.TreasureCount,
		"nickname":          requestingPlayer.nickname,
		"player_id":         requestingPlayer.id,
		"points":            snapshot.Points,
		"members":           existingPlayers,
		"is_creator":        requestingPlayer.isCreator,
		"game_info": map[string]any{
			"code":            r.Code,
			"max_players":     r.maxPlayers,
			"current_players": len(existingPlayers),
		},
	}})
	if err != nil {
		r.logger.Error("error sending map summary to rejoining player", "error", err, "playerId", requestingPlayer.id)
	}
}

func (r *Room) resumeMidGame(requestingPlayer *Client) {
	snapshot := requestingPlayer.Snapshot()

	var existingPlayers []data.PlayerInfo

	for p, online := range r.Players {
		if requestingPlayer != p && online {
			existingPlayers = append(existingPlayers, data.PlayerInfo{ID: p.id, Nickname: p.nickname})
		}
	}

	now := time.Now()
	var serverNow int64
	var endsAt int64

	if r.State == data.GameStateInProgress {
		serverNow = now.UnixMilli()
		endsAt = r.roundEndTime
	}

	err := requestingPlayer.Send(wsmessages.WSMessageOut{
		Type: wsmessages.ResumeGameInProgress,
		Data: map[string]any{
			"map":            r.currentMap,
			"inventory":      snapshot.Inventory,
			"treasure_set":   snapshot.TreasureSet,
			"treasure_count": snapshot.TreasureCount,
			"nickname":       requestingPlayer.nickname,
			"player_id":      requestingPlayer.id,
			"points":         snapshot.Points,
			"members":        existingPlayers,
			"is_creator":     requestingPlayer.isCreator,
			"server_now":     serverNow,
			"ends_at":        endsAt,
			"game_info": map[string]any{
				"code":            r.Code,
				"max_players":     r.maxPlayers,
				"current_players": len(existingPlayers),
			},
		},
	})
	if err != nil {
		r.logger.Error("failed to send resume game message", "playerSession", requestingPlayer.session, "error", err)
	}
}

func (r *Room) startNextRound(client *Client) {
	if r.State != data.GameStateTransitioning {
		r.logger.Warn("received next round request when game state not in transitioning", "gameState", r.State, "requesterSession", client.session)
		return
	}

	if r.lastMap == data.GameMapThird {
		r.logger.Warn("received next round request when last game map was third", "requesterSession", client.session)
		return
	}

	playerSnapshot := r.getPlayerSnapshot()

	playerCount := len(r.Players)

	var nextMap data.GameMap

	if r.lastMap == data.GameMapFirst {
		nextMap = data.GameMapSecond
	} else {
		nextMap = data.GameMapThird
	}

	go func() {
		r.startRound(playerSnapshot, playerCount, nextMap)
	}()
}

func (r *Room) endRound() {
	if r.State != data.GameStateInProgress {
		return
	}

	r.lastMap = r.currentMap

	r.State = data.GameStateTransitioning
	r.currentMap = data.GameMapNone

	playerStatsMap := make(map[string]ClientSnapshot)

	// Backup game and player states
	r.GameModel.SetGameAndMapState(r.Code, r.State, r.currentMap)
	for p := range r.Players {
		snapshot := p.Snapshot()
		playerStatsMap[p.id] = snapshot
		err := r.PlayerModel.BackupPlayerStats(p.session, snapshot.TreasureSet, snapshot.TreasureCount, snapshot.Inventory, snapshot.Points)
		if err != nil {
			r.logger.Error("error backing up player stats after round over", "session_token", p.session, "error", err.Error())
		}
	}

	allPlayerPoints := make(map[string]int)
	for p := range r.Players {
		allPlayerPoints[p.id] = playerStatsMap[p.id].Points
	}

	for p, online := range r.Players {
		if !online {
			continue
		}
		err := p.Send(wsmessages.WSMessageOut{Type: wsmessages.MapRoundOver, Data: map[string]any{
			"map":               r.lastMap,
			"treasures_claimed": playerStatsMap[p.id].SuccessfullyClaimed,
			"treasures_failed":  playerStatsMap[p.id].FailedToClaimCount,
			"all_player_points": allPlayerPoints,
		}})
		if err != nil {
			r.logger.Error("error sending map round over message to player", "error", err, "playerId", p.id)
		}
	}
}

func (r *Room) getPlayerClientFromId(id string) *Client {
	for player := range r.Players {
		if player.id == id {
			return player
		}
	}

	return nil
}

func (r *Room) sendMessageToPlayer(player *Client, message wsmessages.WSMessageOut, customErr string) bool {
	if online := r.Players[player]; !online {
		r.logger.Error("sendMessageToPlayer: player not online; skipping send message", "playerId", player.id)
		return false
	}

	err := player.Send(message)
	errMsg := customErr
	if errMsg == "" {
		errMsg = "error sending message to player"
	}
	if err != nil {
		r.logger.Error(errMsg, "error", err, "playerId", player.id)
		return false
	}
	return true
}
