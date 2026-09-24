package ws

import (
	"log/slog"
	"time"

	"atlas.sageyutta.com/internal/data"
	"atlas.sageyutta.com/internal/engine"
	"atlas.sageyutta.com/internal/gamedata"
	"atlas.sageyutta.com/internal/wsmessages"
)

type playerLookupRequest struct {
	session string
	result  chan *Client
}

type playerNotification struct {
	Client  *Client
	Message wsmessages.WSMessage
	Info    map[string]any
}

type roundStartData struct {
	roundTime int64
	roundMap  data.GameMap
}

type Player interface {
	Delete(session string) error
	BackupPlayerStats(token string, treasureSet []gamedata.TreasureItem, treasureCount map[string]int, inventory []gamedata.InventoryItem, points int) error
	UpdateTreasureSet(token string, treasureSet []gamedata.TreasureItem) error
	SetTTL(token string) error
}

type Game interface {
	GetAllPlayers(gameCode string) ([]string, error)
	CleanupGame(gameCode string) error
	IncrementPlayerCount(gameCode string) error
	DecrementPlayerCount(gameCode string) error
	RemovePlayerFromSet(gameCode, playerSessionToken string) error
	SetGameAndMapState(code string, state data.GameState, gameMap data.GameMap) error
	SetTTL(code string) error
	SetTTLForPlayerSet(gameCode string, ttl time.Duration) error
}

type Room struct {
	Players               map[*Client]bool
	Register              chan *Client
	Unregister            chan *Client
	StartGame             chan *Client
	NextRound             chan *Client
	NewGame               chan *Client
	Shutdown              chan struct{}
	ResumeGame            chan *Client
	GameModel             Game
	PlayerModel           Player
	Code                  string
	State                 data.GameState
	lastMap               data.GameMap
	currentMap            data.GameMap
	logger                *slog.Logger
	maxPlayers            int
	Broadcast             chan *wsmessages.WSMessage
	engine                *engine.Engine
	PlayerLookup          chan playerLookupRequest
	NotifyTreasureClaimed chan playerNotification
	NotifyBonusClaimed    chan playerNotification
	timer                 *time.Timer
	resetTimer            chan roundStartData
	backupTimer           *time.Timer
	roundEndTime          int64
	onExit                func(code string)
	roundTime             time.Duration
	backupTime            time.Duration
	ttlResetPeriod        time.Duration
}
