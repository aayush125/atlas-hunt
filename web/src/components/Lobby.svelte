<script lang="ts">
	import { gameStore } from '../stores/gameStore';
	import avatarSrc from '$lib/assets/avatars/avatar.jpg';
</script>

<div class="flex w-full flex-col items-center justify-center gap-6 p-4 sm:p-6">
	<div class="w-full max-w-lg space-y-6 rounded-xl bg-base-100 p-6 shadow-xl sm:p-8">
		<!-- Room Code -->
		<div class="space-y-2 text-center">
			<h2 class="text-xs font-medium tracking-wide text-base-content/60 sm:text-sm">ROOM CODE</h2>

			<div
				class="rounded-lg bg-primary py-3 text-3xl font-bold tracking-widest text-primary-content shadow-md select-all sm:py-4 sm:text-4xl"
			>
				{$gameStore.joinCode}
			</div>

			<p class="text-xs text-base-content/70 sm:text-sm">
				Share this code with your friends to join the game.
			</p>
		</div>

		<!-- Players -->
		<div class="space-y-4">
			<div class="flex items-center justify-between">
				<h3 class="text-xs font-semibold tracking-wide text-base-content/80 uppercase sm:text-sm">
					Players
				</h3>
				<span class="text-xs text-base-content/60 sm:text-sm"> 4 / 8 Joined </span>
			</div>

			<div class="grid grid-cols-3 gap-4 sm:grid-cols-4 md:grid-cols-5">
				<!-- Player Card -->
				<div class="flex flex-col items-center space-y-1 sm:space-y-2">
					<img
						src={avatarSrc}
						alt="Player avatar"
						class="h-16 w-16 rounded-full object-cover shadow-md ring-2 ring-secondary sm:h-20 sm:w-20"
					/>
					<span class="text-center text-xs font-medium text-base-content sm:text-sm">
						{$gameStore.playerName}
					</span>
				</div>

				<div class="flex flex-col items-center space-y-1 sm:space-y-2">
					<img
						src={avatarSrc}
						alt="Player avatar"
						class="h-16 w-16 rounded-full object-cover shadow-md ring-2 ring-secondary sm:h-20 sm:w-20"
					/>
					<span class="text-center text-xs font-medium text-base-content sm:text-sm">CoolNick</span>
				</div>

				<div class="flex flex-col items-center space-y-1 sm:space-y-2">
					<img
						src={avatarSrc}
						alt="Player avatar"
						class="h-16 w-16 rounded-full object-cover shadow-md ring-2 ring-secondary sm:h-20 sm:w-20"
					/>
					<span class="text-center text-xs font-medium text-base-content sm:text-sm">GamerX</span>
				</div>

				<div class="flex flex-col items-center space-y-1 sm:space-y-2">
					<img
						src={avatarSrc}
						alt="Player avatar"
						class="h-16 w-16 rounded-full object-cover shadow-md ring-2 ring-secondary sm:h-20 sm:w-20"
					/>
					<span class="text-center text-xs font-medium text-base-content sm:text-sm">GamerX</span>
				</div>

				<!-- Empty Slot -->
				<div class="flex flex-col items-center space-y-1 opacity-60 sm:space-y-2">
					<div
						class="flex h-16 w-16 items-center justify-center rounded-full bg-base-300 text-lg font-bold text-base-content/40 sm:h-20 sm:w-20 sm:text-xl"
					>
						+
					</div>
					<span class="text-center text-xs text-base-content/50 sm:text-sm">Waiting...</span>
				</div>
			</div>
		</div>
	</div>

	<button class="btn mt-4 w-full max-w-lg btn-primary sm:mt-6">Start Game</button>
</div>
