package main

import (
	"net/http"

	"game.crisis-control.com/internal/cc_context"
	"game.crisis-control.com/internal/ws"
)

func (app *application) websocketHandler(w http.ResponseWriter, r *http.Request) {
	player := cc_context.ContextGetPlayer(r)

	app.mu.Lock()
	room, ok := app.gameRooms[player.GameID]
	app.mu.Unlock()

	if !ok {
		app.roomNotFoundResponse(w, r)
		return
	}

	conn, err := ws.UpgradeToWebsocket(w, r)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	client := ws.NewClient(conn, room, player.Nickname, player.PlayerID.String(), player.SessionToken, app.logger)
	app.background(client.WritePump)
	app.background(client.ReadPump)

	room.Register <- client
}
