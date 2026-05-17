<script lang="ts">
	import type { TreasureItem } from '$lib/types/treasure';

	let { treasures, onClose }: { treasures: TreasureItem[]; onClose: () => void } = $props();

	function rarityClass(rarity: string) {
		switch (rarity.toLowerCase()) {
			case 'common':
				return 'badge-ghost';
			case 'uncommon':
				return 'badge-info';
			case 'rare':
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
		class="animate-pop pointer-events-auto flex w-full max-w-lg flex-col rounded-2xl border border-base-300 bg-base-100 shadow-2xl"
		role="dialog"
		aria-label="Treasure review"
	>
		<!-- Close -->
		<button class="btn absolute top-3 right-3 z-10 btn-circle btn-ghost btn-sm" onclick={onClose}>
			✕
		</button>

		<!-- Header -->
		<div class="flex items-center gap-3 border-b border-base-300 px-5 pt-5 pb-4">
			<div class="flex-1">
				<p class="text-[10px] font-medium tracking-widest text-base-content/40 uppercase">
					Round complete
				</p>
				<h2 class="mt-0.5 text-lg leading-tight font-semibold">Your treasures</h2>
			</div>
			<span class="badge badge-ghost badge-md"
				>{treasures.length} item{treasures.length === 1 ? '' : 's'}</span
			>
		</div>

		<!-- Grid -->
		<div
			class="grid grid-cols-4 gap-2 overflow-y-auto p-4 sm:grid-cols-5"
			style="max-height: 400px;"
		>
			{#each treasures as treasure (treasure.instanceId)}
				<div
					class="flex flex-col items-center gap-1.5 rounded-xl border border-base-300 bg-base-200 px-2 py-3 select-none"
				>
					<div
						class="flex h-11 w-11 items-center justify-center rounded-full bg-base-100 text-2xl shadow-inner"
					>
						{treasure.icon}
					</div>
					<span
						class="line-clamp-2 text-center text-[11px] leading-tight font-medium text-base-content/80"
					>
						{treasure.name}
					</span>
					<span class="badge badge-xs {rarityClass(treasure.rarity)}">{treasure.rarity}</span>
					<span class="text-[10px] text-base-content/40">{treasure.points}</span>
				</div>
			{/each}
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
