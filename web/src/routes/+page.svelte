<script>
	import { resolve } from '$app/paths';
</script>

<div class="relative h-1/2 w-full border border-red-500">
	Home Page Graphics here
	<div class="absolute bottom-10 left-1/2 -translate-x-1/2 transform">
		<a href={resolve('/play')}>
			<button class="btn btn-primary">Play</button>
		</a>
	</div>
</div>
<div>Game info here</div>
