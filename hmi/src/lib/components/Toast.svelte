<script lang="ts">
	// Toast viewport — mount once anywhere in your app (root layout is usual);
	// it renders whatever's queued in the `toast` store (../toast.svelte.js).
	// Severity reuses the kit's reserved StatusKind colors (see StatusPill).
	import { toast } from '../toast.svelte.js';
	import type { StatusKind } from '../types.js';
	import Icon from './Icon.svelte';

	let { position = 'bottom-right' }: { position?: 'bottom-right' | 'top-right' | 'bottom-left' | 'top-left' } =
		$props();

	const labels: Record<StatusKind, string> = {
		good: 'Success',
		warning: 'Warning',
		serious: 'Warning',
		critical: 'Error',
		off: 'Info'
	};
	const colors: Record<StatusKind, string> = {
		good: 'var(--good)',
		warning: 'var(--warn)',
		serious: 'var(--serious)',
		critical: 'var(--crit)',
		off: 'var(--muted)'
	};
</script>

<div class="viewport {position}" role="region" aria-label="Notifications">
	{#each toast.entries as t (t.id)}
		<div
			class="toast"
			style="--c: {colors[t.kind]}"
			role={t.kind === 'critical' || t.kind === 'serious' ? 'alert' : 'status'}
			aria-live={t.kind === 'critical' || t.kind === 'serious' ? 'assertive' : 'polite'}
		>
			<span class="dot" aria-hidden="true"></span>
			<div class="content">
				<span class="kind">{labels[t.kind]}</span>
				<p>{t.message}</p>
			</div>
			<button type="button" class="x" aria-label="Dismiss" onclick={() => toast.dismiss(t.id)}>
				<Icon name="x-mark" size={14} strokeWidth={2} />
			</button>
		</div>
	{/each}
</div>

<style>
	.viewport {
		position: fixed;
		z-index: 1000;
		display: flex;
		flex-direction: column;
		gap: 8px;
		padding: 16px;
		max-width: min(360px, calc(100vw - 32px));
		pointer-events: none;
	}
	.viewport.bottom-right {
		right: 0;
		bottom: 0;
		align-items: flex-end;
	}
	.viewport.top-right {
		right: 0;
		top: 0;
		align-items: flex-end;
	}
	.viewport.bottom-left {
		left: 0;
		bottom: 0;
		align-items: flex-start;
	}
	.viewport.top-left {
		left: 0;
		top: 0;
		align-items: flex-start;
	}
	.toast {
		pointer-events: auto;
		display: flex;
		align-items: flex-start;
		gap: 10px;
		width: 100%;
		padding: 10px 12px;
		background: var(--surface);
		border: 1px solid color-mix(in srgb, var(--c) 45%, var(--border));
		border-radius: var(--radius, 10px);
		box-shadow: 0 10px 28px rgba(0, 0, 0, 0.3);
		animation: slide-in 0.16s ease-out;
	}
	@keyframes slide-in {
		from {
			opacity: 0;
			transform: translateY(6px);
		}
	}
	@media (prefers-reduced-motion: reduce) {
		.toast {
			animation: none;
		}
	}
	.dot {
		flex: none;
		margin-top: 5px;
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: var(--c);
		box-shadow: 0 0 6px color-mix(in srgb, var(--c) 70%, transparent);
	}
	.content {
		flex: 1;
		min-width: 0;
	}
	.kind {
		display: block;
		font-size: var(--font-2xs);
		font-weight: 700;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		color: var(--c);
	}
	.content p {
		margin: 2px 0 0;
		font-size: var(--font-xs);
		color: var(--ink);
		word-break: break-word;
	}
	.x {
		flex: none;
		width: 20px;
		height: 20px;
		padding: 0;
		display: grid;
		place-items: center;
		border: none;
		border-radius: 5px;
		background: transparent;
		color: var(--muted);
		line-height: 1;
		cursor: pointer;
	}
	.x:hover {
		color: var(--ink);
		background: var(--hover);
	}
	.x:focus-visible {
		outline: 2px solid var(--accent);
		outline-offset: 1px;
	}
</style>
