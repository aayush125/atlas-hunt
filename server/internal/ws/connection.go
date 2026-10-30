package ws

import (
	"net/http"

	"github.com/gorilla/websocket"
)

func NewUpgrader(origin string) websocket.Upgrader {
	return websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return r.Header.Get("Origin") == origin
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
}

func UpgradeToWebsocket(w http.ResponseWriter, r *http.Request, upgrader websocket.Upgrader) (*websocket.Conn, error) {
	return upgrader.Upgrade(w, r, nil)
}
