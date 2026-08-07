<script lang="ts">
	import './layout.css';
	import Navbar from '../components/Navbar.svelte';
	import Toaster from '../components/Toaster.svelte';

	let { children } = $props();
</script>

<svelte:head><link rel="icon" href="/favicon.svg" /></svelte:head>

<div class="min-h-screen">
	<div class="flex h-screen flex-col">
		<Navbar />
		<main class="flex-1 overflow-auto">
			{@render children()}
		</main>
		<Toaster />
	</div>
</div>
