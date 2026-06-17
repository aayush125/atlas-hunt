package main

import (
	"errors"
	"fmt"
	"net/http"

	"game.crisis-control.com/internal/cc_context"
	"game.crisis-control.com/internal/ws"
	"github.com/google/uuid"
)

func (app *application) checkTriviaAnswer(w http.ResponseWriter, r *http.Request) {
	player := cc_context.ContextGetPlayer(r)

	var input struct {
		InstanceID     string `json:"instanceId"`
		QuestionID     string `json:"questionId"`
		SelectedAnswer string `json:"selectedAnswer"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	treasureInstanceID, err := uuid.Parse(input.InstanceID)
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

	claimed, err := playerClient.TryClaimTreasure(treasureInstanceID, input.QuestionID, input.SelectedAnswer)
	if err != nil {
		switch {
		case errors.Is(err, ws.ErrInventoryFull):
			app.customErrorResponse(w, r, http.StatusConflict, fmt.Errorf("Inventory full"))
		case errors.Is(err, ws.ErrTreasureNotFound):
			app.customErrorResponse(w, r, http.StatusNotFound, fmt.Errorf("Treasure not found"))
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"claimed": claimed}, nil, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
