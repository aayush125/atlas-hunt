<script lang="ts">
	import type { TreasureItem } from '$lib/types/treasure';

	export let treasures: TreasureItem[] = [];
</script>

<div class="grid grid-cols-6 gap-1.5">
	{#each treasures as treasure (treasure.id)}
		<div
			class="group relative flex cursor-pointer flex-col items-center justify-center gap-0.5 rounded-lg border border-base-300 bg-base-200 px-1 py-2 transition-all duration-150 hover:border-warning/60 hover:bg-base-300 hover:shadow-md active:scale-95"
			title={treasure.id}
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
		</div>
	{/each}
</div>
