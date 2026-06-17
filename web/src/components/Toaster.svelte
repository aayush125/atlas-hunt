<script lang="ts">
	import { fade } from 'svelte/transition';
	import { toasts } from '../stores/toast';
</script>

<div class="toast toast-end toast-bottom z-50">
	{#each $toasts as t (t.id)}
		<div class="alert alert-{t.type} transition-all duration-300" out:fade={{ duration: 300 }}>
			<span>{t.message}</span>
		</div>
	{/each}
</div>
