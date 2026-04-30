<script lang="ts">
	import type { Treasure } from '$lib/types/treasure';

	let {
		treasure,
		onClose,
		onTreasureClaim,
		claimResult
	}: {
		treasure: Treasure;
		onClose: () => void;
		onTreasureClaim: (treasure: Treasure, selectedIndex: number) => void;
		claimResult: 'correct' | 'incorrect' | null;
	} = $props();

	let selectedOption = $state<number | null>(null);
	let submitted = $state(false);

	function handleClaim() {
		if (selectedOption === null) return;
		submitted = true;
		onTreasureClaim(treasure, selectedOption);
	}

	function optionClass(i: number): string {
		const base = 'btn btn-outline w-full justify-start text-left font-normal normal-case';
		if (!submitted || claimResult === null)
			return `${base} ${selectedOption === i ? 'btn-primary' : ''}`;
		if (selectedOption === i)
			return `${base} ${claimResult === 'correct' ? 'btn-success' : 'btn-error'}`;
		return `${base} opacity-40`;
	}
</script>

<!-- Backdrop -->
<div
	class="fixed inset-0 z-40 bg-black/40 backdrop-blur-sm"
	role="presentation"
	onclick={onClose}
></div>

<!-- Dialog -->
<div class="pointer-events-none fixed inset-0 z-50 flex items-center justify-center">
	<div
		class="animate-pop pointer-events-auto card w-96 border border-base-300 bg-base-100 shadow-2xl"
		role="dialog"
		aria-modal="true"
		aria-labelledby="treasure-title"
	>
		<div class="card-body gap-4">
			<!-- Close button -->
			<button
				class="btn absolute top-3 right-3 btn-circle btn-ghost btn-sm"
				onclick={onClose}
				aria-label="Close">✕</button
			>

			<!-- Header -->
			<div class="flex items-center gap-3">
				<div
					class="flex h-14 w-14 shrink-0 items-center justify-center rounded-full bg-base-200 text-4xl shadow-inner"
				>
					{treasure.icon}
				</div>
				<h2 id="treasure-title" class="card-title text-lg leading-tight">
					{treasure.name}
				</h2>
			</div>

			<div class="divider my-0"></div>

			<!-- Question -->
			<p class="text-sm leading-relaxed font-medium">
				{treasure.question.text}
			</p>

			<!-- Options -->
			<div class="flex flex-col gap-2">
				{#each treasure.question.options as option, i (i)}
					<button class={optionClass(i)} disabled={submitted} onclick={() => (selectedOption = i)}>
						<span class="mr-2 opacity-50">{String.fromCharCode(65 + i)}.</span>
						{option}
					</button>
				{/each}
			</div>

			<!-- Feedback -->
			{#if claimResult !== null}
				<div
					class={`alert-sm alert ${claimResult === 'correct' ? 'alert-success' : 'alert-error'}`}
				>
					<span class="text-sm">
						{claimResult === 'correct'
							? '✨ Correct! Treasure claimed.'
							: '✗ Wrong answer, better luck next time.'}
					</span>
				</div>
			{/if}

			<!-- Claim button -->
			<div class="mt-1 card-actions">
				<button
					class="btn w-full gap-2 btn-sm btn-primary"
					disabled={selectedOption === null || submitted}
					onclick={handleClaim}
				>
					{#if submitted && claimResult === null}
						<span class="loading loading-xs loading-spinner"></span>
					{/if}
					<span>Claim treasure</span>
					<span>✨</span>
				</button>
			</div>
		</div>
	</div>
</div>

<style>
	@keyframes pop {
		0% {
			opacity: 0;
			transform: scale(0.88) translateY(8px);
		}
		60% {
			transform: scale(1.03) translateY(-2px);
		}
		100% {
			opacity: 1;
			transform: scale(1) translateY(0);
		}
	}
	.animate-pop {
		animation: pop 0.22s cubic-bezier(0.34, 1.56, 0.64, 1) both;
	}
</style>
