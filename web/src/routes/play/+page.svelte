<script lang="ts">
	import CreateGame from '../../components/CreateGame.svelte';
	import JoinGame from '../../components/JoinGame.svelte';
	import avatarSrc from '$lib/assets/avatars/avatar.jpg';
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
				handleMessage(data);
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

	function handleMessage(data: unknown) {
		// Route incoming server messages to your store
		console.log('Message received:', data);
		// e.g. gameStore.handleServerMessage(data);
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
		<div class="flex w-full flex-col items-center justify-center gap-6 p-4 sm:p-6">
			<div class="w-full max-w-lg space-y-6 rounded-xl bg-base-100 p-6 shadow-xl sm:p-8">
				<!-- Room Code -->
				<div class="space-y-2 text-center">
					<h2 class="text-xs font-medium tracking-wide text-base-content/60 sm:text-sm">
						ROOM CODE
					</h2>

					<div
						class="rounded-lg bg-primary py-3 text-3xl font-bold tracking-widest text-primary-content shadow-md select-all sm:py-4 sm:text-4xl"
					>
						{$gameStore.joinCode}
					</div>

					<p class="text-xs text-base-content/70 sm:text-sm">
						Share this code with your friends to join the game.
					</p>
				</div>

				<!-- Players -->
				<div class="space-y-4">
					<div class="flex items-center justify-between">
						<h3
							class="text-xs font-semibold tracking-wide text-base-content/80 uppercase sm:text-sm"
						>
							Players
						</h3>
						<span class="text-xs text-base-content/60 sm:text-sm"> 4 / 8 Joined </span>
					</div>

					<div class="grid grid-cols-3 gap-4 sm:grid-cols-4 md:grid-cols-5">
						<!-- Player Card -->
						<div class="flex flex-col items-center space-y-1 sm:space-y-2">
							<img
								src={avatarSrc}
								alt="Player avatar"
								class="h-16 w-16 rounded-full object-cover shadow-md ring-2 ring-secondary sm:h-20 sm:w-20"
							/>
							<span class="text-center text-xs font-medium text-base-content sm:text-sm">
								{$gameStore.playerName}
							</span>
						</div>

						<div class="flex flex-col items-center space-y-1 sm:space-y-2">
							<img
								src={avatarSrc}
								alt="Player avatar"
								class="h-16 w-16 rounded-full object-cover shadow-md ring-2 ring-secondary sm:h-20 sm:w-20"
							/>
							<span class="text-center text-xs font-medium text-base-content sm:text-sm"
								>CoolNick</span
							>
						</div>

						<div class="flex flex-col items-center space-y-1 sm:space-y-2">
							<img
								src={avatarSrc}
								alt="Player avatar"
								class="h-16 w-16 rounded-full object-cover shadow-md ring-2 ring-secondary sm:h-20 sm:w-20"
							/>
							<span class="text-center text-xs font-medium text-base-content sm:text-sm"
								>GamerX</span
							>
						</div>

						<div class="flex flex-col items-center space-y-1 sm:space-y-2">
							<img
								src={avatarSrc}
								alt="Player avatar"
								class="h-16 w-16 rounded-full object-cover shadow-md ring-2 ring-secondary sm:h-20 sm:w-20"
							/>
							<span class="text-center text-xs font-medium text-base-content sm:text-sm"
								>GamerX</span
							>
						</div>

						<!-- Empty Slot -->
						<div class="flex flex-col items-center space-y-1 opacity-60 sm:space-y-2">
							<div
								class="flex h-16 w-16 items-center justify-center rounded-full bg-base-300 text-lg font-bold text-base-content/40 sm:h-20 sm:w-20 sm:text-xl"
							>
								+
							</div>
							<span class="text-center text-xs text-base-content/50 sm:text-sm">Waiting...</span>
						</div>
					</div>
				</div>
			</div>

			<button class="btn mt-4 w-full max-w-lg btn-primary sm:mt-6">Start Game</button>
		</div>
	{:else if gate === 'join'}
		<JoinGame onCreateGame={() => (gate = 'create')} />
	{:else}
		<CreateGame
			onJoinGame={() => (gate = 'join')}
			onGameCreated={(joinCode: string, nickname: string, numPlayers: number, createdAt: Date) => {
				connectToWs();
				gameStore.setupGame(joinCode, numPlayers, nickname, createdAt);
			}}
		/>
	{/if}
</div>
