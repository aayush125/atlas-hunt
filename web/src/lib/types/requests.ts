export type TreasureRequestState = {
	treasureId: string;
	requestedAmount: number;
	amountReceived: number;
};

export type TreasureRequests = Record<string, TreasureRequestState>;
