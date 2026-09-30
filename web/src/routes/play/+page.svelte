<script lang="ts">
	import { ErrorEventsMap, Events } from '$lib/constants/wsmessages';
	import type { Game, GameMap, MapRoundOverStats } from '$lib/types/game';
	import type { Member, Player } from '$lib/types/player';
	import type { Treasure, TreasureItem } from '$lib/types/treasure';
	import type { WSMessage } from '$lib/types/wsmessage';
	import CreateGame from '../../components/CreateGame.svelte';
	import GameLoading from '../../components/ingame/GameLoading.svelte';
	import JoinGame from '../../components/JoinGame.svelte';
	import Lobby from '../../components/Lobby.svelte';
	import GamePlay from '../../components/ingame/GamePlay.svelte';
	import { gameStore } from '../../stores/gameStore';
	import { onDestroy } from 'svelte';
	import Transitioning from '../../components/ingame/Transitioning.svelte';
	import { PUBLIC_WS_BASE } from '$env/static/public';
	import type { PageProps } from './$types';
	import Rejoin from '../../components/Rejoin.svelte';
	import { toast } from '../../stores/toast';
	import { getPlayerClaimedTreasureToastMessage } from '$lib/helpers';

	let { data }: PageProps = $props();

	let reconnectInfo = (() => $state.snapshot(data.reconnectInfo))();

	let gate = $state<'join' | 'create'>('join');
	let rejoinPromptDismissed = $state(false);

	const MAX_RETRIES = 5;
	const BASE_RETRY_DELAY_MS = 1000;

	let ws: WebSocket | null = null;
	let retryCount = $state(0);
	let retryTimeout: ReturnType<typeof setTimeout> | null = null;
	let manualClose = false;

	let mapTreasures = $state<Treasure[]>([]);
	let wsReady = $state(false);
	let connectionState = $state<'connected' | 'reconnecting' | 'failed'>('connected');

	function connectToWs(reconnecting: boolean = false) {
		if (ws && (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING)) return;

		manualClose = false;
		ws = new WebSocket(`${PUBLIC_WS_BASE}/${reconnecting ? 'game/reconnect' : 'ws'}`);

		ws.onopen = () => {
			console.log('WebSocket connected');
			wsReady = true;
			connectionState = 'connected';
			retryCount = 0;
		};

		ws.onmessage = (event) => {
			try {
				const data = JSON.parse(event.data);
				const message: WSMessage = {
					type: data.type,
					data: data.data
				};
				handleMessage(message);
			} catch {
				console.warn('Received non-JSON message:', event.data);
			}
		};

		ws.onclose = (event) => {
			console.log(`WebSocket closed (code: ${event.code})`);
			wsReady = false;
			if (!manualClose && retryCount < MAX_RETRIES) {
				connectionState = 'reconnecting';
				const delay = BASE_RETRY_DELAY_MS * 2 ** retryCount;
				console.log(`Reconnecting in ${delay}ms (attempt ${retryCount + 1}/${MAX_RETRIES})...`);
				retryTimeout = setTimeout(() => {
					retryCount++;
					connectToWs(true);
				}, delay);
			} else if (retryCount >= MAX_RETRIES) {
				connectionState = 'failed';
				console.error('Max reconnection attempts reached.');
			}
		};

		ws.onerror = (error) => {
			console.error('WebSocket error:', error);
			wsReady = false;
			ws?.close();
		};
	}

	function handleMessage(data: WSMessage) {
		console.log('Message received:', data);

		const errorMessage = ErrorEventsMap[data.type];
		if (errorMessage) {
			toast(errorMessage, 'error');
			return;
		}

		switch (data.type) {
			case Events.PlayerJoined: {
				const info = data.data as { id: string; nickname: string };
				if (info.id !== $gameStore.playerId) {
					gameStore.addPlayer(info.id, info.nickname);
				}
				toast(`${info.nickname} has joined the game`, 'info');
				break;
			}
			case Events.PlayerLeft: {
				const info = data.data as { id: string };
				toast(`${gameStore.getPlayerNicknameFromId(info.id)} has left the game`, 'info');
				gameStore.removePlayer(info.id);
				break;
			}
			case Events.ShowGameLoading: {
				gameStore.setGameState('loading');
				break;
			}
			case Events.PlayerClaimedTreasure: {
				const info = data.data as { claimer: string; rarity: string };
				toast(getPlayerClaimedTreasureToastMessage(info.claimer, info.rarity), 'info');
				break;
			}
			case Events.GameDataTreasures: {
				const info = data.data as { treasures: Treasure[]; map: 'first' | 'second' | 'third' };
				mapTreasures = info.treasures;
				gameStore.updateRoundTreasureCount(mapTreasures.length);
				gameStore.setMap(info.map);
				break;
			}
			case Events.ResetGame: {
				gameStore.newGame();
				break;
			}
			case Events.BeginGame: {
				const info = data.data as { endsAt: number; serverNow: number };
				gameStore.startRound(info.serverNow, info.endsAt);
				gameStore.setGameState('in-game');
				break;
			}
			case Events.ResumeGameTransitioning: {
				const info = data.data as {
					map: GameMap;
					treasures_claimed: number;
					treasures_failed: number;
					all_player_points: Record<string, number>;
					inventory: TreasureItem[];
					treasure_count: Record<string, number>;
					nickname: string;
					player_id: string;
					points: number;
					members: Member[];
					is_creator: boolean;
					game_info: { code: string; max_players: number; current_players: number };
				};
				gameStore.setPlayerInfo(info.nickname, info.player_id, info.points, info.is_creator);
				gameStore.updateRoundTreasureCount(40);
				gameStore.setGameInfo(
					info.game_info.code,
					info.game_info.max_players,
					info.game_info.current_players
				);
				gameStore.setTreasureCount(info.treasure_count);
				gameStore.setInventory(info.inventory);
				gameStore.setMembers(info.members);
				gameStore.setMap(info.map);

				const stats: MapRoundOverStats = {
					map: info.map,
					treasures_claimed: info.treasures_claimed,
					treasures_failed: info.treasures_failed,
					all_player_points: info.all_player_points
				};
				gameStore.setMapRoundStats(stats.map, stats);
				if ($gameStore.playerId)
					gameStore.updatePoints(stats.all_player_points[$gameStore.playerId]);
				gameStore.setGameState('transitioning');

				rejoinPromptDismissed = true;
				break;
			}
			case Events.ResumeGame: {
				const info = data.data as {
					map: GameMap;
					inventory: TreasureItem[];
					treasure_set: Treasure[];
					treasure_count: Record<string, number>;
					nickname: string;
					player_id: string;
					points: number;
					members: Member[];
					is_creator: boolean;
					server_now: number;
					ends_at: number;
					game_info: { code: string; max_players: number; current_players: number };
				};
				gameStore.setPlayerInfo(info.nickname, info.player_id, info.points, info.is_creator);
				gameStore.setGameInfo(
					info.game_info.code,
					info.game_info.max_players,
					info.game_info.current_players
				);
				mapTreasures = info.treasure_set;
				gameStore.updateRoundTreasureCount(mapTreasures.length);
				gameStore.setTreasureCount(info.treasure_count);
				gameStore.setInventory(info.inventory);
				gameStore.setMembers(info.members);
				gameStore.setMap(info.map);
				gameStore.startRound(info.server_now, info.ends_at);
				gameStore.setGameState('in-game');
				rejoinPromptDismissed = true;
				break;
			}
			case Events.TreasureClaimSuccessful: {
				const info = data.data as {
					instance_id: string;
					treasure_id: string;
					treasure_count: number;
					points: number;
				};
				console.log('received stat updates:', info);
				gameStore.updatePoints(info.points);
				gameStore.updateTreasureCount(info.treasure_id, info.treasure_count);
				break;
			}
			case Events.BonusClaimSuccessful: {
				const info = data.data as {
					points: number;
					treasure_id: string;
				};
				console.log('received bonus claimed update:', info);
				gameStore.updatePoints(info.points);
				gameStore.deleteTreasureCount(info.treasure_id);
				gameStore.removeTreasureFromInventory(info.treasure_id);
				break;
			}
			case Events.MapRoundOver: {
				const stats = data.data as MapRoundOverStats;
				gameStore.setMapRoundStats(stats.map, stats);
				console.log('gamestate:', $gameStore.mapRoundStats);
				if ($gameStore.playerId)
					gameStore.updatePoints(stats.all_player_points[$gameStore.playerId]);
				gameStore.setGameState('transitioning');
				console.log('Received map round over message');
				break;
			}
			default: {
				console.warn('Unhandled message type:', data.type);
			}
		}
	}

	function sendMessage(msg: WSMessage) {
		if (ws?.readyState === WebSocket.OPEN) {
			ws.send(JSON.stringify(msg));
		} else {
			console.warn('WebSocket not open, cannot send:', msg);
		}
	}

	function disconnectWs() {
		manualClose = true;
		if (retryTimeout) clearTimeout(retryTimeout);
		ws?.close(1000, 'User left game');
		ws = null;
	}

	async function leaveGameHandler() {
		const res = await fetch('/api/game/leave', { method: 'POST' });

		if (!res.ok) {
			console.log('failed to logout');
		}

		disconnectWs();
		gameStore.reset();
		gate = 'join';
	}

	onDestroy(disconnectWs);
