<script>
	import InfoIcon from '../icons/InfoIcon.svelte';
	import LeaveIcon from '../icons/LeaveIcon.svelte';
	import GameInfoDialog from './GameInfoDialog.svelte';
	import LeaveGameDialog from './LeaveGameDialog.svelte';
	import Map from './Map.svelte';

	let showInfoModal = false;
	let showLeaveConfirm = false;
	let activeTab = 'inventory';

	function openInfo() {
		showInfoModal = true;
	}
	function closeInfo() {
		showInfoModal = false;
	}
	function promptLeave() {
		showLeaveConfirm = true;
	}
	function cancelLeave() {
		showLeaveConfirm = false;
	}
	function confirmLeave() {
		showLeaveConfirm = false;
		// TODO: wire up leave game logic
	}
</script>

<div class="flex h-full w-full flex-col">
	<!-- MAP -->
	<div class="relative min-h-0 flex-1">
		<Map />
		<!-- FLOATING BUTTONS (top-right corner) -->
		<div class="absolute top-3 right-3 z-10 flex gap-2">
			<button
				class="btn gap-1.5 shadow-md btn-sm btn-info"
				onclick={openInfo}
				aria-label="Game info"
			>
				<InfoIcon />
				Info
			</button>
			<button
				class="btn gap-1.5 shadow-md btn-sm btn-error"
				onclick={promptLeave}
				aria-label="Leave game"
			>
				<LeaveIcon />
				Leave
			</button>
		</div>
	</div>

	<!-- TABBED SECTION -->
	<div class="flex flex-col border-t border-base-300 bg-base-100">
		<!-- Tab headers -->
		<div role="tablist" class="tabs-border tabs px-4 pt-2">
			<button
				role="tab"
				class="tab {activeTab === 'inventory' ? 'tab-active' : ''}"
				onclick={() => (activeTab = 'inventory')}
			>
				Inventory
			</button>
			<button
				role="tab"
				class="tab {activeTab === 'players' ? 'tab-active' : ''}"
				onclick={() => (activeTab = 'players')}
			>
				Players
			</button>
		</div>

		<!-- Tab content -->
		<div class="h-48 overflow-y-auto p-4">
			{#if activeTab === 'inventory'}
				<div class="mt-8 text-center text-sm text-base-content/50">No items in inventory yet.</div>
			{:else}
				<div class="mt-8 text-center text-sm text-base-content/50">No players to show.</div>
			{/if}
		</div>
	</div>

	<!-- INFO MODAL -->
	{#if showInfoModal}
		<GameInfoDialog onClose={closeInfo} />
	{/if}

	<!-- LEAVE CONFIRM MODAL -->
	{#if showLeaveConfirm}
		<LeaveGameDialog onCancelLeave={cancelLeave} onConfirmLeave={confirmLeave} />
	{/if}
</div>
