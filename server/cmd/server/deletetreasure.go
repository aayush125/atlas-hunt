package main

import (
	"errors"
	"net/http"

	"game.crisis-control.com/internal/cc_context"
)

func (app *application) deleteTreasureHandler(w http.ResponseWriter, r *http.Request) {
	player := cc_context.ContextGetPlayer(r)

	var input struct {
		InstanceID string `json:"instanceId"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
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

	playerClient.DeleteTreasure(input.InstanceID)

	err = app.writeJSON(w, http.StatusOK, envelope{"instanceId": input.InstanceID}, nil, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
