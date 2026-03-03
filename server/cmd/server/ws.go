package main

import (
	"net/http"

	"game.crisis-control.com/internal/ws"
	"game.crisis-control.com/internal/wsmessages"
)

func (app *application) websocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.UpgradeToWebsocket(w, r)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	client := ws.NewClient(conn, app.logger)

	app.background(client.WritePump)
	app.background(client.ReadPump)

	// notify client that the websocket connection has been established
	err = client.Send(wsmessages.WSMessage{Type: wsmessages.ConnectionEstablished})
	if err != nil {
		app.logger.Error("error sending wsmessages.ConnectionEstablished", "error", err.Error())
	}
}
