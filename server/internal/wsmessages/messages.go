package wsmessages

type WSEvent string

const (
	ConnectionEstablished = WSEvent("connection:established")
)

type WSMessage struct {
	Type WSEvent        `json:"type"`
	Data map[string]any `json:"data"`
}
