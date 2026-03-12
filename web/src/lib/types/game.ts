type GAMESTATE = 'lobby' | 'in_progress';

export type Game = {
	max_players: number;
	current_player_count: number;
	id: string;
	created_at: Date;
	creator_nickname: string;
	state: GAMESTATE;
};
