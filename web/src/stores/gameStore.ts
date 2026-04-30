import type { Game } from '$lib/types/game';
import type { Member, Player } from '$lib/types/player';
import type { TreasureItem } from '$lib/types/treasure';
import { writable } from 'svelte/store';

type GamePhase = 'idle' | 'waiting' | 'in-game' | 'game-over';
type GameRound = 'one' | 'two' | 'three';

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
	round: GameRound | null;
	treasures: TreasureItem[];
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
	otherPlayers: [],
	round: null,
	treasures: []
};

function createGameStore() {
	const { subscribe, update, set } = writable<GameState>(initial);

	return {
		subscribe,
		setupGame: (game: Game, player: Player, members: Member[]) =>
			update(() => ({
				joinCode: game.id,
				maxPlayers: game.max_players,
				playerName: player.nickname,
				playerId: player.id,
				state: 'waiting',
				createdAt: new Date(game.created_at),
				playerIsCreator: player.is_creator,
				currentPlayerCount: game.current_player_count,
				otherPlayers: members,
				round: null,
				treasures: []
			})),
		setGameState: (state: GamePhase) => update((game) => ({ ...game, state })),
		addPlayer: (id: string, nickname: string) =>
			update((game) => ({
				...game,
				otherPlayers: [...game.otherPlayers, { id, nickname }]
			})),
		addTreasure: (treasure: TreasureItem) =>
			update((game) => ({
				...game,
				treasures: [...game.treasures, treasure]
			})),
		removePlayer: (id: string) =>
			update((game) => ({
				...game,
				otherPlayers: game.otherPlayers.filter((member) => member.id !== id)
			})),
		reset: () => set(initial)
	};
}

export const gameStore = createGameStore();
export type { GamePhase, GameState };
