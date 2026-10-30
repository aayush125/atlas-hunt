package ws

import (
	"atlas.sageyutta.com/internal/gamedata"
	"atlas.sageyutta.com/internal/wsmessages"
	"github.com/google/uuid"
)

func (r *Room) validateTreasureRequest(request treasureRequestPayload) *Client {
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
		return nil
	}

	if !gamedata.TreasureExists(request.TreasureId) || request.RequestedAmount <= 0 {
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
		return nil
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
		return nil
	}

	return requester
}

func (r *Room) createTreasureRequest(request treasureRequestPayload, requester *Client) {
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
		return
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
}

func (r *Room) validateDonationRequest(donation treasureDonationPayload) (*Client, *Client) {
	// basic validation
	donor := r.getPlayerClientFromId(donation.From)
	receiver := r.getPlayerClientFromId(donation.To)

	if donor == nil || receiver == nil {
		r.logger.Error("receiver or donor nil")
		return nil, nil
	}

	if donation.DonatedAmount <= 0 {
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
		return nil, nil
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
		return nil, nil
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
		return nil, nil
	}

	if donation.TreasureId != req.TreasureId {
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
		return nil, nil
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
		return nil, nil
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
		return nil, nil
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
		return nil, nil
	}

	return donor, receiver
}

func (r *Room) constructTreasureSetAndDonate(donation treasureDonationPayload, donor, receiver *Client) {
	treasureInfo := r.engine.GetTreasureInfoFromId(donation.TreasureId)
	req := r.treasureRequests[donation.To]

	// construct the treasure set
	set := []gamedata.InventoryItem{}
	for range donation.DonatedAmount {
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
		return
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
		return
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

func (r *Room) deleteTreasureRequestForPlayer(player *Client) {
	_, ok := r.treasureRequests[player.id]
	if !ok {
		r.logger.Error("received treasure request delete for non-existent request")
		return
	}

	delete(r.treasureRequests, player.id)

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
