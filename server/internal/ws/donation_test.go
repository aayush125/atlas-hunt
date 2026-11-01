package ws

import (
	"testing"

	"atlas.sageyutta.com/internal/gamedata"
	"atlas.sageyutta.com/internal/wsmessages"
)

func TestTreasureRequestValidation(t *testing.T) {
	validTreasureID := gamedata.GetAllTreasures()[0].ID

	tests := []struct {
		name        string
		setup       func(t *testing.T, p *Client)
		request     func(p *Client) treasureRequestPayload
		wantClient  bool
		wantMsgType wsmessages.WSEvent
	}{
		{
			name: "valid treasure",
			request: func(p *Client) treasureRequestPayload {
				return treasureRequestPayload{PlayerId: p.id, TreasureId: validTreasureID, RequestedAmount: 10}
			},
			wantClient: true,
		},
		{
			name: "invalid player",
			request: func(p *Client) treasureRequestPayload {
				return treasureRequestPayload{PlayerId: "non-existent-player", TreasureId: validTreasureID, RequestedAmount: 10}
			},
		},
		{
			name: "invalid treasure id",
			request: func(p *Client) treasureRequestPayload {
				return treasureRequestPayload{PlayerId: p.id, TreasureId: "fk10", RequestedAmount: 10}
			},
			wantMsgType: wsmessages.ErrorInvalidTreasureRequest,
		},
		{
			name: "empty treasure id",
			request: func(p *Client) treasureRequestPayload {
				return treasureRequestPayload{PlayerId: p.id, TreasureId: "", RequestedAmount: 10}
			},
			wantMsgType: wsmessages.ErrorInvalidTreasureRequest,
		},
		{
			name: "insufficient inventory space",
			setup: func(t *testing.T, p *Client) {
				p.inventory = makeFakeInventory(t, inventorySize)
			},
			request: func(p *Client) treasureRequestPayload {
				return treasureRequestPayload{PlayerId: p.id, TreasureId: validTreasureID, RequestedAmount: 1}
			},
			wantMsgType: wsmessages.InsufficientInventorySpace,
		},
		{
			name: "negative amount",
			request: func(p *Client) treasureRequestPayload {
				return treasureRequestPayload{PlayerId: p.id, TreasureId: validTreasureID, RequestedAmount: -1}
			},
			wantMsgType: wsmessages.ErrorInvalidTreasureRequest,
		},
		{
			name: "zero amount",
			request: func(p *Client) treasureRequestPayload {
				return treasureRequestPayload{PlayerId: p.id, TreasureId: validTreasureID, RequestedAmount: 0}
			},
			wantMsgType: wsmessages.ErrorInvalidTreasureRequest,
		},
		{
			name: "one treasure requested for one free inventory slot",
			setup: func(t *testing.T, p *Client) {
				p.inventory = makeFakeInventory(t, inventorySize-1)
			},
			request: func(p *Client) treasureRequestPayload {
				return treasureRequestPayload{PlayerId: p.id, TreasureId: validTreasureID, RequestedAmount: 1}
			},
			wantClient: true,
		},
		{
			name: "two treasures requested for one free inventory slot",
			setup: func(t *testing.T, p *Client) {
				p.inventory = makeFakeInventory(t, inventorySize-1)
			},
			request: func(p *Client) treasureRequestPayload {
				return treasureRequestPayload{PlayerId: p.id, TreasureId: validTreasureID, RequestedAmount: 2}
			},
			wantMsgType: wsmessages.InsufficientInventorySpace,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			room, _ := initializeFakeRoom(t)
			player := initializeFakeClient(t, room, "player", false)
			room.Players[player] = true

			if tt.setup != nil {
				tt.setup(t, player)
			}

			got := room.validateTreasureRequest(tt.request(player))

			if (got == player) != tt.wantClient {
				t.Errorf("validateTreasureRequest returned %v; wantClient=%v", got, tt.wantClient)
			}

			if tt.wantMsgType == "" {
				expectNoMsg(t, player)
			} else if msg := recvMsg(t, player); msg.Type != tt.wantMsgType {
				t.Errorf("player got %v; want %v", msg.Type, tt.wantMsgType)
			}

			expectNoMsg(t, player)
		})
	}
}
