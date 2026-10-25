<script lang="ts">
	import { gameStore } from '../../stores/gameStore';
	import { rarityClass, rawTreasures } from '$lib/treasures';

	let {
		treasureId,
		reqAmount,
		recAmount,
		donateTo,
		onCancel,
		onDonate
	}: {
		treasureId: string;
		reqAmount: number;
		recAmount: number;
		donateTo: string;
		onCancel: () => void;
		onDonate: (treasureId: string, amount: number, donateTo: string) => void;
	} = $props();

	let treasure = $derived(rawTreasures.find((t) => t.id === treasureId));
	let treasureCount = $derived($gameStore.treasureCount[treasureId] ?? 0);

	let donatedAmount = $state<number | null>(1);

	let isValidAmount = $derived(
		Number.isInteger(donatedAmount) &&
			(donatedAmount ?? 0) > 0 &&
			(donatedAmount ?? 0) <= treasureCount
	);
	let canSubmit = $derived(!!treasure && treasureCount > 0 && isValidAmount);

	let errorMessage = $derived.by(() => {
		if (donatedAmount === null || isValidAmount) return '';
		if (!Number.isInteger(donatedAmount) || donatedAmount < 1) {
			return 'Enter a whole number of 1 or more.';
		}
		return `You only have ${treasureCount} of this treasure.`;
	});

	function submit(event: SubmitEvent) {
		event.preventDefault();
		if (canSubmit && donatedAmount) onDonate(treasureId, donatedAmount, donateTo);
	}

	function onKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') onCancel();
	}
</script>

<svelte:window onkeydown={onKeydown} />

<dialog class="modal-open modal" aria-modal="true" aria-labelledby="treasure-donate-title">
	<form class="modal-box max-w-md p-0" onsubmit={submit}>
		<header class="flex items-start gap-3 border-b border-base-300 px-6 py-5">
			<div
				class="flex size-11 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-xl"
				aria-hidden="true"
			>
				🎁
			</div>

			<div>
				<h3 id="treasure-donate-title" class="text-lg font-bold">Donate a treasure</h3>
				<p class="mt-0.5 text-sm text-base-content/60">
					Give some of your treasure to another player.
				</p>
			</div>
		</header>

		{#if treasure && treasureCount > 0}
			<div class="space-y-5 px-6 py-5">
				<div class="flex items-center gap-3 rounded-xl border border-base-300 bg-base-100 p-3">
					<div
						class="flex size-12 shrink-0 items-center justify-center rounded-xl bg-base-200 text-2xl"
						aria-hidden="true"
					>
						{treasure.icon}
					</div>

					<div class="min-w-0 flex-1">
						<span class="block truncate font-semibold">{treasure.name}</span>

						<div class="mt-1 flex flex-wrap gap-1.5">
							<span class="badge badge-ghost badge-sm">{treasure.cat}</span>
							<span class={`badge badge-xs ${rarityClass(treasure.rarity)}`}>{treasure.rarity}</span
							>
						</div>

						<div class="mt-1 text-xs text-base-content/60">
							{treasure.points} points · You have {treasureCount} · Needs {reqAmount - recAmount}
						</div>
					</div>
				</div>

				<div>
					<label class="label px-0" for="donated-amount">
						<span class="label-text font-semibold">Amount</span>
					</label>

					<div class="join w-full">
						<input
							id="donated-amount"
							type="number"
							min="1"
							max={treasureCount}
							step="1"
							inputmode="numeric"
							placeholder="1"
							class="input-bordered input join-item w-full {errorMessage ? 'input-error' : ''}"
							bind:value={donatedAmount}
						/>
						<button
							type="button"
							class="btn join-item btn-outline"
							onclick={() =>
								(donatedAmount =
									treasureCount < reqAmount - recAmount ? treasureCount : reqAmount - recAmount)}
						>
							Max
						</button>
					</div>

					{#if errorMessage}
						<p class="mt-1 text-sm text-error">{errorMessage}</p>
					{/if}
				</div>
			</div>

			<footer
				class="flex items-center justify-between gap-2 border-t border-base-300 bg-base-200/30 px-6 py-4"
			>
				<p class="min-w-0 truncate text-sm text-base-content/60">
					{#if isValidAmount}
						{donatedAmount} × {treasure.name}
					{/if}
				</p>

				<div class="flex gap-2">
					<button type="button" class="btn btn-ghost" onclick={onCancel}>Cancel</button>
					<button type="submit" class="btn btn-primary" disabled={!canSubmit}>Donate</button>
				</div>
			</footer>
		{:else}
			<div class="px-6 py-8 text-center text-sm text-base-content/60">
				You do not have this treasure.
			</div>

			<footer class="flex justify-end border-t border-base-300 bg-base-200/30 px-6 py-4">
				<button type="button" class="btn btn-ghost" onclick={onCancel}>Close</button>
			</footer>
		{/if}
	</form>

	<button type="button" class="modal-backdrop" aria-label="Close" tabindex="-1" onclick={onCancel}
	></button>
</dialog>
