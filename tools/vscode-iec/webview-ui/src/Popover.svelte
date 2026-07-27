<script lang="ts">
	// The floating card under the toolbar (palette, vars panel, …): one home
	// for the chrome and the click-swallow that keeps the window-level
	// "click anywhere closes" handler from eating clicks inside the card.
	import type { Snippet } from 'svelte';

	let {
		style = '',
		scroll = false,
		onkeydown,
		children
	}: {
		style?: string;
		// Opt-in: overflow clips absolutely-positioned children (suggest
		// dropdowns), so only long-content panels should scroll.
		scroll?: boolean;
		onkeydown?: (ev: KeyboardEvent) => void;
		children: Snippet;
	} = $props();
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="popover" class:scroll {style} onclick={(e) => e.stopPropagation()} {onkeydown}>
	{@render children()}
</div>

<style>
	.popover {
		position: absolute;
		right: 8px;
		top: 30px;
		z-index: 20;
		display: flex;
		flex-direction: column;
		min-width: 250px;
		background: var(--nx-panel-bg);
		border: 1px solid var(--nx-border);
		border-radius: 5px;
		padding: 4px;
		box-shadow: var(--nx-shadow);
	}
	.popover.scroll {
		max-height: 70vh;
		overflow-y: auto;
	}
</style>
