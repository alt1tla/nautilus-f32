// Shared menu machinery for MenuBar and DropdownMenu — one item shape, one
// outside-click action, so the two components stay siblings instead of
// diverging implementations of the same dropdown behavior.

export interface MenuItem {
	label: string;
	/** Icon name from ../icons.ts. */
	icon?: string;
	onSelect?: () => void;
	href?: string;
	disabled?: boolean;
	/** Render a divider instead of an item (label/onSelect/href ignored). */
	separator?: boolean;
	/** Display-only hint, e.g. "Ctrl+S" — not wired to a real key handler. */
	shortcut?: string;
}

export interface MenuBarMenu {
	label: string;
	items: MenuItem[];
	disabled?: boolean;
}

/**
 * Svelte action: invokes `onOutside` on any pointerdown outside `node`
 * (capture phase, so it still fires when the inner click also stops
 * propagation). Used to close an open menu/dropdown/tooltip on outside click.
 */
export function outsideClick(node: HTMLElement, onOutside: () => void) {
	function handle(e: PointerEvent) {
		if (!node.contains(e.target as Node)) onOutside();
	}
	document.addEventListener('pointerdown', handle, true);
	return {
		update(fn: () => void) {
			onOutside = fn;
		},
		destroy() {
			document.removeEventListener('pointerdown', handle, true);
		}
	};
}

/** Index of the next enabled item in `items`, wrapping, skipping separators/disabled. */
export function nextEnabledIndex(items: MenuItem[], from: number, dir: 1 | -1): number {
	const n = items.length;
	if (n === 0) return -1;
	let i = from;
	for (let step = 0; step < n; step++) {
		i = (i + dir + n) % n;
		const it = items[i];
		if (!it.separator && !it.disabled) return i;
	}
	return -1;
}
