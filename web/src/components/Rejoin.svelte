<script lang="ts">
	import WifiIcon from './icons/WifiIcon.svelte';

	interface ReconnectInfo {
		code: string;
		nickname: string;
		player_count: string;
	}

	let {
		reconnectInfo,
		onRejoin,
		onLeaveGame: onDismiss
	}: {
		reconnectInfo: ReconnectInfo;
		onRejoin: () => void;
		onLeaveGame: () => void;
	} = $props();

	let loggingOut = $state(false);
	let rejoining = $state(false);

	async function handleLeaveGame() {
		loggingOut = true;
		const res = await fetch('/api/game/leave', { method: 'POST' });

		if (!res.ok) {
			console.log('failed to logout');
		}

		onDismiss();
		loggingOut = false;
	}
</script>

<div class="flex w-full max-w-sm flex-col gap-6">
	<div class="flex flex-col gap-4 rounded-2xl border border-base-300 bg-base-100 p-6">
		<div class="flex items-center gap-3">
			<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-info/10">
				<WifiIcon />
			</div>
			<div>
				<p class="text-[15px] font-semibold text-base-content">Rejoin session</p>
				<p class="text-[13px] text-base-content/60">A previous game session was found</p>
			</div>
		</div>

		<div class="flex flex-col gap-2 border-t border-base-200 pt-4">
			<div class="flex items-center justify-between">
				<span class="text-xs font-bold tracking-widest text-base-content/50 uppercase"
					>Game code</span
				>
				<span
					class="rounded-lg border border-base-300 bg-base-200 px-2.5 py-0.5 text-sm font-semibold tracking-widest"
				>
					{reconnectInfo.code}
				</span>
			</div>
			<div class="flex items-center justify-between">
				<span class="text-xs font-bold tracking-widest text-base-content/50 uppercase"
					>Nickname</span
				>
				<span class="text-sm font-semibold">{reconnectInfo.nickname}</span>
			</div>
			<div class="flex items-center justify-between">
				<span class="text-xs font-bold tracking-widest text-base-content/50 uppercase"
					>Current Players</span
				>
				<span
					class="rounded-lg border border-base-300 bg-base-200 px-2.5 py-0.5 text-sm font-semibold tracking-widest"
				>
					{reconnectInfo.player_count}
				</span>
			</div>
		</div>

		<div class="flex flex-col gap-2 pt-1">
			<button class="btn w-full rounded-xl btn-primary" disabled={rejoining} onclick={onRejoin}
				>Rejoin game</button
			>
			<button
				class="btn w-full rounded-xl text-sm btn-ghost"
				onclick={handleLeaveGame}
				disabled={loggingOut}>Leave Game</button
			>
		</div>
	</div>
</div>
