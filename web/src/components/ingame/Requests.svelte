<script lang="ts">
	import { treasuresById } from '$lib/treasures';
	import type { TreasureRequests } from '$lib/types/requests';
	import { gameStore } from '../../stores/gameStore';
	import TreasureDonateDialog from './TreasureDonateDialog.svelte';
	import TreasureRequestDialog from './TreasureRequestDialog.svelte';

	let {
		requests,
		onRequest,
		onDonate,
		onRequestDelete
	}: {
		requests: TreasureRequests;
		onRequest: (treasureId: string, amount: number) => void;
		onDonate: (treasureId: string, amount: number, donateTo: string) => void;
		onRequestDelete: () => void;
	} = $props();

	let showRequestModal = $state(false);
	let showDonateModal = $state(false);
	let donateModalData = $state<{
		treasureId: string;
		reqAmount: number;
		recAmount: number;
		donateTo: string;
	}>();
</script>

<div class="space-y-3">
	<!-- Header -->
	<div class="flex items-center justify-between">
		<div>
			<h3 class="font-semibold text-base-content">Treasure Requests</h3>
			<p class="text-xs text-base-content/50">Help other players complete their requests.</p>
		</div>

		<button class="btn btn-sm btn-primary" onclick={() => (showRequestModal = true)}>
			+ Request
		</button>
	</div>

	<!-- Requests -->
	{#if Object.keys(requests).length === 0}
		<div class="mt-8 text-center text-sm text-base-content/50">No treasure requests.</div>
	{:else}
		<div class="space-y-2">
			{#each Object.entries(requests) as [requesterId, request] (requesterId)}
				<div
					class="flex items-center justify-between gap-3 rounded-lg bg-base-200 px-3 py-2.5 transition-colors hover:bg-base-300"
				>
					<div class="min-w-0 flex-1">
						<div class="flex items-center gap-2">
							<span class="text-sm font-medium text-base-content">
								{gameStore.getPlayerNicknameFromId(requesterId)}
							</span>

							<span class="badge badge-sm badge-info">
								{request.amountReceived}/{request.requestedAmount}
							</span>
						</div>

						<div class="mt-1 text-xs text-base-content/60">
							Treasure: {#if treasuresById.get(request.treasureId)}
								{treasuresById.get(request.treasureId)?.icon}
								{treasuresById.get(request.treasureId)?.name}
							{:else}
								{request.treasureId}
							{/if}
						</div>

						<div class="flex items-center gap-2">
							<progress
								class="progress mt-1 w-full progress-primary"
								value={request.amountReceived}
								max={request.requestedAmount}
							></progress>
							<span class="text-sm font-medium"
								>{request.amountReceived}/{request.requestedAmount}</span
							>
						</div>
					</div>

					{#if requesterId !== $gameStore.playerId}
						<button
							class="btn shrink-0 btn-sm btn-primary"
							onclick={() => {
								const data = {
									treasureId: request.treasureId,
									reqAmount: request.requestedAmount,
									recAmount: request.amountReceived,
									donateTo: requesterId
								};
								donateModalData = data;
								showDonateModal = true;
							}}
						>
							Donate
						</button>
					{:else}
						<button class="btn shrink-0 btn-sm btn-warning" onclick={onRequestDelete}>
							Delete Request
						</button>
					{/if}
				</div>
			{/each}
		</div>
	{/if}
</div>

<!-- Request Modal -->
{#if showRequestModal}
	<TreasureRequestDialog
		onCancel={() => (showRequestModal = false)}
		onRequest={(treasureId: string, amount: number) => {
			onRequest(treasureId, amount);
			showRequestModal = false;
		}}
	/>
{:else if showDonateModal && donateModalData}
	<TreasureDonateDialog
		treasureId={donateModalData.treasureId}
		reqAmount={donateModalData.reqAmount}
		recAmount={donateModalData.recAmount}
		donateTo={donateModalData.donateTo}
		onCancel={() => (showDonateModal = false)}
		onDonate={(treasureId: string, amount: number, donateTo: string) => {
			onDonate(treasureId, amount, donateTo);
			showDonateModal = false;
			donateModalData = undefined;
		}}
	/>
{/if}
