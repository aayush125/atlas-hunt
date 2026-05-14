<script lang="ts">
	import type { TreasureItem } from '$lib/types/treasure';
	import { gameStore } from '../../stores/gameStore';

	let {
		item,
		onClose,
		onDonate,
		onDiscard
	}: {
		item: TreasureItem;
		onClose: () => void;
		onDonate: (item: TreasureItem) => void;
		onDiscard: (item: TreasureItem) => void;
	} = $props();

	function rarityClass(rarity: string) {
		switch (rarity.toLowerCase()) {
			case 'common':
				return 'badge-ghost';
			case 'rare':
				return 'badge-info';
			case 'epic':
				return 'badge-secondary';
			case 'legendary':
				return 'badge-warning';
			default:
				return 'badge-outline';
		}
	}
</script>

<!-- Backdrop -->
<div
	class="fixed inset-0 z-40 bg-black/50 backdrop-blur-sm"
	role="presentation"
	onclick={onClose}
></div>

<!-- Dialog -->
<div class="pointer-events-none fixed inset-0 z-50 flex items-center justify-center p-4">
	<div
		class="animate-pop pointer-events-auto w-full max-w-md rounded-2xl border border-base-300 bg-base-100 shadow-2xl"
		role="dialog"
		aria-label={item.name}
	>
		<div class="space-y-4 p-5">
			<!-- Close -->
			<button class="btn absolute top-3 right-3 btn-circle btn-ghost btn-sm" onclick={onClose}>
				✕
			</button>

			<!-- Header -->
			<div class="flex items-center gap-3">
				<div
					class="flex h-14 w-14 items-center justify-center rounded-full bg-base-200 text-3xl shadow-inner"
				>
					{item.icon}
				</div>
				<div class="flex flex-col">
					<h2 class="text-left text-lg font-semibold">{item.name}</h2>
					<div class="mt-1 flex items-center gap-2">
						<span class={`badge badge-sm ${rarityClass(item.rarity)}`}>
							{item.rarity}
						</span>
						<span class="text-xs opacity-70">×{$gameStore.treasureCount[item.id]} owned</span>
						<span class="text-xs opacity-70">ID: {item.id}</span>
					</div>
				</div>
			</div>

			<!-- Rewards -->
			<div class="grid grid-cols-2 gap-2 text-sm">
				<div class="flex justify-between rounded-lg bg-base-200 px-3 py-2">
					<span>Points</span>
					<span class="font-semibold">+{item.points}</span>
				</div>
				<div class="flex justify-between rounded-lg bg-base-200 px-3 py-2">
					<span>Boost</span>
					<span class="font-semibold text-primary">+{item.boostPoints}</span>
				</div>
			</div>

			<div class="divider my-1"></div>

			<p class="text-sm leading-relaxed opacity-70">
				Donate this treasure to help another player, or discard it to free up inventory space.
			</p>

			<!-- Actions -->
			<div class="flex flex-col gap-2 pt-1">
				<button class="btn w-full gap-2 btn-primary" onclick={() => onDonate(item)}>
					<span>🎁</span>
					<span>Donate Treasure</span>
				</button>
				<button class="btn w-full gap-2 btn-outline btn-error" onclick={() => onDiscard(item)}>
					<span>🗑️</span>
					<span>Discard Treasure</span>
				</button>
			</div>
		</div>
	</div>
</div>

<style>
	@keyframes pop {
		0% {
			opacity: 0;
			transform: scale(0.9) translateY(10px);
		}
		100% {
			opacity: 1;
			transform: scale(1) translateY(0);
		}
	}
	.animate-pop {
		animation: pop 0.2s ease-out;
	}
</style>
