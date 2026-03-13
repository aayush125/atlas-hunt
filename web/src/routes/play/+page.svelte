<script lang="ts">
	import type { Game } from '$lib/types/game';
	import type { Member, Player } from '$lib/types/player';
	import type { WSMessage } from '$lib/types/wsmessage';
	import CreateGame from '../../components/CreateGame.svelte';
	import JoinGame from '../../components/JoinGame.svelte';
	import Lobby from '../../components/Lobby.svelte';
	import { gameStore } from '../../stores/gameStore';
	import { onDestroy } from 'svelte';

	let gate = $state<'join' | 'create'>('join');

	const MAX_RETRIES = 5;
	const BASE_RETRY_DELAY_MS = 1000;

	let ws: WebSocket | null = null;
	let retryCount = 0;
	let retryTimeout: ReturnType<typeof setTimeout> | null = null;
	let manualClose = false;

	function connectToWs() {
		if (ws && (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING)) return;

		manualClose = false;
		ws = new WebSocket('ws://localhost:4000/ws');

		ws.onopen = () => {
			console.log('WebSocket connected');
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
			if (!manualClose && retryCount < MAX_RETRIES) {
				const delay = BASE_RETRY_DELAY_MS * 2 ** retryCount;
				console.log(`Reconnecting in ${delay}ms (attempt ${retryCount + 1}/${MAX_RETRIES})...`);
				retryTimeout = setTimeout(() => {
					retryCount++;
					connectToWs();
				}, delay);
			} else if (retryCount >= MAX_RETRIES) {
				console.error('Max reconnection attempts reached.');
				// Optionally: gameStore.setError('Connection lost. Please refresh.');
			}
		};

		ws.onerror = (error) => {
			console.error('WebSocket error:', error);
			ws?.close(); // triggers onclose → retry
		};
	}

	function handleMessage(data: WSMessage) {
		// Route incoming server messages to your store
		console.log('Message received:', data);
		if (data.type === 'player:joined') {
			const info = data.data as { id: string; nickname: string };
			if (info.id !== $gameStore.playerId) {
				gameStore.addPlayer(info.id, info.nickname);
			}
		} else if (data.type === 'player:left') {
			const info = data.data as { id: string };
			gameStore.removePlayer(info.id);
		}
	}

	function disconnectWs() {
		manualClose = true;
		if (retryTimeout) clearTimeout(retryTimeout);
		ws?.close();
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

	<!-- Open Graph -->
	<meta property="og:title" content="Play - Crisis Control" />
	<meta
		property="og:description"
		content="Join an existing game or create a new one and invite your friends to play."
	/>
	<meta property="og:type" content="website" />
</svelte:head>

<div class="flex flex-col items-center justify-center gap-8 p-8">
	{#if $gameStore.state === 'waiting'}
		<Lobby onLeaveGame={disconnectWs} />
	{:else if gate === 'join'}
		<JoinGame
			onCreateGame={() => (gate = 'create')}
			onGameJoined={(player: Player, game: Game, members: Member[]) => {
				connectToWs();
				console.log('Player received:', player);
				console.log('Game received:', game);
				console.log('Members received:', members);
				gameStore.setupGame(game, player, members);
			}}
		/>
	{:else}
		<CreateGame
			onJoinGame={() => (gate = 'join')}
			onGameCreated={(game: Game, player: Player) => {
				connectToWs();
				console.log('Player received:', player);
				console.log('Game received:', game);
				gameStore.setupGame(game, player, []);
			}}
		/>
	{/if}
</div>
