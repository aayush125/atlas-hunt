package main

import (
	"errors"
	"net/http"

	"game.crisis-control.com/internal/cc_context"
	"game.crisis-control.com/internal/ws"
)

func (app *application) claimBonusPointsHandler(w http.ResponseWriter, r *http.Request) {
	player := cc_context.ContextGetPlayer(r)

	var input struct {
		TreasureID string `json:"treasure_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.TreasureID == "" {
		app.badRequestResponse(w, r, errors.New("treasure_id is required"))
		return
	}

	room, ok := app.gameRooms[player.GameID]
	if !ok {
		app.roomNotFoundResponse(w, r)
		return
	}

	playerClient := room.GetPlayerClientWithSessionToken(player.SessionToken)
	if playerClient == nil {
		app.serverErrorResponse(w, r, errors.New("player client not found"))
		return
	}

	err = playerClient.TryClaimBonusPoints(input.TreasureID)
	if err != nil {
		switch {
		case errors.Is(err, ws.ErrPlayerDoesNotHaveRequestedTreasure):
			app.treasureNotFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"claimed": true}, nil, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
