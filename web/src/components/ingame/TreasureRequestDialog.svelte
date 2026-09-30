<script lang="ts">
	import { rarityClass, rawTreasures } from '$lib/treasures';

	let {
		onCancel,
		onRequest
	}: {
		onCancel: () => void;
		onRequest: (treasureId: string, amount: number) => void;
	} = $props();

	const treasures = rawTreasures;

	let searchTerm = $state('');
	let selectedTreasureId = $state('');
	let requestedAmount = $state<number | null>(1);

	let filtered = $derived.by(() => {
		const query = searchTerm.trim().toLowerCase();
		return query ? treasures.filter((t) => t.name.toLowerCase().includes(query)) : treasures;
	});

	let selectedTreasure = $derived(treasures.find((t) => t.id === selectedTreasureId));

	let isValidAmount = $derived(Number.isInteger(requestedAmount) && (requestedAmount ?? 0) > 0);
	let canSubmit = $derived(!!selectedTreasure && isValidAmount);

	const cardBase =
		'flex items-center gap-3 rounded-xl border p-3 text-left transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary';
	const cardSelected = 'border-primary bg-primary/10 ring-2 ring-primary/20';
	const cardIdle = 'border-base-300 bg-base-100 hover:border-primary/50 hover:bg-base-200/50';

	function submit(event: SubmitEvent) {
		event.preventDefault();
		if (canSubmit && requestedAmount) onRequest(selectedTreasureId, requestedAmount);
	}

	function onKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') onCancel();
	}
</script>

<svelte:window onkeydown={onKeydown} />

<dialog class="modal-open modal" aria-modal="true" aria-labelledby="treasure-request-title">
	<form class="modal-box max-w-2xl p-0" onsubmit={submit}>
		<header class="flex items-start gap-3 border-b border-base-300 px-6 py-5">
			<div
				class="flex size-11 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-xl"
				aria-hidden="true"
			>
				💎
			</div>

			<div>
				<h3 id="treasure-request-title" class="text-lg font-bold">Request a treasure</h3>
				<p class="mt-0.5 text-sm text-base-content/60">
					Ask other players to donate a treasure to you.
				</p>
			</div>
		</header>

		<div class="space-y-5 px-6 py-5">
			<fieldset class="space-y-3">
				<legend class="label-text mb-2 font-semibold">Choose a treasure</legend>

				<input
					type="search"
					placeholder="Search treasures"
					aria-label="Search treasures"
					class="input-bordered input w-full"
					bind:value={searchTerm}
				/>

				{#if filtered.length === 0}
					<p class="py-8 text-center text-sm text-base-content/60">
						No treasures match "{searchTerm}". Try a different name.
					</p>
				{:else}
					<div class="grid max-h-80 grid-cols-1 gap-2 overflow-y-auto pr-1 sm:grid-cols-2">
						{#each filtered as treasure (treasure.id)}
							{@const selected = selectedTreasureId === treasure.id}

							<button
								type="button"
								class="{cardBase} {selected ? cardSelected : cardIdle}"
								aria-pressed={selected}
								onclick={() => (selectedTreasureId = treasure.id)}
							>
								<div
									class="flex size-12 shrink-0 items-center justify-center rounded-xl text-2xl {selected
										? 'bg-primary/15'
										: 'bg-base-200'}"
									aria-hidden="true"
								>
									{treasure.icon}
								</div>

								<div class="min-w-0 flex-1">
									<div class="flex items-center justify-between gap-2">
										<span class="truncate font-semibold">{treasure.name}</span>
										{#if selected}
											<span class="font-bold text-primary" aria-hidden="true">✓</span>
										{/if}
									</div>

									<div class="mt-1 flex flex-wrap gap-1.5">
										<span class="badge badge-ghost badge-sm">{treasure.cat}</span>
										<span class={`badge badge-xs ${rarityClass(treasure.rarity)}`}
											>{treasure.rarity}</span
										>
									</div>

									<div class="mt-1 text-xs text-base-content/60">{treasure.points} points</div>
								</div>
							</button>
						{/each}
					</div>
				{/if}
			</fieldset>

			<div>
				<label class="label px-0" for="requested-amount">
					<span class="label-text font-semibold">Amount</span>
				</label>

				<input
					id="requested-amount"
					type="number"
					min="1"
					step="1"
					inputmode="numeric"
					placeholder="1"
					class="input-bordered input w-full {requestedAmount !== null && !isValidAmount
						? 'input-error'
						: ''}"
					bind:value={requestedAmount}
				/>

				{#if requestedAmount !== null && !isValidAmount}
					<p class="mt-1 text-sm text-error">Enter a whole number of 1 or more.</p>
				{/if}
			</div>
		</div>

		<footer
			class="flex items-center justify-between gap-2 border-t border-base-300 bg-base-200/30 px-6 py-4"
		>
			<p class="min-w-0 truncate text-sm text-base-content/60">
				{#if selectedTreasure && isValidAmount}
					{requestedAmount} × {selectedTreasure.name}
				{/if}
			</p>

			<div class="flex gap-2">
				<button type="button" class="btn btn-ghost" onclick={onCancel}>Cancel</button>
				<button type="submit" class="btn btn-primary" disabled={!canSubmit}>Create request</button>
			</div>
		</footer>
	</form>

	<button type="button" class="modal-backdrop" aria-label="Close" tabindex="-1" onclick={onCancel}
	></button>
</dialog>
