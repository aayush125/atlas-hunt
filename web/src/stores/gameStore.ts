import type { Game } from '$lib/types/game';
import type { Member, Player } from '$lib/types/player';
import { writable } from 'svelte/store';

type GamePhase = 'idle' | 'waiting' | 'in-game' | 'game-over';

interface GameState {
	joinCode: string | null;
	maxPlayers: number;
	playerName: string | null;
	playerId: string | null;
	state: GamePhase;
	createdAt: Date | null;
	playerIsCreator: boolean;
	currentPlayerCount: number;
	otherPlayers: Member[];
}

const initial: GameState = {
	joinCode: null,
	maxPlayers: 0,
	playerName: null,
	playerId: null,
	state: 'idle',
	createdAt: null,
	playerIsCreator: false,
	currentPlayerCount: 0,
	otherPlayers: []
};

function createGameStore() {
	const { subscribe, update, set } = writable<GameState>(initial);

	return {
		subscribe,
		setupGame: (game: Game, player: Player) =>
			update(() => ({
				joinCode: game.id,
				maxPlayers: game.max_players,
				playerName: player.nickname,
				playerId: player.id,
				state: 'waiting',
				createdAt: new Date(game.created_at),
				playerIsCreator: player.is_creator,
				currentPlayerCount: game.current_player_count,
				otherPlayers: []
			})),
		setGameState: (state: GamePhase) => update((game) => ({ ...game, state })),
		addPlayer: (id: string, nickname: string) =>
			update((game) => ({
				...game,
				otherPlayers: [...game.otherPlayers, { id, nickname }]
			})),
		reset: () => set(initial)
	};
}

export const gameStore = createGameStore();
export type { GameState, GamePhase };
