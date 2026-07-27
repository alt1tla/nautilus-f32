// Resolved component ports: the extension host aggregates every
// `{ComponentName}.component.json` sidecar in the workspace (see
// src/mimicComponentIndex.ts + src/mimicComponents.ts) into this same
// components map and pushes it as the `mimicManifest` message's `components`
// field — the webview only ever sees the resolved map, unchanged in shape from when it
// came from one central mimic.components.json, and asks for edits via
// postManifestOp ("the host applies it, the resolved map flows back", same
// shape as mimicOp).
import type { MimicEquipment } from '@joyautomation/nautilus-hmi';
// .ts extension kept on this one specifier (unlike its sibling imports
// elsewhere) so ports.test.ts can run this module directly with
// `node --experimental-strip-types --test` — Node's ESM loader needs an
// explicit extension; Vite/esbuild resolve it identically either way.
import { PORTS } from './builtinPorts.ts';
import type { Port } from './portsGestures';

export type { Port };

export type ComponentManifestEntry = {
	/** Named fractions of the rendered box. */
	ports?: Port[];
};
export type ComponentsManifest = Record<string, ComponentManifestEntry>;

/** Resolve an equipment instance's connection points (named fractions of
 * its rendered box), in precedence order: instance override (`eq.ports`) ->
 * the project's sidecar entry for its component -> built-in default
 * (registry's PORTS table). Only `undefined` falls through to the next
 * tier — an explicit `[]` at any tier means "no ports" and wins outright,
 * which is why every check below is presence-based rather than
 * length-based. */
export function resolvePorts(
	eq: Pick<MimicEquipment, 'component' | 'ports'>,
	manifest: ComponentsManifest | null
): Port[] {
	if (eq.ports !== undefined) return eq.ports;
	const entry = manifest?.[eq.component];
	if (entry?.ports !== undefined) return entry.ports;
	return PORTS[eq.component] ?? [];
}
