import type { Game } from '$lib/types/game';
import type { Member, Player } from '$lib/types/player';
import type { TreasureItem } from '$lib/types/treasure';
import { writable } from 'svelte/store';

type GamePhase = 'idle' | 'waiting' | 'loading' | 'in-game' | 'game-over';
type GameMap = 'first' | 'second' | 'third';

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
	map: GameMap | null;
	treasures: TreasureItem[];
	points: number;
	treasureCount: Record<string, number>;
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
	map: null,
	treasures: [],
	points: 0,
	treasureCount: {}
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
				map: null,
				treasures: [],
				points: 0,
				treasureCount: {}
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
		setMap: (map: 'first' | 'second' | 'third') => update((game) => ({ ...game, map: map })),
		updatePoints: (points: number) => update((game) => ({ ...game, points: points })),
		updateTreasureCount: (treasureId: string, count: number) =>
			update((game) => ({
				...game,
				treasureCount: {
					...game.treasureCount,
					[treasureId]: count
				}
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
