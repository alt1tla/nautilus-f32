<script lang="ts" module>
	import type { DriverState } from '../types.js';

	// One place maps a driver state to its color, label, and dot behavior —
	// reused by the badge and the card. Error/waiting states pulse; healthy
	// ones glow steady. Never color alone: every state has a label and a
	// distinct dot.
	export const STATE_META: Record<
		DriverState,
		{ label: string; color: string; pulse: boolean }
	> = {
		connected: { label: 'Connected', color: 'var(--good)', pulse: false },
		connecting: { label: 'Connecting', color: 'var(--warn)', pulse: true },
		waiting: { label: 'Waiting', color: 'var(--warn)', pulse: true },
		degraded: { label: 'Degraded', color: 'var(--serious)', pulse: true },
		error: { label: 'Error', color: 'var(--crit)', pulse: true },
		offline: { label: 'Offline', color: 'var(--muted)', pulse: false }
	};
</script>

<script lang="ts">
	let {
		state,
		label,
		size = 'md'
	}: { state: DriverState; label?: string; size?: 'sm' | 'md' } = $props();

	const meta = $derived(STATE_META[state] ?? STATE_META.offline);
</script>

<span
	class="badge {size}"
	class:pulse={meta.pulse}
	style="--c: {meta.color}"
	role="status"
	aria-label={label ?? meta.label}
>
	<span class="dot" aria-hidden="true"></span>
	{label ?? meta.label}
</span>

<style>
	.badge {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		padding: 3px 10px;
		border: 1px solid color-mix(in srgb, var(--c) 45%, var(--border));
		border-radius: 999px;
		font-size: var(--font-2xs);
		font-weight: 600;
		letter-spacing: 0.02em;
		color: var(--ink-2);
		background: color-mix(in srgb, var(--c) 10%, var(--surface-2));
		white-space: nowrap;
	}
	.badge.sm {
		padding: 1px 7px;
		font-size: var(--font-2xs);
		gap: 5px;
	}
	.dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: var(--c);
		box-shadow: 0 0 6px color-mix(in srgb, var(--c) 70%, transparent);
		flex: none;
	}
	.badge.sm .dot {
		width: 7px;
		height: 7px;
	}
	.pulse .dot {
		animation: pulse 1.4s ease-in-out infinite;
	}
	@keyframes pulse {
		0%,
		100% {
			opacity: 1;
			box-shadow: 0 0 6px color-mix(in srgb, var(--c) 70%, transparent);
		}
		50% {
			opacity: 0.45;
			box-shadow: 0 0 2px color-mix(in srgb, var(--c) 30%, transparent);
		}
	}
	/* Respect reduced-motion: no pulsing dot. */
	@media (prefers-reduced-motion: reduce) {
		.pulse .dot {
			animation: none;
		}
	}
</style>
