<script lang="ts">
	// Inline icon from the kit's curated Heroicons subset (see ../icons.ts —
	// MIT-licensed path data, credit Tailwind Labs). Renders as an inline SVG
	// sized in px, colored via currentColor so it always matches surrounding
	// text/buttons. Pass `name` for a registry icon, or `path`/`paths` as an
	// escape hatch for one-off icons the registry doesn't have.
	import { icons, type IconDef } from '../icons.js';

	let {
		name,
		path,
		paths,
		solid = false,
		size = 18,
		strokeWidth = 1.5,
		title,
		class: className = ''
	}: {
		/** Registry icon name, e.g. "bell" — see icons.ts for the full list. */
		name?: string;
		/** Escape hatch: a single path `d` string, used when `name` is omitted. */
		path?: string;
		/** Escape hatch: multiple path `d` strings. */
		paths?: string[];
		/** Force solid (filled) rendering for an escape-hatch path. Ignored for `name` (the registry decides). */
		solid?: boolean;
		/** Square size in px. */
		size?: number;
		/** Outline stroke width — ignored for solid icons. */
		strokeWidth?: number;
		/** Accessible name. Omit for a purely decorative icon (default aria-hidden). */
		title?: string;
		class?: string;
	} = $props();

	const def: IconDef | undefined = $derived(name ? icons[name] : undefined);
	const resolvedPaths = $derived(def?.paths ?? paths ?? (path ? [path] : []));
	const resolvedSolid = $derived(def ? (def.solid ?? false) : solid);

	$effect(() => {
		if (name && !icons[name] && !path && !paths) {
			// Missing icon: fail loud in dev rather than render a blank box.
			console.warn(`[nautilus-hmi] Icon: unknown icon name "${name}"`);
		}
	});
</script>

{#if resolvedSolid}
	<svg
		class="icon {className}"
		width={size}
		height={size}
		viewBox="0 0 24 24"
		fill="currentColor"
		aria-hidden={title ? undefined : 'true'}
		role={title ? 'img' : undefined}
	>
		{#if title}<title>{title}</title>{/if}
		{#each resolvedPaths as d (d)}
			<path fill-rule="evenodd" clip-rule="evenodd" {d} />
		{/each}
	</svg>
{:else}
	<svg
		class="icon {className}"
		width={size}
		height={size}
		viewBox="0 0 24 24"
		fill="none"
		stroke="currentColor"
		stroke-width={strokeWidth}
		stroke-linecap="round"
		stroke-linejoin="round"
		aria-hidden={title ? undefined : 'true'}
		role={title ? 'img' : undefined}
	>
		{#if title}<title>{title}</title>{/if}
		{#each resolvedPaths as d (d)}
			<path {d} />
		{/each}
	</svg>
{/if}

<style>
	.icon {
		flex: none;
		display: inline-block;
		vertical-align: middle;
	}
</style>
