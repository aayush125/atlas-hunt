package ws

import (
	"log/slog"
	"time"

	"atlas.sageyutta.com/internal/data"
	"atlas.sageyutta.com/internal/engine"
	"atlas.sageyutta.com/internal/wsmessages"
)

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
		Broadcast:             make(chan *wsmessages.WSMessageOut),
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

		treasureRequests:     make(map[string]treasureRequestState),
		treasureRequestChan:  make(chan treasureRequestPayload),
		treasureDonationChan: make(chan treasureDonationPayload),
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
			if r.State == data.GameStateTransitioning {
				r.resumeAtTransitioning(player)
			} else {
				r.resumeMidGame(player)
			}
		case client := <-r.StartGame:
			if !client.isCreator {
				r.logger.Warn("non-creator player requested game start", "playerToken", client.session)
				break
			}

			r.startNewGame(client)
		case client := <-r.NewGame:
			if !client.isCreator {
				r.logger.Warn("non-creator player requested new game", "playerToken", client.session)
				break
			}

			r.resetAndStartNew(client)
		case client := <-r.NextRound:
			if !client.isCreator {
				r.logger.Warn("non-creator player requested next round start", "playerToken", client.session)
				break
			}

			r.startNextRound(client)
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

			treasureClaimedMessage := wsmessages.WSMessageOut{
				Type: wsmessages.PlayerClaimedTreasure,
				Data: map[string]any{
					"claimer": notification.Client.nickname,
					"rarity":  notification.Info["rarity"],
				},
			}

			r.broadcastMessage(
				&treasureClaimedMessage,
				func(online bool, player *Client) bool {
					return !online || player == notification.Client
				},
				"error broadcasting treasure claim notification to player",
			)

		case notification := <-r.NotifyBonusClaimed:
			err := notification.Client.Send(notification.Message)
			if err != nil {
				r.logger.Error("error notifying bonus claim", "playerSession", notification.Client.session)
			}
		case player := <-r.Register:
			r.registerPlayer(player)
		case player := <-r.Unregister:
			zeroPlayersLeft := r.unregisterPlayer(player)
			if zeroPlayersLeft {
				return
			}
		case message := <-r.Broadcast:
			r.broadcastMessage(
				message,
				func(online bool, player *Client) bool {
					return !online
				},
				"error sending broadcast message to player",
			)
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
			r.endRound()
		case <-ttlResetTicker.C:
			playerSnapshot := r.getPlayerSnapshot()

			r.setTTLForGameAndPlayers(playerSnapshot)
		case request := <-r.treasureRequestChan:
			requester := r.validateTreasureRequest(request)

			if requester == nil { // break if nil; error already handled by validator
				break
			}

			r.createTreasureRequest(request, requester)
		case donation := <-r.treasureDonationChan:
			donor, receiver := r.validateDonationRequest(donation)
			shouldBreak := donor == nil || receiver == nil

			if shouldBreak { // break if nil; error already handled by validator
				break
			}

			r.constructTreasureSetAndDonate(donation, donor, receiver)
		}
	}
}
