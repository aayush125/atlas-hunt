<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
	import Navbar from '../components/Navbar.svelte';
	import { resolve } from '$app/paths';

	let { children } = $props();
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>

<div class="drawer">
	<input id="app-drawer" type="checkbox" class="drawer-toggle" />

	<div class="drawer-content flex min-h-screen flex-col">
		<Navbar />
		<main class="flex-1">
			{@render children()}
		</main>
	</div>

	<div class="drawer-side">
		<label for="app-drawer" class="drawer-overlay"></label>

		<ul class="menu min-h-full w-80 bg-base-200 p-4">
			<li><a href={resolve('/')}>Home</a></li>
			<li><a href={resolve('/')}>Socials</a></li>
			<li><a href={resolve('/test')}>Test</a></li>
		</ul>
	</div>
</div>
