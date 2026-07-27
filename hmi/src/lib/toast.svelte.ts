// Toast notification queue — a store-based API (mirrors theme.svelte.ts /
// motion.svelte.ts) plus the <Toast/> component that renders it. Mount
// <Toast/> once (e.g. in your root layout, alongside ThemeSwitch's host
// chrome); call `toast.show(...)` (or the success/error/etc. sugar) from
// anywhere.
import type { StatusKind } from './types.js';

export interface ToastEntry {
	id: number;
	message: string;
	/** Reuses the kit's reserved status roles for severity styling. */
	kind: StatusKind;
	/** ms before auto-dismiss; 0 = sticky (dismiss button only). */
	duration: number;
}

let nextId = 1;

function createToastQueue() {
	let entries = $state<ToastEntry[]>([]);
	const timers = new Map<number, ReturnType<typeof setTimeout>>();

	function dismiss(id: number) {
		entries = entries.filter((t) => t.id !== id);
		const t = timers.get(id);
		if (t) {
			clearTimeout(t);
			timers.delete(id);
		}
	}

	function show(opts: { message: string; kind?: StatusKind; duration?: number }): number {
		const id = nextId++;
		const duration = opts.duration ?? 4500;
		entries = [...entries, { id, message: opts.message, kind: opts.kind ?? 'good', duration }];
		if (duration > 0) {
			timers.set(
				id,
				setTimeout(() => dismiss(id), duration)
			);
		}
		return id;
	}

	return {
		get entries() {
			return entries;
		},
		show,
		dismiss,
		success: (message: string, duration?: number) => show({ message, kind: 'good', duration }),
		error: (message: string, duration?: number) => show({ message, kind: 'critical', duration }),
		warning: (message: string, duration?: number) => show({ message, kind: 'warning', duration }),
		info: (message: string, duration?: number) => show({ message, kind: 'off', duration }),
		clear() {
			for (const t of timers.values()) clearTimeout(t);
			timers.clear();
			entries = [];
		}
	};
}

export const toast = createToastQueue();
