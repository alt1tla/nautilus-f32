// Shared state for the standalone *.component.json editor (ComponentApp.svelte)
// + the op seam to the extension host (componentEditor.ts). Every gesture
// ends in postComponentPortsOp() with the FULL ports list; the host patches
// just the `ports` key of the entry (any other, future-metadata keys pass
// through untouched) and rewrites the whole document as one WorkspaceEdit —
// same "the webview never owns the document" shape as the mimic editor's
// postOp. Today this editor only shows/edits `ports`, but the state is
// named for the file (a general per-component metadata sidecar), not the
// one thing it currently edits.
import { vscode } from '../vscodeApi';
import type { Port } from './portsGestures';

export type { Port };
export type ComponentDoc = { component: string; ports: Port[] };

export const cs = $state({
	doc: null as ComponentDoc | null,
	title: 'component',
	/** Parse error from the host — the JSON is being hand-edited mid-keystroke. */
	error: ''
});

export function postComponentPortsOp(ports: Port[]): void {
	vscode.postMessage({ type: 'componentOp', ports });
}

/** Tell the host the message listener is up — it answers with the current
 * doc. Without this handshake the one-shot componentDoc post can land
 * before the app's listener exists (same race the mimic editor guards against). */
export function announceComponentReady(): void {
	vscode.postMessage({ type: 'componentReady' });
}
