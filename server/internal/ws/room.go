package ws

import (
	"log/slog"
	"time"

	"atlas.sageyutta.com/internal/data"
	"atlas.sageyutta.com/internal/engine"
	"atlas.sageyutta.com/internal/gamedata"
	"atlas.sageyutta.com/internal/wsmessages"
	"github.com/google/uuid"
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
			// basic validation
			requester := r.getPlayerClientFromId(request.PlayerId)
			if requester == nil {
				r.sendMessageToPlayer(
					requester,
					wsmessages.WSMessageOut{
						Type: wsmessages.ErrorDonatingTreasure,
						Data: map[string]any{
							"request": request,
						},
					},
					"error sending ErrorDonatingTreasure message",
				)
				break
			}

			if request.RequestedAmount <= 0 {
				r.sendMessageToPlayer(
					requester,
					wsmessages.WSMessageOut{
						Type: wsmessages.ErrorInvalidTreasureRequest,
						Data: map[string]any{
							"request": request,
						},
					},
					"error sending ErrorInvalidTreasureRequest message",
				)
				break
			}

			if !requester.canReceiveTreasure(request.RequestedAmount) {
				r.sendMessageToPlayer(
					requester,
					wsmessages.WSMessageOut{
						Type: wsmessages.InsufficientInventorySpace,
						Data: map[string]any{
							"request": request,
						},
					},
					"error sending InsufficientInventorySpace messagge",
				)
				break
			}

			_, ok := r.treasureRequests[request.PlayerId]

			if ok {
				r.sendMessageToPlayer(
					requester,
					wsmessages.WSMessageOut{
						Type: wsmessages.TreasureRequestAlreadyExists,
						Data: map[string]any{
							"request": request,
						},
					},
					"error sending ErrorInvalidTreasureRequest message",
				)
				break
			}

			r.treasureRequests[request.PlayerId] = treasureRequestState{
				TreasureId:      request.TreasureId,
				RequestedAmount: request.RequestedAmount,
			}

			r.broadcastMessage(
				&wsmessages.WSMessageOut{
					Type: wsmessages.UpdateTreasureRequests,
					Data: map[string]any{
						"requests": r.treasureRequests,
					},
				},
				func(online bool, player *Client) bool {
					return !online
				},
				"error sending UpdateTreasureRequests message to player",
			)
		case donation := <-r.treasureDonationChan:
			// basic validation
			donor := r.getPlayerClientFromId(donation.From)
			receiver := r.getPlayerClientFromId(donation.To)

			if donor == nil || receiver == nil {
				r.logger.Error("receiver or donor nil")
				break
			}

			if donation.DonatedAmount < 0 {
				r.sendMessageToPlayer(
					donor,
					wsmessages.WSMessageOut{
						Type: wsmessages.ErrorInvalidTreasureDonateRequest,
						Data: map[string]any{
							"request": donation,
						},
					},
					"error sending ErrorInvalidTreasureDonateRequest message",
				)
				break
			}

			if donation.From == donation.To {
				r.sendMessageToPlayer(
					donor,
					wsmessages.WSMessageOut{
						Type: wsmessages.ErrorInvalidTreasureDonateRequest,
						Data: map[string]any{
							"request": donation,
						},
					},
					"error sending ErrorInvalidTreasureDonateRequest message",
				)
				break
			}

			// check if receiving player has an active request
			req, ok := r.treasureRequests[donation.To]
			if !ok {
				r.sendMessageToPlayer(
					donor,
					wsmessages.WSMessageOut{
						Type: wsmessages.ErrorTreasureRequestNotFound,
						Data: map[string]any{
							"request": donation,
						},
					},
					"error sending ErrorTreasureRequestNotFound message",
				)
				break
			}

			// ensure the requesting player has inventory space
			canReceive := receiver.canReceiveTreasure(donation.DonatedAmount)
			if !canReceive {
				r.sendMessageToPlayer(
					donor,
					wsmessages.WSMessageOut{
						Type: wsmessages.ErrorReceiverInventoryInsufficient,
						Data: map[string]any{
							"request": donation,
						},
					},
					"error sending ErrorReceiverInventoryInsufficient message",
				)
				break
			}

			// ensure donated amount is less than or equal to available slots
			if donation.DonatedAmount > req.RequestedAmount-req.AmountReceived {
				// can't donate more than available slots
				r.sendMessageToPlayer(
					donor,
					wsmessages.WSMessageOut{
						Type: wsmessages.ErrorDonationSlotsInsufficient,
						Data: map[string]any{
							"request": donation,
						},
					},
					"error sending ErrorDonationSlotsInsufficient message",
				)
				break
			}

			treasureInfo := r.engine.GetTreasureInfoFromId(donation.TreasureId)
			if treasureInfo == nil {
				// return an "error donating treasure" to client
				r.sendMessageToPlayer(
					donor,
					wsmessages.WSMessageOut{
						Type: wsmessages.ErrorDonatingTreasure,
						Data: map[string]any{
							"request": donation,
						},
					},
					"error sending ErrorDonatingTreasure message",
				)
				break
			}

			// construct the treasure set
			set := []gamedata.InventoryItem{}
			for _ = range donation.DonatedAmount {
				treasure_uuid, err := uuid.NewRandom()
				if err != nil {
					r.logger.Error("error generating uuid for donated treasure item", "error", err, "donatingPlayerId", donation.From)
					break
				}
				treasureItem := gamedata.InventoryItem{
					InventoryItemBase: gamedata.InventoryItemBase{
						InstanceID:  treasure_uuid,
						ID:          treasureInfo.ID,
						Name:        treasureInfo.Name,
						Icon:        treasureInfo.Icon,
						Cat:         treasureInfo.Cat,
						Rarity:      treasureInfo.Rarity,
						Points:      treasureInfo.Points,
						BoostPoints: treasureInfo.BoostPoints,
					},
				}
				set = append(set, treasureItem)
			}

			if len(set) < donation.DonatedAmount {
				r.sendMessageToPlayer(
					donor,
					wsmessages.WSMessageOut{
						Type: wsmessages.ErrorDonatingTreasure,
						Data: map[string]any{
							"request": donation,
						},
					},
					"error sending ErrorDonatingTreasure message",
				)
				break
			}

			// finally, donate
			err := donor.tryDonateTreasure(donation.TreasureId, donation.DonatedAmount)
			if err != nil {
				r.sendMessageToPlayer(
					donor,
					wsmessages.WSMessageOut{
						Type: wsmessages.ErrorInsufficientBalance,
						Data: map[string]any{
							"request": donation,
						},
					},
					"error sending ErrorInsufficientBalance message",
				)
				break
			}

			// receive treasure
			receiver.receiveTreasure(donation.TreasureId, set)

			// remove or update treasure request
			req.AmountReceived += donation.DonatedAmount
			if req.AmountReceived >= req.RequestedAmount {
				delete(r.treasureRequests, donation.To)
			} else {
				r.treasureRequests[donation.To] = req
			}

			donorSnapshot := donor.Snapshot()
			r.sendMessageToPlayer(
				donor,
				wsmessages.WSMessageOut{
					Type: wsmessages.TreasureDonatedSuccessfully,
					Data: map[string]any{
						"treasureId":    donation.TreasureId,
						"donatedAmount": donation.DonatedAmount,
						"donatedTo":     donation.To,
						"inventory":     donorSnapshot.Inventory,
						"treasureCount": donorSnapshot.TreasureCount,
						"points":        donorSnapshot.Points,
					},
				},
				"error sending TreasureDonatedSuccessfully message",
			)

			receiverSnapshot := receiver.Snapshot()
			r.sendMessageToPlayer(
				receiver,
				wsmessages.WSMessageOut{
					Type: wsmessages.TreasureReceivedSuccessfully,
					Data: map[string]any{
						"treasureId":     donation.TreasureId,
						"receivedAmount": donation.DonatedAmount,
						"donatedBy":      donation.From,
						"inventory":      receiverSnapshot.Inventory,
						"treasureCount":  receiverSnapshot.TreasureCount,
					},
				},
				"error sending TreasureReceivedSuccessfully message",
			)

			r.broadcastMessage(
				&wsmessages.WSMessageOut{
					Type: wsmessages.UpdateTreasureRequests,
					Data: map[string]any{
						"requests": r.treasureRequests,
					},
				},
				func(online bool, player *Client) bool {
					return !online
				},
				"error sending UpdateTreasureRequests message to player",
			)
		}
	}
}
