package wsmessages

import "encoding/json"

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
	PlayerClaimedTreasure   = WSEvent("player:claimed:treasure")       // server->client

	RequestTreasure = WSEvent("game:treasure:request") // client->server
	DonateTreasure  = WSEvent("game:treasure:donate")  // client->server

	ErrorDonatingTreasure          = WSEvent("game:treasuredonate:error")                // server->client
	ErrorInvalidTreasureRequest    = WSEvent("game:treasuredonate:invalid:request")      // server->client
	ErrorTreasureRequestNotFound   = WSEvent("game:treasuredonate:request:notfound")     // server->client
	ErrorDonationSlotsInsufficient = WSEvent("game:treasuredonate:slots:insufficient ")  // server->client
	ErrorInsufficientBalance       = WSEvent("game:treasuredonate:insufficient:balance") // server->client
	TreasureRequestAlreadyExists   = WSEvent("game:treasurerequest:exists")              // server->client

	TreasureDonatedSuccessfully  = WSEvent("game:treasuredonate:successful")          // server->client
	TreasureReceivedSuccessfully = WSEvent("game:treasurerequest:receive:successful") // server->client
)

type WSMessageOut struct {
	Type WSEvent        `json:"type"`
	Data map[string]any `json:"data"`
}

type WSMessageIn struct {
	Type WSEvent                    `json:"type"`
	Data map[string]json.RawMessage `json:"data"`
}
