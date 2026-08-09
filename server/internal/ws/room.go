package ws

import (
	"log/slog"
	"time"

	"atlas.sageyutta.com/internal/data"
	"atlas.sageyutta.com/internal/engine"
	"atlas.sageyutta.com/internal/gamedata"
	"atlas.sageyutta.com/internal/wsmessages"
)

type playerLookupRequest struct {
	session string
	result  chan *Client
}

type playerNotification struct {
	Client  *Client
	Message wsmessages.WSMessage
}

type roundStartData struct {
	roundTime int64
	roundMap  data.GameMap
}

type Player interface {
	Delete(session string) error
	BackupPlayerStats(token string, treasureSet []gamedata.TreasureItem, treasureCount map[string]int, inventory []gamedata.InventoryItem, points int) error
	UpdateTreasureSet(token string, treasureSet []gamedata.TreasureItem) error
	SetTTL(token string) error
}

type Game interface {
	GetAllPlayers(gameCode string) ([]string, error)
	CleanupGame(gameCode string) error
	IncrementPlayerCount(gameCode string) error
	DecrementPlayerCount(gameCode string) error
	RemovePlayerFromSet(gameCode, playerSessionToken string) error
	SetGameAndMapState(code string, state data.GameState, gameMap data.GameMap) error
	SetTTL(code string) error
	SetTTLForPlayerSet(gameCode string, ttl time.Duration) error
}

type Room struct {
	Players               map[*Client]bool
	Register              chan *Client
	Unregister            chan *Client
	StartGame             chan *Client
	NextRound             chan *Client
	NewGame               chan *Client
	Shutdown              chan struct{}
	ResumeGame            chan *Client
	GameModel             Game
	PlayerModel           Player
	Code                  string
	State                 data.GameState
	lastMap               data.GameMap
	currentMap            data.GameMap
	logger                *slog.Logger
	maxPlayers            int
	Broadcast             chan *wsmessages.WSMessage
	engine                *engine.Engine
	PlayerLookup          chan playerLookupRequest
	NotifyTreasureClaimed chan playerNotification
	NotifyBonusClaimed    chan playerNotification
	timer                 *time.Timer
	resetTimer            chan roundStartData
	backupTimer           *time.Timer
	roundEndTime          int64
	onExit                func(code string)
	roundTime             time.Duration
	backupTime            time.Duration
	ttlResetPeriod        time.Duration
}

func NewRoom(
	code string,
	maxPlayers int,
	gm Game,
	pm Player,
	logger *slog.Logger,
	engine *engine.Engine,
	onExit func(code string),
	roundTime,
	backupTime,
	ttlResetPeriod time.Duration,
) *Room {
	return &Room{
		Players:               make(map[*Client]bool),
		Register:              make(chan *Client),
		Unregister:            make(chan *Client),
		StartGame:             make(chan *Client),
		NextRound:             make(chan *Client),
		NewGame:               make(chan *Client),
		Shutdown:              make(chan struct{}),
		ResumeGame:            make(chan *Client),
		GameModel:             gm,
		PlayerModel:           pm,
		currentMap:            data.GameMapNone,
		lastMap:               data.GameMapNone,
		State:                 data.GameStateLobby,
		Code:                  code,
		maxPlayers:            maxPlayers,
		logger:                logger,
		Broadcast:             make(chan *wsmessages.WSMessage),
		engine:                engine,
		PlayerLookup:          make(chan playerLookupRequest),
		NotifyTreasureClaimed: make(chan playerNotification),
		NotifyBonusClaimed:    make(chan playerNotification),
		timer:                 time.NewTimer(roundTime),
		backupTimer:           time.NewTimer(backupTime),
		resetTimer:            make(chan roundStartData),
		roundEndTime:          0,
		onExit:                onExit,

		roundTime:      roundTime,
		backupTime:     backupTime,
		ttlResetPeriod: ttlResetPeriod,
	}
}

