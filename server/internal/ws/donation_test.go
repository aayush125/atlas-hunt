package ws

import (
	"testing"

	"atlas.sageyutta.com/internal/gamedata"
	"atlas.sageyutta.com/internal/wsmessages"
)

func TestTreasureDonationValidation(t *testing.T) {
	validTreasureID := gamedata.GetAllTreasures()[0].ID

	tests := []struct {
		name            string
		treasureRequest func(receiver, donor *Client) (string, treasureRequestState)
		donationRequest func(donor, receiver *Client) treasureDonationPayload
		receiverSetup   func(t *testing.T, p *Client)
		wantReturns     bool
		wantDonorMsg    wsmessages.WSEvent
	}{
		{
			name: "valid donation",
			treasureRequest: func(receiver, donor *Client) (string, treasureRequestState) {
				return receiver.id, treasureRequestState{TreasureId: validTreasureID, RequestedAmount: 10, AmountReceived: 0}
			},
			donationRequest: func(donor, receiver *Client) treasureDonationPayload {
				return treasureDonationPayload{From: donor.id, To: receiver.id, TreasureId: validTreasureID, DonatedAmount: 10}
			},
			wantReturns: true,
		},
		{
			name: "invalid donor",
			treasureRequest: func(receiver, donor *Client) (string, treasureRequestState) {
				return receiver.id, treasureRequestState{TreasureId: validTreasureID, RequestedAmount: 10, AmountReceived: 0}
			},
			donationRequest: func(donor, receiver *Client) treasureDonationPayload {
				return treasureDonationPayload{From: "invalid-donor-id", To: receiver.id, TreasureId: validTreasureID, DonatedAmount: 10}
			},
		},
		{
			name: "invalid receiver",
			treasureRequest: func(receiver, donor *Client) (string, treasureRequestState) {
				return receiver.id, treasureRequestState{TreasureId: validTreasureID, RequestedAmount: 10, AmountReceived: 0}
			},
			donationRequest: func(donor, receiver *Client) treasureDonationPayload {
				return treasureDonationPayload{From: donor.id, To: "invalid-receiver-id", TreasureId: validTreasureID, DonatedAmount: 10}
			},
			wantDonorMsg: wsmessages.ErrorInvalidTreasureDonateRequest,
		},
		{
			name: "invalid donor and receiver",
			treasureRequest: func(receiver, donor *Client) (string, treasureRequestState) {
				return receiver.id, treasureRequestState{TreasureId: validTreasureID, RequestedAmount: 10, AmountReceived: 0}
			},
			donationRequest: func(donor, receiver *Client) treasureDonationPayload {
				return treasureDonationPayload{From: "invalid-donor-id", To: "invalid-receiver-id", TreasureId: validTreasureID, DonatedAmount: 10}
			},
		},
		{
			name: "negative donation amount",
			treasureRequest: func(receiver, donor *Client) (string, treasureRequestState) {
				return receiver.id, treasureRequestState{TreasureId: validTreasureID, RequestedAmount: 10, AmountReceived: 0}
			},
			donationRequest: func(donor, receiver *Client) treasureDonationPayload {
				return treasureDonationPayload{From: donor.id, To: receiver.id, TreasureId: validTreasureID, DonatedAmount: -1}
			},
			wantDonorMsg: wsmessages.ErrorInvalidTreasureDonateRequest,
		},
		{
			name: "zero donation amount",
			treasureRequest: func(receiver, donor *Client) (string, treasureRequestState) {
				return receiver.id, treasureRequestState{TreasureId: validTreasureID, RequestedAmount: 10, AmountReceived: 0}
			},
			donationRequest: func(donor, receiver *Client) treasureDonationPayload {
				return treasureDonationPayload{From: donor.id, To: receiver.id, TreasureId: validTreasureID, DonatedAmount: 0}
			},
			wantDonorMsg: wsmessages.ErrorInvalidTreasureDonateRequest,
		},
		{
			name: "donating to self",
			treasureRequest: func(receiver, donor *Client) (string, treasureRequestState) {
				return donor.id, treasureRequestState{TreasureId: validTreasureID, RequestedAmount: 10, AmountReceived: 0}
			},
			donationRequest: func(donor, receiver *Client) treasureDonationPayload {
				return treasureDonationPayload{From: donor.id, To: donor.id, TreasureId: validTreasureID, DonatedAmount: 10}
			},
			wantDonorMsg: wsmessages.ErrorInvalidTreasureDonateRequest,
		},
		{
			name: "non-existent treasure request",
			donationRequest: func(donor, receiver *Client) treasureDonationPayload {
				return treasureDonationPayload{From: donor.id, To: receiver.id, TreasureId: validTreasureID, DonatedAmount: 10}
			},
			wantDonorMsg: wsmessages.ErrorTreasureRequestNotFound,
		},
		{
			name: "invalid treasure id",
			treasureRequest: func(receiver, donor *Client) (string, treasureRequestState) {
				return receiver.id, treasureRequestState{TreasureId: validTreasureID, RequestedAmount: 10, AmountReceived: 0}
			},
			donationRequest: func(donor, receiver *Client) treasureDonationPayload {
				return treasureDonationPayload{From: donor.id, To: receiver.id, TreasureId: "invalid-treasure-id", DonatedAmount: 10}
			},
			wantDonorMsg: wsmessages.ErrorInvalidTreasureDonateRequest,
		},
		{
			name: "insufficient receiver inventory",
			treasureRequest: func(receiver, donor *Client) (string, treasureRequestState) {
				return receiver.id, treasureRequestState{TreasureId: validTreasureID, RequestedAmount: 10, AmountReceived: 0}
			},
			donationRequest: func(donor, receiver *Client) treasureDonationPayload {
				return treasureDonationPayload{From: donor.id, To: receiver.id, TreasureId: validTreasureID, DonatedAmount: 10}
			},
			receiverSetup: func(t *testing.T, p *Client) {
				p.inventory = makeFakeInventory(t, inventorySize)
			},
			wantDonorMsg: wsmessages.ErrorReceiverInventoryInsufficient,
		},
		{
			name: "insufficient slots in treasure request",
			treasureRequest: func(receiver, donor *Client) (string, treasureRequestState) {
				return receiver.id, treasureRequestState{TreasureId: validTreasureID, RequestedAmount: 10, AmountReceived: 5}
			},
			donationRequest: func(donor, receiver *Client) treasureDonationPayload {
				return treasureDonationPayload{From: donor.id, To: receiver.id, TreasureId: validTreasureID, DonatedAmount: 10}
			},
			wantDonorMsg: wsmessages.ErrorDonationSlotsInsufficient,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			room, _ := initializeFakeRoom(t)
			receiver := initializeFakeClient(t, room, "receiver", false)
			donor := initializeFakeClient(t, room, "donor", false)
			room.Players[receiver] = true
			room.Players[donor] = true

			if tt.treasureRequest != nil {
				id, state := tt.treasureRequest(receiver, donor)
				room.treasureRequests[id] = state
			}

			if tt.receiverSetup != nil {
				tt.receiverSetup(t, receiver)
			}

			donorClient, receiverClient := room.validateDonationRequest(tt.donationRequest(donor, receiver))

			if tt.wantReturns {
				if donorClient != donor || receiverClient != receiver {
					t.Errorf("got %p & %p; want %p & %p", donorClient, receiverClient, donor, receiver)
				}
			} else if donorClient != nil || receiverClient != nil {
				t.Errorf("expected nil, nil on failure; got %p & %p", donorClient, receiverClient)
			}

			if tt.wantDonorMsg == "" {
				expectNoMsg(t, donor)
			} else if msg := recvMsg(t, donor); msg.Type != tt.wantDonorMsg {
				t.Errorf("donor got %v; want %v", msg.Type, tt.wantDonorMsg)
			}

			expectNoMsg(t, donor)
			expectNoMsg(t, receiver)
		})
	}
}
