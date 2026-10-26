export const Events = {
	ConnectionEstablished: 'connection:established',
	WelcomeToTheGame: 'game:welcome',
	PlayerJoined: 'player:joined',
	PlayerLeft: 'player:left',
	StartGameRequest: 'request:game:start',
	ShowGameLoading: 'ingame:show:loading',
	GameDataTreasures: 'game:data:treasures',
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

	UpdateTreasureRequests: 'game:update:treasurerequests',

	RequestTreasure: 'game:treasure:request',
	DonateTreasure: 'game:treasure:donate',

	NotEnoughPlayers: 'game:insufficient:players',
	ErrorDonatingTreasure: 'game:treasuredonate:error',
	ErrorInvalidTreasureRequest: 'game:treasurerequest:invalid:request',
	ErrorInvalidTreasureDonateRequest: 'game:treasuredonate:invalid:request',
	ErrorTreasureRequestNotFound: 'game:treasuredonate:request:notfound',
	ErrorDonationSlotsInsufficient: 'game:treasuredonate:slots:insufficient ',
	ErrorInsufficientBalance: 'game:treasuredonate:insufficient:balance',
	TreasureRequestAlreadyExists: 'game:treasurerequest:exists',
	ErrorReceiverInventoryInsufficient: 'game:treasuredonate:receiver:inventory:insufficient',
	InsufficientInventorySpace: 'player:insufficient:inventory',

	TreasureDonatedSuccessfully: 'game:treasuredonate:successful',
	TreasureReceivedSuccessfully: 'game:treasurerequest:receive:successful',

	DeleteTreasureRequest: 'game:treasurerequest:delete'
} as const;

export type Event = (typeof Events)[keyof typeof Events];

export const ErrorEventsMap: Partial<Record<Event, string>> = {
	[Events.NotEnoughPlayers]: 'Not enough players to start the game.',
	[Events.ErrorDonatingTreasure]: 'An error occurred while donating treasure.',
	[Events.ErrorInvalidTreasureRequest]: 'The treasure request was invalid.',
	[Events.ErrorInvalidTreasureDonateRequest]: 'The treasure donation was invalid.',
	[Events.ErrorTreasureRequestNotFound]:
		'Treasure request not found. It may have already been satisfied or deleted.',
	[Events.ErrorDonationSlotsInsufficient]:
		'The receiver does not have enough slots for your donation.',
	[Events.ErrorInsufficientBalance]: 'You do not have enough treasures for this donation.',
	[Events.TreasureRequestAlreadyExists]: 'You already have a treasure reqeust.',
	[Events.ErrorReceiverInventoryInsufficient]:
		"The receiver's inventory does not have enough space for your donation.",
	[Events.InsufficientInventorySpace]: 'You do not have enough space in your inventory.'
};
