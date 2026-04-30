<script lang="ts">
	import thirdMap from '$lib/assets/maps/round_3.svg';
	let svg: SVGSVGElement;
	let coords = 'Click on the SVG to get coordinates';

	type game_coords = {
		x: number;
		y: number;
	};

	let coordsSet: game_coords[] = [];

	async function handleClick(event: MouseEvent) {
		const pt = svg.createSVGPoint();
		pt.x = event.clientX;
		pt.y = event.clientY;

		const svgPt = pt.matrixTransform(svg.getScreenCTM()!.inverse());
		const coordObj = { x: parseFloat(svgPt.x.toFixed(2)), y: parseFloat(svgPt.y.toFixed(2)) };

		coordsSet = [...coordsSet, coordObj]; // reactive reassignment
		coords = `SVG coords: x=${coordObj.x}, y=${coordObj.y}`;
	}
</script>

<p>{coords}</p>
<button on:click={() => console.log(JSON.stringify(coordsSet))}>Print coords</button>
<button on:click={() => (coordsSet = [])}>Clear markers</button>

<svg
	bind:this={svg}
	on:click={handleClick}
	xmlns="http://www.w3.org/2000/svg"
	version="1.1"
	viewBox="0 0 1386 756"
	width="1386pt"
	height="756pt"
	xmlns:xlink="http://www.w3.org/1999/xlink"
	style="border:1px solid black; cursor: crosshair;"
>
	<image href={thirdMap} width="100%" height="100%" preserveAspectRatio="xMidYMid meet" />

	{#each coordsSet as coord, i (i)}
		<text
			x={coord.x}
			y={coord.y}
			font-size="24"
			text-anchor="middle"
			dominant-baseline="middle"
			style="pointer-events: none; user-select: none;">📍</text
		>
		<text
			x={coord.x + 16}
			y={coord.y - 16}
			font-size="11"
			fill="white"
			stroke="black"
			stroke-width="2"
			paint-order="stroke"
			text-anchor="middle"
			style="pointer-events: none; user-select: none;"
		>
			{i + 1}
		</text>
	{/each}
</svg>
