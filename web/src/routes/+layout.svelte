<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
	import Navbar from '../components/Navbar.svelte';

	let { children } = $props();
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>

<div class="min-h-screen">
	<div class="flex h-screen flex-col">
		<Navbar />
		<main class="flex-1 overflow-auto">
			{@render children()}
		</main>
	</div>
</div>
