import { writable } from 'svelte/store';

type GamePhase = 'idle' | 'waiting' | 'in-game' | 'game-over';

interface GameState {
	joinCode: string | null;
	numberOfPlayers: number;
	playerName: string | null;
	state: GamePhase;
	createdAt: Date | null;
}

const initial: GameState = {
	joinCode: null,
	numberOfPlayers: 0,
	playerName: null,
	state: 'idle',
	createdAt: null
};

function createGameStore() {
	const { subscribe, update, set } = writable<GameState>(initial);

	return {
		subscribe,
		setupGame: (code: string, numPlayers: number, nickname: string, createdAt: Date) =>
			update((game) => ({
				...game,
				joinCode: code,
				numberOfPlayers: numPlayers,
				playerName: nickname,
				createdAt: createdAt,
				state: 'waiting'
			})),
		setGameState: (state: GamePhase) => update((game) => ({ ...game, state })),
		reset: () => set(initial)
	};
}

export const gameStore = createGameStore();
export type { GameState, GamePhase };
