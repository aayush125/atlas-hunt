<script lang="ts">
	import { resolve } from '$app/paths';
	import { onMount, onDestroy } from 'svelte';
	import map1 from '$lib/assets/images/map1.png';
	import map2 from '$lib/assets/images/map2.png';
	import legendaryItem from '$lib/assets/images/legendary-item.png';

	const slides = [
		{
			src: map1,
			alt: 'Map 1',
			heading: 'Explore the World',
			sub: 'Navigate vast maps filled with hidden treasures and challenges.'
		},
		{
			src: legendaryItem,
			alt: 'Legendary Item',
			heading: 'Claim Legendary Rewards',
			sub: 'Answer questions correctly to unlock rare items.'
		},
		{
			src: map2,
			alt: 'Map 2',
			heading: 'Discover New Territories',
			sub: 'Each map brings fresh questions and untold riches.'
		}
	];

	let current = 0;
	let interval: NodeJS.Timeout | null = null;

	function prev() {
		current = (current - 1 + slides.length) % slides.length;
	}
	function next() {
		current = (current + 1) % slides.length;
	}
	function goTo(i: number) {
		current = i;
	}
	onMount(() => {
		interval = setInterval(next, 4000);
	});
	onDestroy(() => {
		if (interval) clearInterval(interval);
	});
	function resetInterval() {
		if (interval) clearInterval(interval);
		interval = setInterval(next, 4000);
	}

	const features = [
		{ icon: '⚔️', title: 'Epic Combat', desc: 'Fast-paced battles with dozens of questions.' },
		{
			icon: '🗺️',
			title: 'Big Maps',
			desc: 'Explore three different maps full of interesting treasures and questions.'
		},
		{
			icon: '🧙',
			title: 'Sharing is Caring',
			desc: 'Share treasures, collect boost points, add to your knowledge of the world.'
		}
	];
</script>

<svelte:head>
	<title>Crisis Control</title>
	<meta name="description" content="Treasure game." />
	<meta name="robots" content="noindex, nofollow" />
	<meta property="og:title" content="Crisis Control" />
	<meta property="og:description" content="Treasure game." />
	<meta property="og:type" content="website" />
</svelte:head>

<!-- Carousel -->
<div class="relative h-1/2 w-full overflow-hidden bg-black">
	<!-- Slides -->
	{#each slides as slide, i (i)}
		<img
			src={slide.src}
			alt={slide.alt}
			class="absolute inset-0 h-full w-full object-cover transition-opacity duration-700"
			style="opacity: {i === current ? 1 : 0};"
		/>
	{/each}

	<!-- Dark tint -->
	<div class="absolute inset-0 z-1 bg-black/50"></div>

	<!-- Slide text overlays -->
	{#each slides as slide, i (i)}
		<div
			class="absolute inset-0 z-2 flex flex-col items-center justify-center px-6 text-center transition-opacity duration-700"
			style="opacity: {i === current ? 1 : 0};"
		>
			<h2 class="mb-2 text-3xl font-bold text-white drop-shadow-lg sm:text-4xl">{slide.heading}</h2>
			<p class="max-w-md text-sm text-white/80 drop-shadow sm:text-base">{slide.sub}</p>
		</div>
	{/each}

	<!-- Prev / Next buttons -->
	<button
		on:click={() => {
			prev();
			resetInterval();
		}}
		class="absolute top-1/2 left-3 z-10 flex h-10 w-10 -translate-y-1/2 items-center justify-center rounded-full bg-black/50 text-xl text-white transition-colors hover:bg-black/80"
		aria-label="Previous slide">‹</button
	>
	<button
		on:click={() => {
			next();
			resetInterval();
		}}
		class="absolute top-1/2 right-3 z-10 flex h-10 w-10 -translate-y-1/2 items-center justify-center rounded-full bg-black/50 text-xl text-white transition-colors hover:bg-black/80"
		aria-label="Next slide">›</button
	>

	<!-- Dot indicators -->
	<div class="absolute bottom-4 left-1/2 z-10 flex -translate-x-1/2 gap-2">
		{#each slides as _, i (i)}
			<button
				on:click={() => {
					goTo(i);
					resetInterval();
				}}
				class="h-2.5 w-2.5 rounded-full transition-colors {i === current
					? 'bg-white'
					: 'bg-white/40 hover:bg-white/70'}"
				aria-label="Go to slide {i + 1}"
			></button>
		{/each}
	</div>

	<!-- Play button overlay -->
	<div class="absolute bottom-10 left-1/2 z-10 -translate-x-1/2">
		<a href={resolve('/play')}>
			<button class="btn btn-primary">Play</button>
		</a>
	</div>
</div>

<!-- Game Info -->
<div class="bg-base-200 px-8 py-10">
	<div class="mx-auto mb-10 max-w-3xl text-center">
		<h1 class="mb-3 text-4xl font-bold">Crisis Control</h1>
		<p class="text-lg text-base-content/70">
			Just... answer the quiz and take your treasures. That's it, really.
		</p>
		<div class="mt-5 flex flex-wrap justify-center gap-4">
			<span class="badge badge-outline">Trivia</span>
			<span class="badge badge-outline">Treasure Hunt</span>
			<span class="badge badge-outline">Multi Player</span>
		</div>
	</div>
	<div class="mx-auto grid max-w-4xl grid-cols-1 gap-6 sm:grid-cols-3">
		{#each features as f (f.title)}
			<div class="card bg-base-100 p-6 text-center shadow-md">
				<div class="mb-3 text-4xl">{f.icon}</div>
				<h3 class="mb-1 text-lg font-semibold">{f.title}</h3>
				<p class="text-sm text-base-content/60">{f.desc}</p>
			</div>
		{/each}
	</div>
</div>
