<script lang="ts">
	import { resolve } from '$app/paths';
</script>

<div class="navbar bg-base-100 shadow-sm">
	<div class="flex-none">
		<label for="app-drawer" class="btn btn-square btn-ghost" aria-label="Open sidebar">
			<svg
				xmlns="http://www.w3.org/2000/svg"
				fill="none"
				viewBox="0 0 24 24"
				class="inline-block h-5 w-5 stroke-current"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M4 6h16M4 12h16M4 18h16"
				/>
			</svg>
		</label>
	</div>
	<div class="flex flex-1 justify-center">
		<a class="btn text-xl btn-ghost" href={resolve('/')}>Crisis Control</a>
	</div>
	<div class="flex-none">
		<div class="btn invisible btn-square btn-ghost" aria-hidden="true"></div>
	</div>
</div>
