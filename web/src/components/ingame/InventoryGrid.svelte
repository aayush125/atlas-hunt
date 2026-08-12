<script lang="ts">
	import type { TreasureItem } from '$lib/types/treasure';
	import TreasureItemDialog from './TreasureItemDialog.svelte';

	let { treasures }: { treasures: TreasureItem[] } = $props();

	let sort = $state(false);
	let treasuresShown = $derived(
		sort ? [...treasures].sort((a, b) => a.name.localeCompare(b.name)) : treasures
	);

	let showTreasureModal = $state(false);
	let selectedTreasure = $state<TreasureItem | null>(null);
</script>

<div class="mb-3 flex justify-end">
	<button
		class={`btn gap-2 ${sort ? 'btn-active' : 'btn-outline'} btn-sm`}
		onclick={() => (sort = !sort)}
	>
		<svg
			xmlns="http://www.w3.org/2000/svg"
			width="16"
			height="16"
			viewBox="0 0 24 24"
			fill="none"
			stroke="currentColor"
			stroke-width="2"
			stroke-linecap="round"
			stroke-linejoin="round"
		>
			<path d="M3 6h18" />
			<path d="M6 12h12" />
			<path d="M10 18h4" />
		</svg>
		Sort
	</button>
</div>
<div class="grid grid-cols-6 gap-1.5">
	{#each treasuresShown as treasure (treasure.instanceId)}
		<button
			class="group relative flex cursor-pointer flex-col items-center justify-center gap-0.5 rounded-lg border border-base-300 bg-base-200 px-1 py-2 transition-all duration-150 hover:border-warning/60 hover:bg-base-300 hover:shadow-md active:scale-95"
			title={treasure.id}
			onclick={() => {
				selectedTreasure = treasure;
				showTreasureModal = true;
			}}
		>
			<div
				class="pointer-events-none absolute inset-0 rounded-lg opacity-0 transition-opacity duration-200 group-hover:opacity-100"
				style="background: radial-gradient(circle at 50% 30%, oklch(var(--wa) / 0.12), transparent 70%);"
			></div>
			<span class="text-lg leading-none select-none" role="img" aria-label={treasure.id}>
				{treasure.icon}
			</span>
			<span
				class="max-w-full truncate px-0.5 text-center text-[9px] font-medium tracking-wide text-base-content/50 group-hover:text-base-content/80"
			>
				{treasure.name}
			</span>
		</button>
	{/each}

	{#if showTreasureModal && selectedTreasure}
		<TreasureItemDialog
			item={selectedTreasure}
			onClose={() => {
				showTreasureModal = false;
				selectedTreasure = null;
			}}
			onDonate={() => {}}
		/>
	{/if}
</div>
