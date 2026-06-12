<script lang="ts">
	import type { Member } from '$lib/types/player';
	import avatarSrc from '$lib/assets/avatars/avatar.jpg';

	export let players: Member[] = [];
</script>

<div class="space-y-1">
	{#if players.length === 0}
		<div class="mt-8 text-center text-sm text-base-content/50">No players to show.</div>
	{:else}
		{#each players as player (player.id)}
			<div
				class="flex items-center gap-3 rounded-lg px-2 py-1.5 transition-colors hover:bg-base-200"
			>
				<img
					src={avatarSrc}
					alt="Player avatar"
					class="h-10 w-10 rounded-full object-cover shadow-md ring-2 ring-secondary sm:h-10 sm:w-10"
				/>
				<span class="text-sm font-medium text-base-content">{player.nickname}</span>
			</div>
		{/each}
	{/if}
</div>
