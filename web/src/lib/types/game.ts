type GAMESTATE = 'lobby' | 'in_progress' | 'transitioning';

export type GameMap = 'first' | 'second' | 'third';

export type Game = {
	max_players: number;
	current_player_count: number;
	id: string;
	created_at: Date;
	creator_nickname: string;
	state: GAMESTATE;
};

export type MapRoundOverStats = {
	map: GameMap;
	treasures_claimed: number;
	treasures_failed: number;
	all_player_points: Record<string, number>;
};
