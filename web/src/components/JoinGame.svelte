<script lang="ts">
	import type { Game } from '$lib/types/game';
	import type { Player } from '$lib/types/player';

	let {
		onCreateGame,
		onGameJoined
	}: { onCreateGame: () => void; onGameJoined: (player: Player, game: Game) => void } = $props();

	let gameCode = $state('');
	let nickname = $state('');

	async function joinGame() {
		const res = await fetch('/api/game/join', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				nickname: nickname,
				game_code: gameCode
			})
		});

		if (res.ok) {
			const json = await res.json();
			const player = json.player;
			const game = json.game;
			onGameJoined(
				{
					...player,
					join_timestamp: new Date(player.join_timestamp),
					last_rejoin_timestamp: new Date(player.last_rejoin_timestamp)
				},
				{ ...game, created_at: new Date(game.created_at) }
			);
		} else {
			console.error('Error joining game:', await res.json());
		}
	}
</script>

<div class="flex w-full max-w-sm flex-col gap-6">
	<div class="flex flex-col gap-2">
		<label for="game-code" class="text-sm font-bold tracking-widest text-gray-700 uppercase">
			Game Code
		</label>
		<input
			bind:value={gameCode}
			oninput={(e) => (gameCode = e.currentTarget.value.toUpperCase().slice(0, 6))}
			id="game-code"
			type="text"
			placeholder="ENTER GAME CODE"
			class="input w-full rounded-xl border-2 px-4 py-3 text-lg font-semibold tracking-widest input-neutral placeholder:text-gray-300 focus:ring-2 focus:ring-offset-2 focus:outline-none"
		/>
	</div>

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
	<button class="btn btn-primary" onclick={joinGame}>Join</button>
</div>

<div class="h-px w-full bg-linear-to-r from-transparent via-gray-300 to-transparent"></div>

<div class="flex w-full max-w-sm flex-col gap-6">
	<button class="btn btn-secondary" onclick={onCreateGame}>Create a Game</button>
</div>
