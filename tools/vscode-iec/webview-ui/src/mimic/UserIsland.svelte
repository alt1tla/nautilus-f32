<script lang="ts">
	// One user-authored component, mounted as an isolated island via
	// window.__NX_USER_COMPONENTS__ (userComponents.ts's esbuild bundle,
	// compiled with the USER's own node_modules). Falls back to the same
	// dashed placeholder chip the editor has always shown for an unknown
	// component when there's no bundle yet, the name isn't in it, or its
	// compile failed — the failure's message becomes the chip's tooltip.
	//
	// The mount/update/destroy handle (registry.ts's mount() wraps a $state
	// props object) is the seam: creating it hands over the DOM node once,
	// and every later prop change flows through handle.update() rather than
	// remounting — same "gestures mutate, the host confirms" shape as the
	// rest of the editor, just one level down.
	import { hasUserComponent, userComponentApi, userComponentDiagnostic } from './userComponentsState.svelte';

	let { name, props = {} }: { name: string; props?: Record<string, unknown> } = $props();

	let el: HTMLDivElement | null = $state(null);
	let handle: { update(p: Record<string, unknown>): void; destroy(): void } | null = null;

	const available = $derived(hasUserComponent(name));
	const diagnostic = $derived(userComponentDiagnostic(name));

	$effect(() => {
		if (!available || !el) return;
		const api = userComponentApi(name);
		if (!api) return;
		handle = api.mount(el, { ...props });
		return () => {
			handle?.destroy();
			handle = null;
		};
	});

	// A separate effect so a props-only change updates the live instance
	// instead of tearing it down and remounting.
	$effect(() => {
		const p = props;
		if (handle) handle.update({ ...p });
	});
</script>

{#if available}
	<div class="nx-user-island" bind:this={el}></div>
{:else}
	<span class="unknown" title={diagnostic ?? `no compiled component for ${name}`}>{name}?</span>
{/if}

<style>
	.nx-user-island {
		display: contents;
	}
	.unknown {
		display: inline-block;
		padding: 4px 8px;
		border: 1px dashed var(--nx-err);
		border-radius: 6px;
		color: var(--nx-err);
		font-size: 12px;
		font-family: var(--nx-mono);
	}
</style>
