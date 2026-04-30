<script lang="ts">
	import { onMount } from 'svelte';
	import type { Treasure } from '$lib/types/treasure';
	import * as Panzoom from '@panzoom/panzoom';
	// import firstMap from '$lib/assets/maps/round_1.svg';
	// import secondMap from '$lib/assets/maps/round_2.svg';
	import thirdMap from '$lib/assets/maps/round_3.svg';
	import { SvelteSet } from 'svelte/reactivity';
	import { thirdMapPoints as validPoints } from '$lib/points';
	import TreasureDialog from './TreasureDialog.svelte';
	import { gameStore } from '../../stores/gameStore';

	let svgEl: SVGSVGElement;

	const usedIndices = new SvelteSet<number>();
	let icons: Treasure[] = $state(
		Array.from({ length: Math.min(50, validPoints.length) }, (_, i) => {
			let idx;
			do {
				idx = Math.floor(Math.random() * validPoints.length);
			} while (usedIndices.has(idx));
			usedIndices.add(idx);
			return {
				id: String(i),
				name: 'randomname',
				icon: '💎',
				x: validPoints[idx].x,
				y: validPoints[idx].y,
				question: {
					text: 'What are the answers to this question?',
					options: ['Answer One', 'Answer Two', 'Answer Three', 'Answer Four']
				}
			};
		})
	);

	let selectedTreasure = $state<Treasure | null>(null);
	let claimResult = $state<'correct' | 'incorrect' | null>(null);
	let claimResultTimer: ReturnType<typeof setTimeout>;

	function handleIconClick(icon: Treasure) {
		console.log('Clicked icon', icon.id, 'at', icon.x, icon.y);
		selectedTreasure = icon;
	}

	function handleTreasureClaim(treasure: Treasure, selectedIndex: number) {
		console.log('selected answer:', selectedIndex);
		claimResult = 'correct';
		gameStore.addTreasure(treasure);
		icons = icons.filter((ic) => ic.id !== treasure.id);

		clearTimeout(claimResultTimer);
		claimResultTimer = setTimeout(() => {
			selectedTreasure = null;
			claimResult = null;
		}, 1000);
	}

	onMount(() => {
		const instance = Panzoom.default(svgEl, {
			minScale: 0.5,
			maxScale: 6,
			contain: 'outside',
			cursor: 'default'
		});
		svgEl.parentElement?.addEventListener('wheel', instance.zoomWithWheel);
		return () => {
			instance.destroy();
			svgEl.parentElement?.removeEventListener('wheel', instance.zoomWithWheel);
		};
	});
</script>

<div class="flex h-full w-full items-center justify-center">
	<div>
		<svg
			xmlns="http://www.w3.org/2000/svg"
			version="1.1"
			viewBox="0 0 1386 756"
			width="800"
			height="435"
			xmlns:xlink="http://www.w3.org/1999/xlink"
			bind:this={svgEl}
		>
			<image href={thirdMap} width="100%" height="100%" preserveAspectRatio="xMidYMid meet" />
			<defs>
				<radialGradient id="iconGradient" cx="40%" cy="35%" r="60%">
					<stop offset="0%" stop-color="#f05050" />
					<stop offset="100%" stop-color="#c0202a" />
				</radialGradient>
			</defs>
			{#each icons as icon (icon.id)}
				<g transform="translate({icon.x}, {icon.y})">
					<g
						role="button"
						tabindex="0"
						on:click={() => handleIconClick(icon)}
						on:keydown={(e) => e.key === 'Enter' && handleIconClick(icon)}
						class="map-icon"
					>
						<circle
							r="16"
							fill="none"
							stroke="rgba(220,50,50,0.35)"
							stroke-width="2"
							class="pulse-ring"
						/>
						<circle r="14" fill="rgba(0,0,0,0.18)" transform="translate(0, 2)" />
						<circle r="13" fill="url(#iconGradient)" stroke="white" stroke-width="2" />
						<text
							x="0"
							y="0"
							text-anchor="middle"
							dominant-baseline="central"
							dy="0.05em"
							font-size="14"
							style="user-select: none; pointer-events: none;">{icon.icon}</text
						>
					</g>
				</g>
			{/each}
		</svg>
	</div>
</div>

{#if selectedTreasure}
	<TreasureDialog
		treasure={selectedTreasure}
		onClose={() => {
			selectedTreasure = null;
			claimResult = null;
		}}
		onTreasureClaim={handleTreasureClaim}
		{claimResult}
	/>
{/if}

<style>
	.map-icon {
		cursor: pointer;
		transform-box: fill-box;
		transform-origin: center;
		transition: transform 0.15s ease;
	}

	.map-icon:hover {
		transform: scale(1.2);
	}

	.pulse-ring {
		animation: pulse 1.5s ease-out infinite;
		transform-origin: center;
	}

	@keyframes pulse {
		0% {
			r: 13;
			opacity: 0.6;
		}
		70% {
			r: 22;
			opacity: 0;
		}
		100% {
			r: 22;
			opacity: 0;
		}
	}
</style>
