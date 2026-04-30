<script lang="ts">
	import CloseIcon from '../icons/CloseIcon.svelte';

	let { onClose }: { onClose: () => void } = $props();
</script>

<div
	class="fixed inset-0 z-50 flex items-center justify-center bg-base-content/30 backdrop-blur-sm"
	role="dialog"
	aria-modal="true"
	tabindex="-1"
	onclick={onClose}
	onkeydown={(e) => e.key === 'Escape' && onClose()}
>
	<div
		class="card mx-2 w-full max-w-md border border-base-300 bg-base-100 shadow-xl"
		onclick={(e) => e.stopPropagation()}
		onkeydown={(e) => e.stopPropagation()}
		tabindex="-1"
		role="dialog"
		aria-modal="true"
		aria-label="How to Play"
	>
		<div class="card-body gap-3">
			<div class="flex items-start justify-between">
				<h2 class="card-title text-lg">How to Play</h2>
				<button class="btn btn-square btn-ghost btn-sm" onclick={onClose} aria-label="Close">
					<CloseIcon />
				</button>
			</div>

			<div class="text-sm leading-relaxed text-base-content/80">
				<p>
					Tap treasures, answer questions, and earn points. Spend your collection on power-ups to
					score even faster — and if you're missing a piece, ask other players to trade. Just keep
					an eye on your inventory: donate or discard treasures to make room for what matters.
				</p>
			</div>

			<div class="card-actions justify-end pt-1">
				<button class="btn btn-sm btn-primary" onclick={onClose}>Got it!</button>
			</div>
		</div>
	</div>
</div>