</script>

<svelte:head>
	<title>Play - Atlas Hunt</title>
	<meta
		name="description"
		content="Join an existing game or create a new one and invite your friends to play."
	/>
	<meta name="robots" content="noindex, nofollow" />
	<meta property="og:title" content="Play - Atlas Hunt" />
	<meta
		property="og:description"
		content="Join an existing game or create a new one and invite your friends to play."
	/>
	<meta property="og:type" content="website" />
</svelte:head>

{#if connectionState === 'reconnecting'}
	<div
		class="fixed top-0 right-0 left-0 z-50 bg-yellow-500 px-4 py-2 text-center text-sm font-medium text-white"
	>
		Connection lost — reconnecting... (attempt {retryCount}/{MAX_RETRIES})
	</div>
{:else if connectionState === 'failed'}
	<div
		class="fixed top-0 right-0 left-0 z-50 bg-red-600 px-4 py-2 text-center text-sm font-medium text-white"
	>
		Connection lost. Please refresh the page.
	</div>
{/if}

<div class="flex flex-col items-center justify-center gap-8 p-8">
	{#if !rejoinPromptDismissed && reconnectInfo !== null}
		<Rejoin
			{reconnectInfo}
			onLeaveGame={() => (rejoinPromptDismissed = true)}
			onRejoin={() => {
				connectToWs(true);
			}}
		/>
	{:else if $gameStore.state === 'waiting'}
		<Lobby
			{wsReady}
			onStartGame={async () => {
				sendMessage({ type: 'request:game:start', data: null });
			}}
			onLeaveGame={leaveGameHandler}
		/>
	{:else if $gameStore.state === 'loading'}
		<GameLoading />
	{:else if $gameStore.state === 'idle'}
		{#if gate === 'join'}
			<!-- <JoinGame
				onCreateGame={() => (gate = 'create')}
				onGameJoined={(player: Player, game: Game, members: Member[]) => {
					connectToWs();
					gameStore.setupGame(game, player, members);
				}}
			/> -->
			<GamePlay
				treasures={mapTreasures}
				onClaim={(t) => {
					mapTreasures = mapTreasures.filter((tr) => tr.instanceId !== t.instanceId);
					gameStore.addTreasure(t);
				}}
				onFail={(t) => {
					mapTreasures = mapTreasures.filter((tr) => tr.instanceId !== t.instanceId);
				}}
				onLeaveGame={leaveGameHandler}
			/>
		{:else}
			<CreateGame
				onJoinGame={() => (gate = 'join')}
				onGameCreated={(game: Game, player: Player) => {
					connectToWs();
					gameStore.setupGame(game, player, []);
				}}
			/>
		{/if}
	{:else if $gameStore.state === 'in-game'}
		<GamePlay
			treasures={mapTreasures}
			onClaim={(t) => {
				mapTreasures = mapTreasures.filter((tr) => tr.instanceId !== t.instanceId);
				gameStore.addTreasure(t);
			}}
			onFail={(t) => {
				mapTreasures = mapTreasures.filter((tr) => tr.instanceId !== t.instanceId);
			}}
			onLeaveGame={leaveGameHandler}
		/>
	{:else if $gameStore.state === 'transitioning'}
		<Transitioning
			onNextRound={$gameStore.playerIsCreator
				? () => {
						sendMessage({ type: Events.NextRound, data: null });
					}
				: undefined}
			onNewGame={$gameStore.playerIsCreator
				? () => {
						sendMessage({ type: Events.NewGame, data: null });
					}
				: undefined}
		/>
	{/if}
</div>
