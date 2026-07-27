<script lang="ts">
	// Device faceplate chrome: the popup a mimic click opens. Title bar with
	// the device identity, close affordances (×, Esc, backdrop), a scrolling
	// body — the content is yours (compose StatusPill / NumberField / Tabs /
	// Gauge inside). The traditional PLC-HMI shape: monitoring on the left,
	// control on the right, tabbed subviews below.
	//
	// Every direct child of the body keeps its natural height by default
	// (nothing shrinks), so a header row or <Tabs/> strip can't collapse.
	// If your content can grow tall (e.g. a tab panel), wrap just that piece
	// in `<div class="faceplate-scroll">` — it becomes the one region that
	// grows/shrinks and scrolls, while everything else above it stays put.
	// Omit it and the whole body scrolls as one region, as before.
	import type { Snippet } from 'svelte';
	import Icon from './Icon.svelte';

	let {
		title,
		subtitle = '',
		onclose,
		children
	}: {
		title: string;
		subtitle?: string;
		onclose: () => void;
		children: Snippet;
	} = $props();

	let closeBtn = $state<HTMLButtonElement | null>(null);
	$effect(() => {
		closeBtn?.focus();
	});

	function onkeydown(ev: KeyboardEvent) {
		if (ev.key === 'Escape') {
			ev.stopPropagation();
			onclose();
		}
	}
</script>

<svelte:window {onkeydown} />

<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
<div
	class="backdrop"
	onclick={(e) => {
		if (e.target === e.currentTarget) onclose();
	}}
>
	<div class="plate" role="dialog" aria-modal="true" aria-label={title}>
		<header>
			<div class="id">
				<h3>{title}</h3>
				{#if subtitle}<p>{subtitle}</p>{/if}
			</div>
			<button bind:this={closeBtn} class="x" aria-label="Close" onclick={onclose}>
				<Icon name="x-mark" size={16} strokeWidth={1.75} />
			</button>
		</header>
		<div class="body">
			{@render children()}
		</div>
	</div>
</div>

<style>
	.backdrop {
		position: fixed;
		inset: 0;
		z-index: 100;
		background: color-mix(in srgb, var(--bg, #0d1117) 55%, transparent);
		backdrop-filter: blur(2px);
		display: grid;
		place-items: center;
		padding: 24px;
	}
	.plate {
		width: min(440px, 100%);
		max-height: min(85vh, 720px);
		display: flex;
		flex-direction: column;
		background: var(--surface, #161b22);
		border: 1px solid var(--border, #21262d);
		border-radius: 12px;
		box-shadow: 0 18px 50px rgba(0, 0, 0, 0.45);
		animation: rise 0.14s ease-out;
	}
	@keyframes rise {
		from {
			opacity: 0;
			transform: translateY(8px) scale(0.985);
		}
	}
	@media (prefers-reduced-motion: reduce) {
		.plate {
			animation: none;
		}
	}
	header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 10px;
		padding: 14px 16px 10px;
		border-bottom: 1px solid var(--border, #21262d);
	}
	.id {
		min-width: 0;
	}
	h3 {
		margin: 0;
		font-size: var(--font-md);
		font-weight: 680;
		color: var(--ink, #e6edf3);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.id p {
		margin: 2px 0 0;
		font-size: var(--font-2xs);
		color: var(--muted, #8b949e);
		font-family: var(--mono, monospace);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.x {
		flex: none;
		width: 28px;
		height: 28px;
		padding: 0;
		display: grid;
		place-items: center;
		border: 1px solid var(--border, #21262d);
		border-radius: 7px;
		background: transparent;
		color: var(--muted, #8b949e);
		line-height: 1;
		cursor: pointer;
	}
	.x:hover {
		color: var(--ink, #e6edf3);
		border-color: var(--muted, #8b949e);
	}
	.x:focus-visible {
		outline: 2px solid var(--s1, #3987e5);
		outline-offset: 1px;
	}
	.body {
		padding: 14px 16px 16px;
		display: flex;
		flex-direction: column;
		gap: 14px;
		min-height: 0;
		/* Fallback for callers that don't opt a child into .faceplate-scroll
		   (below): the whole body scrolls as one region, same as before this
		   fix — nothing regresses for simple/short faceplates. */
		overflow-y: auto;
	}
	/* Every direct child of .body keeps its natural height and never shrinks
	   — this is what keeps a title/monitoring row/tab strip visible instead
	   of being squeezed toward 0 when the faceplate's content is taller than
	   the modal's max-height (see Tabs.svelte for the underlying mechanism).
	   :global() is required because these children are rendered by the
	   consuming component (different style scope), not by Faceplate itself. */
	.body > :global(*) {
		flex-shrink: 0;
	}
	/* Opt-in scroll region: wrap the part of your faceplate content that's
	   allowed to grow long (typically the active tab's panel) in a div with
	   this class, placed as a direct child of <Faceplate>. It — not the rest
	   of the body — absorbs the squeeze and becomes independently
	   scrollable, so the header, monitoring rows, and tab strip above it
	   stay pinned in view. */
	.body > :global(.faceplate-scroll) {
		flex: 1 1 auto;
		min-height: 0;
		overflow-y: auto;
	}
</style>
