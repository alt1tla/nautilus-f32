// The kit built-ins' default connection points — split out of registry.ts
// (which otherwise pulls in the real Svelte components) so pure logic that
// only needs the ports data (ports.ts, and its ad-hoc test) can import this
// without dragging in a Svelte runtime. registry.ts re-exports PORTS from
// here so nothing else needs to change its import.
//
// The DATA itself now lives in hmi/src/lib/mimic.ts's BUILTIN_PORTS — the
// runtime (<Mimic>, via resolveRuntimePorts()) needs this same table to
// resolve a pipe anchor's port against a built-in component with no
// instance-level `ports` override, so it can't stay editor-only anymore.
// This module just re-exports it under the name existing imports expect.
import { BUILTIN_PORTS } from '@joyautomation/nautilus-hmi/mimic';
import type { Port } from './portsGestures';

export const PORTS: Record<string, Port[]> = BUILTIN_PORTS;
