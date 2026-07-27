<script lang="ts">
	// Global chrome: pull in the kit's theme tokens once, initialize the theme
	// store (reads the saved preference, follows the OS otherwise).
	import '@joyautomation/nautilus-hmi/theme.css';
	import './app.css';
	import { onMount } from 'svelte';
	import { theme } from '@joyautomation/nautilus-hmi';

	let { children } = $props();
	onMount(() => theme.init());
</script>

{@render children()}
