<script lang="ts">
	import { INVENTORY_SIZE } from '$lib/constants/gameplay';
	import type { Treasure } from '$lib/types/treasure';
	import { gameStore } from '../../stores/gameStore';
	import InfoIcon from '../icons/InfoIcon.svelte';
	import LeaveIcon from '../icons/LeaveIcon.svelte';
	import GameInfoDialog from './GameInfoDialog.svelte';
	import InventoryGrid from './InventoryGrid.svelte';
	import LeaveGameDialog from './LeaveGameDialog.svelte';
	import Map from './Map.svelte';
	import PlayerList from './PlayerList.svelte';
	import Requests from './Requests.svelte';

	let {
		treasures,
		onClaim,
		onFail,
		onLeaveGame
	}: {
		treasures: Treasure[];
		onClaim: (t: Treasure) => void;
		onFail: (t: Treasure) => void;
		onLeaveGame: () => Promise<void>;
	} = $props();

	let showInfoModal = $state(false);
	let showLeaveConfirm = $state(false);
	let activeTab = $state<'inventory' | 'players' | 'donations'>('inventory');

	let now = $state(Date.now());
	let timeLeft = $derived.by(() => {
		if (!$gameStore.serverNow || !$gameStore.roundEndsAt) return '00:00';
		const effectiveNow = Math.max(now, $gameStore.serverNow);
		const remainingMs = Math.max(0, $gameStore.roundEndsAt - effectiveNow);

		const totalSeconds = Math.floor(remainingMs / 1000);
		const minutes = Math.floor(totalSeconds / 60);
		const seconds = totalSeconds % 60;

		const mm = String(minutes).padStart(2, '0');
		const ss = String(seconds).padStart(2, '0');

		return `${mm}:${ss}`;
	});

	$effect(() => {
		const interval = setInterval(() => {
			now = Date.now();

			if ($gameStore.roundEndsAt && now >= $gameStore.roundEndsAt) {
				clearInterval(interval);
			}
		}, 1000);

		return () => clearInterval(interval);
	});

	function openInfo() {
		showInfoModal = true;
	}
	function closeInfo() {
		showInfoModal = false;
	}
	function promptLeave() {
		showLeaveConfirm = true;
	}
	function cancelLeave() {
		showLeaveConfirm = false;
	}
	async function confirmLeave() {
		await onLeaveGame();
		showLeaveConfirm = false;
	}
</script>

<div class="flex h-full w-full flex-col">
	<!-- MAP -->
	<div class="relative min-h-0 flex-1">
		<Map allTreasures={treasures} {onClaim} {onFail} />

		<!-- LEFT: TIMER -->
		<div class="absolute top-3 left-3 z-10">
			<div
				class="flex items-center gap-2 rounded-xl bg-base-100/90 px-3 py-2 shadow-lg backdrop-blur"
			>
				<div class="text-lg">⏱️</div>

				{#if $gameStore.serverNow && $gameStore.roundEndsAt}
					<div class="leading-tight">
						<div class="text-[10px] uppercase opacity-60">Time Left</div>
						<div class="font-mono text-2xl font-black tracking-widest tabular-nums">
							{timeLeft}
						</div>
					</div>
				{/if}
			</div>
		</div>

		<!-- RIGHT: FLOATING BUTTONS -->
		<div class="absolute top-3 right-3 z-10 grid grid-cols-2 gap-1.5 shadow-md">
			<div class="btn pointer-events-none btn-xs btn-warning">
				🏆 {$gameStore.points}
			</div>

			<div class="btn pointer-events-none btn-xs btn-success">
				🗺️ {treasures.length}/{$gameStore.treasureCountThisRound}
			</div>

			<button class="btn btn-xs btn-info" onclick={openInfo} aria-label="Game info">
				<InfoIcon />
				Info
			</button>

			<button class="btn btn-xs btn-error" onclick={promptLeave} aria-label="Leave game">
				<LeaveIcon />
				Leave
			</button>
		</div>
	</div>

	<!-- TABBED SECTION -->
	<div class="flex flex-col border-t border-base-300 bg-base-100">
		<!-- Tab headers -->
		<div role="tablist" class="tabs-border tabs px-4 pt-2">
			<button
				role="tab"
				class="tab {activeTab === 'inventory' ? 'tab-active' : ''}"
				onclick={() => (activeTab = 'inventory')}
			>
				Inventory ({$gameStore.treasures.length})
				{#if $gameStore.treasures.length >= INVENTORY_SIZE}
					<span class="badge badge-sm badge-warning">Full</span>
				{/if}
			</button>
			<button
				role="tab"
				class="tab {activeTab === 'players' ? 'tab-active' : ''}"
				onclick={() => (activeTab = 'players')}
			>
				Players
			</button>
			<button
				role="tab"
				class="tab {activeTab === 'donations' ? 'tab-active' : ''}"
				onclick={() => (activeTab = 'donations')}
			>
				Requests
			</button>
		</div>

		<!-- Tab content -->
		<div class="h-48 overflow-y-auto p-4">
			{#if activeTab === 'inventory'}
				<div class="mt-8 text-center text-sm text-base-content/50">
					{#if $gameStore.treasures.length === 0}
						No items in inventory yet.
					{:else}
						<InventoryGrid treasures={$gameStore.treasures} />
					{/if}
				</div>
			{:else if activeTab == 'players'}
				<PlayerList players={$gameStore.otherPlayers} />
			{:else}
				<Requests
					requests={{
						player_001: {
							treasureId: 'ng02',
							requestedAmount: 100,
							amountReceived: 35
						},
						player_002: {
							treasureId: 'st05',
							requestedAmount: 250,
							amountReceived: 120
						},
						player_003: {
							treasureId: 'hc07',
							requestedAmount: 150,
							amountReceived: 150
						},
						player_004: {
							treasureId: 'fd07',
							requestedAmount: 500,
							amountReceived: 275
						},
						player_005: {
							treasureId: 'ae07',
							requestedAmount: 75,
							amountReceived: 10
						}
					}}
				/>
			{/if}
		</div>
	</div>

	<!-- INFO MODAL -->
	{#if showInfoModal}
		<GameInfoDialog onClose={closeInfo} />
	{/if}

	<!-- LEAVE CONFIRM MODAL -->
	{#if showLeaveConfirm}
		<LeaveGameDialog onCancelLeave={cancelLeave} onConfirmLeave={confirmLeave} />
	{/if}
</div>
