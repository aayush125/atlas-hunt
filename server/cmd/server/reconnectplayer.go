package main

import (
	"fmt"
	"net/http"

	"atlas.sageyutta.com/internal/cc_context"
	"atlas.sageyutta.com/internal/ws"
)

func (app *application) reconnectPlayerHandler(w http.ResponseWriter, r *http.Request) {
	player := cc_context.ContextGetPlayer(r)

	app.mu.Lock()
	room, ok := app.gameRooms[player.GameID]
	app.mu.Unlock()

	if !ok {
		app.roomNotFoundResponse(w, r)
		return
	}

	conn, err := ws.UpgradeToWebsocket(w, r, app.wsupgrader)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	playerClient := room.GetPlayerClientWithSessionToken(player.SessionToken)
	if playerClient == nil {
		app.customErrorResponse(w, r, http.StatusForbidden, fmt.Errorf("cannot rejoin"))
		return
	}

	playerClient.PumpsWg.Wait()
	playerClient.Instantiate(conn)
	playerClient.PumpsWg.Add(2)
	app.background(playerClient.WritePump)
	app.background(playerClient.ReadPump)

	room.Register <- playerClient
	room.ResumeGame <- playerClient
}
