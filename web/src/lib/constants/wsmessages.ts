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
	ResetGame: 'game:reset'
} as const;

export type Event = (typeof Events)[keyof typeof Events];
