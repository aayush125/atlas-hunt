package wsmessages

type WSEvent string

const (
	ConnectionEstablished = WSEvent("connection:established")
	WelcomeToTheGame      = WSEvent("game:welcome")
	PlayerJoined          = WSEvent("player:joined")
)

type WSMessage struct {
	Type WSEvent        `json:"type"`
	Data map[string]any `json:"data"`
}
