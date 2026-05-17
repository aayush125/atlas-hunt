<script lang="ts">
	import { Events } from '$lib/constants/wsmessages';
	import type { Game, MapRoundOverStats } from '$lib/types/game';
	import type { Member, Player } from '$lib/types/player';
	import type { Treasure } from '$lib/types/treasure';
	import type { WSMessage } from '$lib/types/wsmessage';
	import CreateGame from '../../components/CreateGame.svelte';
	import GameLoading from '../../components/ingame/GameLoading.svelte';
	import JoinGame from '../../components/JoinGame.svelte';
	import Lobby from '../../components/Lobby.svelte';
	import GamePlay from '../../components/ingame/GamePlay.svelte';
	import { gameStore } from '../../stores/gameStore';
	import { onDestroy } from 'svelte';
	import Transitioning from '../../components/ingame/Transitioning.svelte';

	let gate = $state<'join' | 'create'>('join');

	const MAX_RETRIES = 5;
	const BASE_RETRY_DELAY_MS = 1000;

	let ws: WebSocket | null = null;
	let retryCount = $state(0);
	let retryTimeout: ReturnType<typeof setTimeout> | null = null;
	let manualClose = false;

	let mapTreasures = $state<Treasure[]>([]);
	let wsReady = $state(false);
	let connectionState = $state<'connected' | 'reconnecting' | 'failed'>('connected');

	function connectToWs() {
		if (ws && (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING)) return;

		manualClose = false;
		ws = new WebSocket('ws://localhost:4000/ws');

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
					connectToWs();
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

		switch (data.type) {
			case Events.PlayerJoined: {
				const info = data.data as { id: string; nickname: string };
				if (info.id !== $gameStore.playerId) {
					gameStore.addPlayer(info.id, info.nickname);
				}
				break;
			}
			case Events.PlayerLeft: {
				const info = data.data as { id: string };
				gameStore.removePlayer(info.id);
				break;
			}
			case Events.ShowGameLoading: {
				gameStore.setGameState('loading');
				break;
			}
			case Events.NotEnoughPlayers: {
				console.log('NOT ENOUGH PLAYERS');
				break;
			}
			case Events.GameDataTreasures: {
				const info = data.data as { treasures: Treasure[]; map: 'first' | 'second' | 'third' };
				mapTreasures = info.treasures;
				gameStore.updateRoundTreasureCount(mapTreasures.length);
				gameStore.setMap(info.map);
				break;
			}
			case Events.BeginGame: {
				const info = data.data as { endsAt: number; serverNow: number };
				gameStore.startRound(info.serverNow, info.endsAt);
				gameStore.setGameState('in-game');
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

	onDestroy(disconnectWs);
</script>

<svelte:head>
	<title>Play - Crisis Control</title>
	<meta
		name="description"
		content="Join an existing game or create a new one and invite your friends to play."
	/>
	<meta name="robots" content="noindex, nofollow" />
	<meta property="og:title" content="Play - Crisis Control" />
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
	{#if $gameStore.state === 'waiting'}
		<Lobby
			{wsReady}
			onStartGame={async () => {
				sendMessage({ type: 'request:game:start', data: null });
			}}
			onLeaveGame={async () => {
				const res = await fetch('/api/game/leave', { method: 'POST' });

				if (!res.ok) {
					console.log('failed to logout');
				}

				disconnectWs();
				gameStore.reset();
				gate = 'join';
			}}
		/>
	{:else if $gameStore.state === 'loading'}
		<GameLoading />
	{:else if $gameStore.state === 'idle'}
		{#if gate === 'join'}
			<JoinGame
				onCreateGame={() => (gate = 'create')}
				onGameJoined={(player: Player, game: Game, members: Member[]) => {
					connectToWs();
					gameStore.setupGame(game, player, members);
				}}
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
		/>
	{:else if $gameStore.state === 'transitioning'}
		<Transitioning
			onNextRound={$gameStore.playerIsCreator
				? () => {
						console.log('Sending next round message');
						sendMessage({ type: Events.NextRound, data: null });
					}
				: undefined}
		/>
	{/if}
</div>