func (r *Room) Run() {
	ttlResetTicker := time.NewTicker(r.ttlResetPeriod)

	defer func() {
		ttlResetTicker.Stop()
		players, err := r.GameModel.GetAllPlayers(r.Code)
		if err != nil {
			r.logger.Error("error retrieving players for game room", "gameCode", r.Code, "error", err)
		} else {
			for _, session := range players {
				err := r.PlayerModel.Delete(session)
				if err != nil {
					r.logger.Error("error deleting player from redis", "playerSession", session, "error", err)
				}
			}
		}
		err = r.GameModel.CleanupGame(r.Code)
		if err != nil {
			r.logger.Error("error cleaning up game from redis", "gameCode", r.Code, "error", err)
		}
		r.onExit(r.Code)
	}()

	for {
		select {
		case <-r.Shutdown:
			for player := range r.Players {
				player.conn.Close()
			}

			// drain unregister channel and delete online players first
			for r.onlinePlayerCount() > 0 {
				player := <-r.Unregister
				delete(r.Players, player)
			}
			// delete offline players
			for player := range r.Players {
				delete(r.Players, player)
			}
			return
		case player := <-r.ResumeGame:
			snapshot := player.Snapshot()

			var existingPlayers []data.PlayerInfo

			for p, online := range r.Players {
				if player != p && online {
					existingPlayers = append(existingPlayers, data.PlayerInfo{ID: p.id, Nickname: p.nickname})
				}
			}

			if r.State == data.GameStateTransitioning {
				playerStatsMap := make(map[string]ClientSnapshot)

				for p := range r.Players {
					snapshot := p.Snapshot()
					playerStatsMap[p.id] = snapshot
				}

				allPlayerPoints := make(map[string]int)
				for p := range r.Players {
					allPlayerPoints[p.id] = playerStatsMap[p.id].Points
				}

				err := player.Send(wsmessages.WSMessage{Type: wsmessages.ResumeGameTransitioning, Data: map[string]any{
					"map":               r.lastMap,
					"treasures_claimed": playerStatsMap[player.id].SuccessfullyClaimed,
					"treasures_failed":  playerStatsMap[player.id].FailedToClaimCount,
					"all_player_points": allPlayerPoints,
					"inventory":         snapshot.Inventory,
					"treasure_count":    snapshot.TreasureCount,
					"nickname":          player.nickname,
					"player_id":         player.id,
					"points":            snapshot.Points,
					"members":           existingPlayers,
					"is_creator":        player.isCreator,
					"game_info": map[string]any{
						"code":            r.Code,
						"max_players":     r.maxPlayers,
						"current_players": len(existingPlayers),
					},
				}})
				if err != nil {
					r.logger.Error("error sending map summary to rejoining player", "error", err, "playerId", player.id)
				}
				break
			}

			now := time.Now()
			var serverNow int64
			var endsAt int64

			if r.State == data.GameStateInProgress {
				serverNow = now.UnixMilli()
				endsAt = r.roundEndTime
			}

			err := player.Send(wsmessages.WSMessage{
				Type: wsmessages.ResumeGameInProgress,
				Data: map[string]any{
					"map":            r.currentMap,
					"inventory":      snapshot.Inventory,
					"treasure_set":   snapshot.TreasureSet,
					"treasure_count": snapshot.TreasureCount,
					"nickname":       player.nickname,
					"player_id":      player.id,
					"points":         snapshot.Points,
					"members":        existingPlayers,
					"is_creator":     player.isCreator,
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
				r.logger.Error("failed to send resume game message", "playerSession", player.session, "error", err)
			}
		case client := <-r.StartGame:
			if !client.isCreator {
				r.logger.Warn("non-creator player requested game start", "playerToken", client.session)
				break
			}

			if r.onlinePlayerCount() != r.maxPlayers {
				client.Send(wsmessages.WSMessage{Type: wsmessages.NotEnoughPlayers})
				break
			}

			playerSnapshot := r.getPlayerSnapshot()

			playerCount := len(r.Players)

			go func() {
				r.startRound(playerSnapshot, playerCount, data.GameMapFirst)
			}()
		case client := <-r.NewGame:
			if !client.isCreator {
				r.logger.Warn("non-creator player requested new game", "playerToken", client.session)
				break
			}

			if r.State != data.GameStateTransitioning {
				r.logger.Warn("received new game request when game state not in transitioning", "gameState", r.State, "requesterSessio", client.session)
				break
			}

			if r.lastMap != data.GameMapThird {
				r.logger.Warn("received new game request when last map wasn't third", "lastMap", r.lastMap, "requesterSession", client.session)
				break
			}

			for player, online := range r.Players {
				if !online {
					continue
				}

				err := player.Send(wsmessages.WSMessage{Type: wsmessages.ResetGame})
				if err != nil {
					r.logger.Error("error sending reset game state message to player", "playerSession", player.session, "error", err)
				}
			}

			r.newGame(client)
		case client := <-r.NextRound:
			if !client.isCreator {
				r.logger.Warn("non-creator player requested next round start", "playerToken", client.session)
				break
			}

			if r.State != data.GameStateTransitioning {
				r.logger.Warn("received next round request when game state not in transitioning", "gameState", r.State, "requesterSession", client.session)
				break
			}

			if r.lastMap == data.GameMapThird {
				r.logger.Warn("received next round request when last game map was third", "requesterSession", client.session)
				break
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
		case req := <-r.PlayerLookup:
			var found *Client
			for player := range r.Players {
				if player.session == req.session {
					found = player
					break
				}
			}
			req.result <- found
		case notification := <-r.NotifyTreasureClaimed:
			err := notification.Client.Send(notification.Message)
			if err != nil {
				r.logger.Error("error updating player stats", "playerSession", notification.Client.session)
			}
		case notification := <-r.NotifyBonusClaimed:
			err := notification.Client.Send(notification.Message)
			if err != nil {
				r.logger.Error("error notifying bonus claim", "playerSession", notification.Client.session)
			}
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

			for p, online := range r.Players {
				if !online || p.session == player.session {
					continue
				}
				err = p.Send(wsmessages.WSMessage{Type: wsmessages.PlayerJoined, Data: map[string]any{
					"nickname": player.nickname,
					"id":       player.id,
				}})
				if err != nil {
					r.logger.Error("error sending player joined message to player", "error", err, "playerId", p.id)
				}
			}
		case player := <-r.Unregister:
			err := r.GameModel.DecrementPlayerCount(r.Code)
			if err != nil {
				r.logger.Error("error decrementing player count in redis", "error", err)
			}

			if r.State == data.GameStateLobby { // if game is still in lobby, delete the player
				delete(r.Players, player)
				err := r.GameModel.RemovePlayerFromSet(r.Code, player.session)
				if err != nil {
					r.logger.Error("error removing player from set in redis", "playerSession", player.session, "error", err)
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
				return
			}

			for p, online := range r.Players {
				if !online || p.session == player.session {
					continue
				}
				err = p.Send(wsmessages.WSMessage{Type: wsmessages.PlayerLeft, Data: map[string]any{
					"id": player.id,
				}})
				if err != nil {
					r.logger.Error("error sending player left message to player", "error", err, "playerId", p.id)
				}
			}

		case message := <-r.Broadcast:
			for player, online := range r.Players {
				if !online {
					continue
				}
				err := player.Send(*message)
				if err != nil {
					r.logger.Error("error sending broadcast message to player", "error", err, "playerId", player.id)
				}
			}
		case gameData := <-r.resetTimer:
			r.State = data.GameStateInProgress
			r.currentMap = gameData.roundMap
			r.timer.Reset(r.roundTime)
			r.backupTimer.Reset(r.backupTime)
			r.roundEndTime = gameData.roundTime
		case <-r.backupTimer.C:
			if r.State != data.GameStateInProgress {
				break
			}

			for p := range r.Players {
				snapshot := p.Snapshot()
				err := r.PlayerModel.BackupPlayerStats(p.session, snapshot.TreasureSet, snapshot.TreasureCount, snapshot.Inventory, snapshot.Points)
				if err != nil {
					r.logger.Error("error backing up player stats", "session_token", p.session, "error", err.Error())
				}
			}
			r.backupTimer.Reset(r.backupTime)
		case <-r.timer.C:
			if r.State != data.GameStateInProgress {
				break
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
				err := p.Send(wsmessages.WSMessage{Type: wsmessages.MapRoundOver, Data: map[string]any{
					"map":               r.lastMap,
					"treasures_claimed": playerStatsMap[p.id].SuccessfullyClaimed,
					"treasures_failed":  playerStatsMap[p.id].FailedToClaimCount,
					"all_player_points": allPlayerPoints,
				}})
				if err != nil {
					r.logger.Error("error sending map round over message to player", "error", err, "playerId", p.id)
				}
			}
		case <-ttlResetTicker.C:
			playerSnapshot := r.getPlayerSnapshot()

			r.setTTLForGameAndPlayers(playerSnapshot)
		}
	}
}
