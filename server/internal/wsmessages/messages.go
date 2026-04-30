package wsmessages

type WSEvent string

const (
	ConnectionEstablished = WSEvent("connection:established")
	WelcomeToTheGame      = WSEvent("game:welcome")
	PlayerJoined          = WSEvent("player:joined")
	PlayerLeft            = WSEvent("player:left")
	StartGameRequest      = WSEvent("request:game:start")

	// temporary
	AcknowledgeStartRequest = WSEvent("ack:game:start")
)

type WSMessage struct {
	Type WSEvent        `json:"type"`
	Data map[string]any `json:"data"`
}
