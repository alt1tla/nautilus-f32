<script lang="ts">
	// A responsive grid of driver-status cards — the whole field-connectivity
	// picture in one panel. Feed it `frame.drivers` from the nautilus stream
	// (or GET /api/drivers). A live clock ticks so "up 12s" and "3s ago"
	// advance even between frames.
	import DriverStatusCard from './DriverStatusCard.svelte';
	import type { DriverStatus } from '../types.js';

	let { drivers = [] }: { drivers?: DriverStatus[] } = $props();

	let now = $state(Date.now());
	$effect(() => {
		const id = setInterval(() => (now = Date.now()), 1000);
		return () => clearInterval(id);
	});
</script>

{#if drivers.length}
	<div class="grid">
		{#each drivers as d (d.kind + ':' + d.name)}
			<DriverStatusCard driver={d} {now} />
		{/each}
	</div>
{:else}
	<p class="empty">No field drivers reporting — this controller runs on the loopback driver.</p>
{/if}

<style>
	.grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		gap: 12px;
	}
	.empty {
		margin: 0;
		font-size: var(--font-xs);
		color: var(--muted);
	}
</style>
