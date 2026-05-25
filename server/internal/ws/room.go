package ws

import (
	"log/slog"
	"time"

	"game.crisis-control.com/internal/data"
	"game.crisis-control.com/internal/engine"
	"game.crisis-control.com/internal/gamedata"
	"game.crisis-control.com/internal/wsmessages"
)

const (
	RoundTime      = 5 * time.Minute
	BackupTime     = 15 * time.Second
	TTLResetPeriod = 2 * time.Minute
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

type Room struct {
	Players               map[*Client]bool
	Register              chan *Client
	Unregister            chan *Client
	StartGame             chan *Client
	NextRound             chan *Client
	Shutdown              chan struct{}
	GameModel             *data.GameModel
	PlayerModel           *data.PlayerModel
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
}

func NewRoom(models *data.Models, code string, maxPlayers int, gm *data.GameModel, pm *data.PlayerModel, logger *slog.Logger, engine *engine.Engine) *Room {
	return &Room{
		Players:               make(map[*Client]bool),
		Register:              make(chan *Client),
		Unregister:            make(chan *Client),
		StartGame:             make(chan *Client),
		NextRound:             make(chan *Client),
		Shutdown:              make(chan struct{}),
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
		timer:                 time.NewTimer(RoundTime),
		backupTimer:           time.NewTimer(BackupTime),
		resetTimer:            make(chan roundStartData),
		roundEndTime:          0,
	}
}

func (r *Room) setupGameForPlayer(player *Client, treasures []gamedata.TreasureItem, gameMap data.GameMap) {
	player.ResetState()
	player.setTreasureSet(treasures)
	r.PlayerModel.UpdateTreasureSet(player.session, treasures)
	err := player.Send(wsmessages.WSMessage{Type: wsmessages.GameDataTreasures, Data: map[string]any{
		"treasures": treasures,
		"map":       gameMap,
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

func (r *Room) setTTLForGameAndPlayers(playerSnapshot []*Client) {
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
			continue
		}
	}
}

func (r *Room) startRound(playerSnapshot []*Client, playerCount int, gameMap data.GameMap) {
	r.Broadcast <- &wsmessages.WSMessage{Type: wsmessages.ShowGameLoading}
	treasures := r.engine.GenerateTreasureSet(playerCount, gameMap)
	for i, player := range playerSnapshot {
		r.setupGameForPlayer(player, treasures[i], gameMap)
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

func (r *Room) Run() {
	ttlResetTicker := time.NewTicker(TTLResetPeriod)

	defer func() {
		ttlResetTicker.Stop()
	}()

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
			if !client.isCreator {
				r.logger.Warn("non-creator player requested game start", "playerToken", client.session)
				break
			}

			if len(r.Players) != r.maxPlayers {
				client.Send(wsmessages.WSMessage{Type: wsmessages.NotEnoughPlayers})
				break
			}

			playerSnapshot := make([]*Client, 0, len(r.Players))
			for p := range r.Players {
				playerSnapshot = append(playerSnapshot, p)
			}

			playerCount := len(r.Players)

			go func() {
				r.startRound(playerSnapshot, playerCount, data.GameMapFirst)
			}()
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

			playerSnapshot := make([]*Client, 0, len(r.Players))
			for p := range r.Players {
				playerSnapshot = append(playerSnapshot, p)
			}

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

			for p := range r.Players {
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

			err = r.GameModel.RemovePlayerFromSet(r.Code, player.session)
			if err != nil {
				r.logger.Error("error removing player from set", "error", err)
			}

			delete(r.Players, player)
			close(player.send)

			for p := range r.Players {
				err = p.Send(wsmessages.WSMessage{Type: wsmessages.PlayerLeft, Data: map[string]any{
					"id": player.id,
				}})
				if err != nil {
					r.logger.Error("error sending player left message to player", "error", err, "playerId", p.id)
				}
			}

			if len(r.Players) <= 0 {
				return
			}
		case message := <-r.Broadcast:
			for player := range r.Players {
				err := player.Send(*message)
				if err != nil {
					r.logger.Error("error sending broadcast message to player", "error", err, "playerId", player.id)
				}
			}
		case gameData := <-r.resetTimer:
			r.State = data.GameStateInProgress
			r.currentMap = gameData.roundMap
			r.timer.Reset(RoundTime)
			r.backupTimer.Reset(BackupTime)
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
			r.backupTimer.Reset(BackupTime)
		case <-r.timer.C:
			if r.State != data.GameStateInProgress {
				break
			}

			r.lastMap = r.currentMap

			r.logger.Debug("lastmap set", "lastmap", r.lastMap)

			r.State = data.GameStateTransitioning
			r.currentMap = data.GameMapNone

			playerStatsMap := make(map[string]ClientSnapshot)

			// Backup game and player states here
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

			for p := range r.Players {
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
			playerSnapshot := make([]*Client, 0, len(r.Players))
			for p := range r.Players {
				playerSnapshot = append(playerSnapshot, p)
			}

			r.setTTLForGameAndPlayers(playerSnapshot)
		}
	}
}
