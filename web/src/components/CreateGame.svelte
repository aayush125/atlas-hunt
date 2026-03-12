<script lang="ts">
	import type { Game } from '$lib/types/game';
	import type { Player } from '$lib/types/player';

	let {
		onJoinGame,
		onGameCreated
	}: {
		onJoinGame: () => void;
		onGameCreated: (game: Game, player: Player) => void;
	} = $props();

	let numberOfPlayers = $state(6);
	let nickname = $state('');

	async function createGame() {
		const res = await fetch('/api/game/create', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				nickname: nickname,
				number_of_players: numberOfPlayers
			})
		});

		if (res.ok) {
			const json = await res.json();
			const game = json.game;
			const player = json.player;
			onGameCreated(
				{ ...game, created_at: new Date(game.created_at) },
				{
					...player,
					join_timestamp: new Date(player.join_timestamp),
					last_rejoin_timestamp: new Date(player.last_rejoin_timestamp)
				}
			);
		} else {
			console.error('Error creating game:', await res.json());
		}
	}
</script>

<div class="flex w-full max-w-sm flex-col gap-6">
	<div class="flex flex-col gap-2">
		<label for="nickname" class="text-sm font-bold tracking-widest text-gray-700 uppercase">
			Nickname
		</label>
		<input
			maxlength="20"
			bind:value={nickname}
			oninput={(e) => (nickname = e.currentTarget.value.slice(0, 20))}
			id="nickname"
			type="text"
			placeholder="ENTER A NICKNAME"
			class="input w-full rounded-xl border-2 px-4 py-3 text-lg font-semibold tracking-wide input-neutral placeholder:text-gray-300 focus:ring-2 focus:ring-offset-2 focus:outline-none"
		/>
	</div>

	<div class="flex flex-col gap-2">
		<label for="num-players" class="text-sm font-bold tracking-widest text-gray-700 uppercase">
			Number of Players
		</label>
		<div class="w-full">
			<input
				bind:value={numberOfPlayers}
				type="range"
				min="4"
				max="10"
				step="2"
				class="range w-full"
			/>
			<div class="mt-2 flex justify-between text-xs">
				{#each [4, 6, 8, 10] as n (n)}
					<div class="flex flex-col items-center gap-0.5">
						<span>|</span>
						<span>{n}</span>
					</div>
				{/each}
			</div>
		</div>
	</div>
	<button class="btn btn-primary" onclick={createGame}>Create</button>
</div>

<div class="h-px w-full bg-linear-to-r from-transparent via-gray-300 to-transparent"></div>

<div class="flex w-full max-w-sm flex-col gap-6">
	<button class="btn btn-secondary" onclick={onJoinGame}>Join a Game</button>
</div>
