<script lang="ts">
	import avatarSrc from '$lib/assets/avatars/avatar.jpg';
</script>

<div class="flex w-full flex-col items-center justify-center gap-6 p-4 sm:p-6">
	<div class="w-full max-w-lg rounded-xl bg-base-100 p-8 shadow-xl">
		<!-- Loading Visual -->
		<div class="flex flex-col items-center justify-center gap-6">
			<!-- Animated Avatar Stack -->
			<div class="flex -space-x-4">
				{#each Array(3) as i (i)}
					<div class="animate-bounce" style="animation-delay: {i * 0.15}s">
						<img
							src={avatarSrc}
							alt="avatar"
							class="h-16 w-16 rounded-full object-cover shadow-md ring-2 ring-primary sm:h-20 sm:w-20"
						/>
					</div>
				{/each}
			</div>

			<!-- Loading Text -->
			<div class="space-y-2 text-center">
				<h2 class="text-xl font-bold text-base-content sm:text-2xl">Loading...</h2>

				<!-- Animated Dots -->
				<span class="loading loading-lg loading-dots text-primary"></span>
			</div>
		</div>
	</div>

	<!-- Subtle hint -->
	<p class="text-xs text-base-content/50 sm:text-sm">Setting things up for you</p>
</div>
