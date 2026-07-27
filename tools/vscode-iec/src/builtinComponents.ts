// Host-side mirror of webview-ui/src/mimic/builtinPorts.ts, which is the
// SOURCE OF TRUTH — the extension host can't import that module (or
// registry.ts) directly since it pulls in the real Svelte components via
// @joyautomation/nautilus-hmi, a browser-only dependency chain the host
// bundle has no business loading. This file is deliberately tiny (name +
// default ports only) and kept in manual sync with the webview copy;
// changing a built-in's default ports means editing BOTH files.
//
// Used by the "Edit Component Ports…" command (editComponentPorts.ts) for
// its QuickPick list and to prefill a brand-new sidecar with the built-in's
// current defaults.
import type { Port } from "./mimicComponentIndex";

export const BUILTIN_COMPONENT_PORTS: Record<string, Port[]> = {
  Tank: [
    { name: "top", x: 0.5, y: 0 },
    { name: "left", x: 0, y: 0.5 },
    { name: "right", x: 1, y: 0.5 },
    { name: "bottom", x: 0.5, y: 1 },
  ],
  Pump: [
    { name: "in", x: 0, y: 0.5 },
    { name: "out", x: 1, y: 0.5 },
  ],
  Valve: [
    { name: "in", x: 0, y: 0.5 },
    { name: "out", x: 1, y: 0.5 },
  ],
  Gauge: [],
  Sparkline: [],
};

export const BUILTIN_COMPONENT_NAMES = Object.keys(BUILTIN_COMPONENT_PORTS);
