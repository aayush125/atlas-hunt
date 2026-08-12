package wsmessages

type WSEvent string

const (
	ConnectionEstablished   = WSEvent("connection:established")        // server->client
	WelcomeToTheGame        = WSEvent("game:welcome")                  // server->client
	PlayerJoined            = WSEvent("player:joined")                 // server->client
	PlayerLeft              = WSEvent("player:left")                   // server->client
	StartGameRequest        = WSEvent("request:game:start")            // client->server
	ShowGameLoading         = WSEvent("ingame:show:loading")           // server->client
	GameDataTreasures       = WSEvent("game:data:treasures")           // server->client
	NotEnoughPlayers        = WSEvent("game:insufficient:players")     // server->client
	TreasureClaimSuccessful = WSEvent("game:treasureclaim:successful") // server->client
	BonusClaimSuccessful    = WSEvent("game:bonusclaim:successful")    // server->client
	BeginGame               = WSEvent("game:begin")                    // server->client
	MapRoundOver            = WSEvent("game:round:over")               // server->client
	NextRound               = WSEvent("game:round:next")               // client->server
	ResumeGameInProgress    = WSEvent("game:resume")                   // server->client
	ResumeGameTransitioning = WSEvent("game:resume:transitioning")     // server->client
	NewGame                 = WSEvent("game:new")                      // client->server
	ResetGame               = WSEvent("game:reset")                    // server->client
)

type WSMessage struct {
	Type WSEvent        `json:"type"`
	Data map[string]any `json:"data"`
}
