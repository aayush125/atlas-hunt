export type Player = {
	nickname: string;
	game_id: string;
	join_timestamp: Date;
	last_rejoin_timestamp: Date;
	id: string;
	is_creator: boolean;
};

export type Member = {
	nickname: string;
	id: string;
};
