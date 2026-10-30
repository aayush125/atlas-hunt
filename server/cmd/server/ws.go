package main

import (
	"net/http"

	"atlas.sageyutta.com/internal/cc_context"
	"atlas.sageyutta.com/internal/ws"
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

	conn, err := ws.UpgradeToWebsocket(w, r, app.wsupgrader)
	if err != nil {
		app.logger.Error("error upgrading to websocket", "error", err)
		return
	}

	client := ws.NewClient(conn, room, player.Nickname, player.PlayerID.String(), player.SessionToken, app.logger, player.IsCreator)
	client.PumpsWg.Add(2)
	app.background(client.WritePump)
	app.background(client.ReadPump)

	room.Register <- client
}
