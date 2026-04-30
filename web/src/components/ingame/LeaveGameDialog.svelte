<script lang="ts">
	let { onCancelLeave, onConfirmLeave }: { onCancelLeave: () => void; onConfirmLeave: () => void } =
		$props();
</script>

<div
	class="fixed inset-0 z-50 flex items-center justify-center bg-base-content/30 backdrop-blur-sm"
	role="dialog"
	aria-modal="true"
	tabindex="-1"
	onclick={onCancelLeave}
	onkeydown={(e) => e.key === 'Escape' && onCancelLeave()}
>
	<div
		class="card mx-2 w-full max-w-sm border border-base-300 bg-base-100 shadow-xl"
		role="dialog"
		aria-modal="true"
		tabindex="-1"
		onclick={(e) => e.stopPropagation()}
		onkeydown={(e) => e.stopPropagation()}
	>
		<div class="card-body items-center gap-3 text-center">
			<h2 class="card-title">Leave the game?</h2>
			<p class="text-sm text-base-content/70">You cannot join again on this game.</p>
			<div class="card-actions justify-center gap-2 pt-1">
				<button class="btn btn-ghost btn-sm" onclick={onCancelLeave}>Stay</button>
				<button class="btn btn-sm btn-error" onclick={onConfirmLeave}>Leave</button>
			</div>
		</div>
	</div>
</div>
