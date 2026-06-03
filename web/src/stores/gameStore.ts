import type { Game, GameMap, MapRoundOverStats } from '$lib/types/game';
import type { Member, Player } from '$lib/types/player';
import type { TreasureItem } from '$lib/types/treasure';
import { derived, writable } from 'svelte/store';

type GamePhase = 'idle' | 'waiting' | 'loading' | 'in-game' | 'game-over' | 'transitioning';

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
	treasureCountThisRound: number | null;
	serverNow: number | null;
	roundEndsAt: number | null;
	mapRoundStats: Record<GameMap, MapRoundOverStats | null>;
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
	treasureCount: {},
	treasureCountThisRound: null,
	serverNow: null,
	roundEndsAt: null,
	mapRoundStats: {
		first: null,
		second: null,
		third: null
	}
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
				treasureCount: {},
				treasureCountThisRound: 0,
				serverNow: null,
				roundEndsAt: null,
				mapRoundStats: {
					first: null,
					second: null,
					third: null
				}
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
		setMembers: (members: Member[]) => update((game) => ({ ...game, otherPlayers: members })),
		setInventory: (inventory: TreasureItem[]) =>
			update((game) => ({ ...game, treasures: inventory })),
		setTreasureCount: (treasure_count: Record<string, number>) =>
			update((game) => ({ ...game, treasureCount: treasure_count })),
		deleteTreasureCount: (treasureId: string) =>
			update((game) => {
				const { [treasureId]: _, ...remainingCounts } = game.treasureCount;
				return {
					...game,
					treasureCount: remainingCounts
				};
			}),
		deleteTreasure: (instanceId: string) =>
			update((game) => ({
				...game,
				treasures: game.treasures.filter((item) => item.instanceId !== instanceId)
			})),
		setGameInfo: (code: string, max_players: number, current_players: number) =>
			update((game) => ({
				...game,
				joinCode: code,
				maxPlayers: max_players,
				currentPlayerCount: current_players
			})),
		setPlayerInfo: (nickname: string, id: string, points: number, is_creator: boolean) =>
			update((game) => ({
				...game,
				playerId: id,
				playerName: nickname,
				points: points,
				playerIsCreator: is_creator
			})),
		removeTreasureFromInventory: (treasureId: string) =>
			update((game) => ({
				...game,
				treasures: game.treasures.filter((t) => t.id !== treasureId)
			})),
		setMapRoundStats: (map: GameMap, stats: MapRoundOverStats) =>
			update((game) => ({ ...game, mapRoundStats: { ...game.mapRoundStats, [map]: stats } })),
		getMapRoundStats: (map: GameMap) =>
			derived({ subscribe }, ($gameStore) => $gameStore.mapRoundStats[map]),
		startRound: (serverNow: number, endsAt: number) =>
			update((game) => ({ ...game, serverNow: serverNow, roundEndsAt: endsAt })),
		updateRoundTreasureCount: (count: number) =>
			update((game) => ({ ...game, treasureCountThisRound: count })),
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
