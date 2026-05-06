package main

import (
	"net/http"

	"game.crisis-control.com/internal/cc_context"
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

	claimed, err := playerClient.TryClaimTreasure(treasureInstanceID, input.QuestionID, input.SelectedAnswer)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"claimed": claimed}, nil, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
