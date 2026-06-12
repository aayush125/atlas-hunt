package wsmessages

type WSEvent string

const (
	ConnectionEstablished   = WSEvent("connection:established")
	WelcomeToTheGame        = WSEvent("game:welcome")
	PlayerJoined            = WSEvent("player:joined")
	PlayerLeft              = WSEvent("player:left")
	StartGameRequest        = WSEvent("request:game:start")
	ShowGameLoading         = WSEvent("ingame:show:loading")
	GameDataTreasures       = WSEvent("game:data:treasures")
	NotEnoughPlayers        = WSEvent("game:insufficient:players")
	TreasureClaimSuccessful = WSEvent("game:treasureclaim:successful")
	BonusClaimSuccessful    = WSEvent("game:bonusclaim:successful")
	BeginGame               = WSEvent("game:begin")
	MapRoundOver            = WSEvent("game:round:over")
	NextRound               = WSEvent("game:round:next")
	ResumeGameInProgress    = WSEvent("game:resume")
	ResumeGameTransitioning = WSEvent("game:resume:transitioning")
	NewGame                 = WSEvent("game:new")
	ResetGame               = WSEvent("game:reset")
)

type WSMessage struct {
	Type WSEvent        `json:"type"`
	Data map[string]any `json:"data"`
}
