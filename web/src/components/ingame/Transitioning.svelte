<script lang="ts">
	import avatarSrc from '$lib/assets/avatars/avatar.jpg';
	import type { GameMap } from '$lib/types/game';
	import { gameStore } from '../../stores/gameStore';
	import InventoryListDialog from './InventoryListDialog.svelte';

	let { onNextRound, onNewGame }: { onNextRound?: () => void; onNewGame?: () => void } = $props();

	let showInventoryDialog = $state(false);

	function translateGameMapRound(map: GameMap | null) {
		if (!map) return 0;
		switch (map) {
			case 'first':
				return 1;
			case 'second':
				return 2;
			default:
				return 3;
		}
	}

	function getPlayerNicknameFromId(id: string) {
		if (id === $gameStore.playerId) return $gameStore.playerName;
		const player = $gameStore.otherPlayers.find((p) => p.id === id);
		return player?.nickname;
	}
</script>

{#if $gameStore.map && $gameStore.mapRoundStats[$gameStore.map] !== null}
	{@const map = $gameStore.map}
	{@const stats = $gameStore.mapRoundStats[map]}

	{#if stats}
		<div class="flex h-full w-full flex-col overflow-y-auto bg-base-100">
			<!-- Header -->
			<div class="py-6 text-center">
				<div class="text-xs tracking-widest uppercase opacity-50">Round complete</div>
				<div class="mt-1 text-3xl font-black">
					Round {translateGameMapRound($gameStore.map)} of 3
				</div>
			</div>

			<div class="flex flex-col gap-4 px-4 pb-6">
				<!-- Stats card -->
				<div class="card border border-base-300 bg-base-100 shadow-sm">
					<div class="card-body gap-3 p-4">
						<div class="text-[10px] tracking-widest uppercase opacity-50">Your stats</div>

						<div class="grid grid-cols-2 gap-2">
							<div class="stat rounded-xl bg-base-200 p-3">
								<div class="stat-title text-[11px]">🏆 Points</div>
								<div class="stat-value text-2xl">{$gameStore.points}</div>
							</div>
							<div class="stat rounded-xl bg-base-200 p-3">
								<div class="stat-title text-[11px]">🗺️ Treasures</div>
								<div class="stat-value text-2xl">
									{stats.treasures_claimed + stats.treasures_failed} / {$gameStore.treasureCountThisRound}
								</div>
							</div>
							<div class="stat rounded-xl bg-base-200 p-3">
								<div class="stat-title text-[11px]">✅ Claims</div>
								<div class="stat-value text-2xl">
									{stats.treasures_claimed}
								</div>
							</div>
							<div class="stat rounded-xl bg-base-200 p-3">
								<div class="stat-title text-[11px]">❌ Failed</div>
								<div class="stat-value text-2xl">
									{stats.treasures_failed}
								</div>
							</div>
						</div>
					</div>
				</div>

				<!-- Rankings card -->
				<div class="card border border-base-300 bg-base-100 shadow-sm">
					<div class="card-body gap-3 p-4">
						<div class="text-[10px] tracking-widest uppercase opacity-50">Standings</div>

						<ul class="flex max-h-80 flex-col gap-1 overflow-y-auto">
							{#each Object.entries(stats.all_player_points).sort((a, b) => b[1] - a[1]) as [player, points], i (player)}
								<li
									class="flex items-center gap-3 rounded-xl px-2 py-2
      {player === $gameStore.playerName ? 'bg-info/10' : ''}"
								>
									<span class="w-5 text-center text-xs font-bold opacity-50">
										{#if i === 0}🥇{:else if i === 1}🥈{:else if i === 2}🥉{:else}{i + 1}{/if}
									</span>
									<div class="placeholder avatar">
										<img
											src={avatarSrc}
											alt="Player avatar"
											class="h-12 w-12 rounded-full object-cover shadow-md ring-2 ring-secondary"
										/>
									</div>
									<span
										class="flex-1 truncate text-sm {player === $gameStore.playerId
											? 'font-bold'
											: ''}"
									>
										{getPlayerNicknameFromId(player)}
									</span>
									{#if player === $gameStore.playerId}
										<span class="badge badge-sm badge-info">you</span>
									{/if}
									<span class="font-mono text-sm font-bold tabular-nums">
										{points.toLocaleString()}
									</span>
								</li>
							{/each}
						</ul>
					</div>
				</div>

				<!-- Actions -->
				<button class="btn btn-outline btn-sm" onclick={() => (showInventoryDialog = true)}>
					🎒 Inventory
				</button>
				{#if $gameStore.playerIsCreator && $gameStore.map !== 'third' && onNextRound}
					<button class="btn btn-outline btn-sm" onclick={onNextRound}>
						⚔️ Begin Next Round
					</button>
				{/if}
				{#if $gameStore.map === 'third' && onNewGame}
					<button class="btn btn-outline btn-sm" onclick={onNewGame}> ⚔️ Start a New Game </button>
				{/if}
			</div>
		</div>
	{/if}
{:else}
	<div>Error</div>
{/if}

{#if showInventoryDialog}
	<InventoryListDialog
		treasures={$gameStore.treasures}
		onClose={() => (showInventoryDialog = false)}
	/>
{/if}
