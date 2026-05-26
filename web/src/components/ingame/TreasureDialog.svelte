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
		onTreasureClaim: (treasure: Treasure, selectedAnswer: string) => Promise<void>;
		claimResult: 'correct' | 'incorrect' | null;
	} = $props();

	let selectedOption = $state<string | null>(null);
	let submitted = $state(false);

	async function handleClaim() {
		if (selectedOption === null) return;
		submitted = true;
		await onTreasureClaim(treasure, selectedOption);
	}

	function optionClass(op: string): string {
		const base = 'btn w-full justify-start text-left normal-case transition-all duration-150';

		if (!submitted || claimResult === null)
			return `${base} ${
				selectedOption === op
					? 'btn-primary ring-2 ring-primary/40'
					: 'btn-outline hover:scale-[1.01]'
			}`;

		if (selectedOption === op)
			return `${base} ${claimResult === 'correct' ? 'btn-success' : 'btn-error'}`;

		return `${base} opacity-40`;
	}

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
					{treasure.icon}
				</div>

				<div class="flex flex-col">
					<h2 class="text-lg leading-tight font-semibold">
						{treasure.name}
					</h2>

					<div class="mt-1 flex items-center gap-2">
						<span class={`badge badge-sm ${rarityClass(treasure.rarity)}`}>
							{treasure.rarity}
						</span>

						<span class="text-xs opacity-70">
							ID: {treasure.id}
						</span>
					</div>
				</div>
			</div>

			<!-- Rewards -->
			<div class="grid w-full grid-cols-2 gap-2 text-sm">
				<div class="flex justify-between rounded-lg bg-base-200 px-3 py-2">
					<span>Points</span>
					<span class="font-semibold">+{treasure.points}</span>
				</div>
			</div>

			<div class="divider my-1"></div>

			<!-- Question -->
			<p class="text-sm leading-relaxed font-medium">
				{treasure.question.q}
			</p>

			<!-- Options -->
			<div class="flex flex-col gap-2">
				{#each Object.entries(treasure.question.options) as [key, option] (key)}
					<button
						class={optionClass(key)}
						disabled={submitted}
						onclick={() => (selectedOption = key)}
					>
						<span class="mr-2 opacity-50">{key}.</span>
						<span>{option}</span>
					</button>
				{/each}
			</div>

			<!-- Feedback -->
			{#if claimResult !== null}
				<div class={`alert text-sm ${claimResult === 'correct' ? 'alert-success' : 'alert-error'}`}>
					<span>
						{claimResult === 'correct'
							? `🎉 Correct! You earned ${treasure.points} points.`
							: '❌ Incorrect. Try again next time.'}
					</span>
				</div>
			{/if}

			<!-- Actions -->
			<div class="pt-1">
				<button
					class="btn w-full gap-2 btn-primary"
					disabled={selectedOption === null || submitted}
					onclick={handleClaim}
				>
					{#if submitted && claimResult === null}
						<span class="loading loading-xs loading-spinner"></span>
					{/if}

					<span>Claim Treasure</span>
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
