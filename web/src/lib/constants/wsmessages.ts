export const Events = {
	ConnectionEstablished: 'connection:established',
	WelcomeToTheGame: 'game:welcome',
	PlayerJoined: 'player:joined',
	PlayerLeft: 'player:left',
	StartGameRequest: 'request:game:start',
	ShowGameLoading: 'ingame:show:loading',
	GameDataTreasures: 'game:data:treasures',
	NotEnoughPlayers: 'game:insufficient:players',
	TreasureClaimSuccessful: 'game:treasureclaim:successful',
	BonusClaimSuccessful: 'game:bonusclaim:successful',
	BeginGame: 'game:begin',
	MapRoundOver: 'game:round:over',
	NextRound: 'game:round:next',
	ResumeGame: 'game:resume',
	ResumeGameTransitioning: 'game:resume:transitioning',
	NewGame: 'game:new',
	ResetGame: 'game:reset',
	PlayerClaimedTreasure: 'player:claimed:treasure',
	RequestTreasure: 'game:treasure:request',
	DonateTreasure: 'game:treasure:donate',

	ErrorDonatingTreasure: 'game:treasuredonate:error',
	ErrorInvalidTreasureRequest: 'game:treasuredonate:invalid:request',
	ErrorTreasureRequestNotFound: 'game:treasuredonate:request:notfound',
	ErrorDonationSlotsInsufficient: 'game:treasuredonate:slots:insufficient ',
	ErrorInsufficientBalance: 'game:treasuredonate:insufficient:balance',
	TreasureRequestAlreadyExists: 'game:treasurerequest:exists',

	TreasureDonatedSuccessfully: 'game:treasuredonate:successful',
	TreasureReceivedSuccessfully: 'game:treasurerequest:receive:successful'
} as const;

export type Event = (typeof Events)[keyof typeof Events];
