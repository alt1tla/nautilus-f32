// Reactive bridge to window.__NX_USER_COMPONENTS__ — the IIFE the extension
// host builds from the user's OWN Svelte components (src/userComponents.ts +
// src/userComponentBuild.ts) and injects as a second <script> tag AFTER the
// editor bundle's. Because that tag necessarily loads (and runs) later, the
// registry isn't there yet at first paint: this module's `tick` bumps on the
// entry's "ready" event (dispatched once window.__NX_USER_COMPONENTS__ is
// set) so every UserIsland re-checks availability instead of being stuck on
// the fallback chip forever. Compile diagnostics arrive as an ordinary
// postMessage (userComponentDiagnostics) and surface as the chip's tooltip.
export type UserComponentHandle = { update(props: Record<string, unknown>): void; destroy(): void };
type UserComponentApi = { mount(target: Element, props: Record<string, unknown>): UserComponentHandle };

declare global {
	interface Window {
		__NX_USER_COMPONENTS__?: Record<string, UserComponentApi>;
	}
}

export const ucState = $state({ tick: 0, diagnostics: {} as Record<string, string> });

if (typeof window !== 'undefined') {
	window.addEventListener('nx:user-components-ready', () => {
		ucState.tick++;
	});
}

/** Reactive: true once `name` has a real user-authored island to mount
 * (reads `ucState.tick` so callers inside `$derived`/`$effect` re-run when
 * a (re)built bundle lands). */
export function hasUserComponent(name: string): boolean {
	ucState.tick;
	return typeof window !== 'undefined' && !!window.__NX_USER_COMPONENTS__?.[name];
}

export function userComponentApi(name: string): UserComponentApi | undefined {
	return window.__NX_USER_COMPONENTS__?.[name];
}

/** The host's last reported compile failure for `name`, if any — shown as
 * the fallback chip's tooltip. */
export function userComponentDiagnostic(name: string): string | undefined {
	return ucState.diagnostics[name];
}

/** Apply a `userComponentDiagnostics` message's payload. */
export function setUserComponentDiagnostics(diagnostics: { component: string; message: string }[]): void {
	ucState.diagnostics = Object.fromEntries(diagnostics.map((d) => [d.component, d.message]));
}
