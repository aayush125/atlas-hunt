package main

import (
	"fmt"
	"net/http"

	"game.crisis-control.com/internal/cc_context"
)

func (app *application) validateTokenHandler(w http.ResponseWriter, r *http.Request) {
	player := cc_context.ContextGetPlayer(r)

	game, err := app.models.Game.GetForCode(player.GameID)
	if err != nil {
		app.logger.Info("error getting room from redis")
		app.serverErrorResponse(w, r, err)
		return
	}

	app.mu.Lock()
	room, ok := app.gameRooms[player.GameID]
	app.mu.Unlock()

	if !ok {
		app.logger.Info("error getting room from map")
		app.roomNotFoundResponse(w, r)
		return
	}

	playerClient := room.GetPlayerClientWithSessionToken(player.SessionToken)
	if playerClient == nil {
		app.logger.Info("player was nil during token validation")
		app.customErrorResponse(w, r, http.StatusForbidden, fmt.Errorf("reconnection not allowed"))
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"code": player.GameID, "player_count": game.CurrentPlayerCount, "nickname": player.Nickname}, nil, nil)
	if err != nil {
		app.logger.Info("error writing json")
		app.serverErrorResponse(w, r, err)
	}
}
