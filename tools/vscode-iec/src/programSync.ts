// Pure helpers for the online-edit feature (onlineEdit.ts) — vscode-free so
// they run under plain node:test (programSync.test.ts). The composition rule
// mirrors internal/stproject: sibling .st libraries (sorted by name) precede
// the program body, and the same prelude joins every program in a
// multi-program project.

/** The `PROGRAM <Name>` POU name of IEC source, "" if none. Programs on a
 * multi-task controller are addressed by this name. */
export function pouOf(src: string): string {
  const m = /^\s*PROGRAM\s+([A-Za-z_][A-Za-z0-9_]*)/m.exec(src);
  return m ? m[1] : "";
}

/** Recover the program body from composed source given the prelude Join
 * placed ahead of it — the TypeScript mirror of stproject.SplitProgram.
 * Returns undefined when the prelude isn't a prefix (the controller's
 * libraries don't match this project). */
export function splitProgram(composed: string, prelude: string): string | undefined {
  if (composed.startsWith(prelude)) return composed.slice(prelude.length);
  const trimmed = prelude.replace(/\n+$/, "");
  if (trimmed !== prelude && composed.startsWith(trimmed)) {
    return composed.slice(trimmed.length).replace(/^\n/, "");
  }
  return undefined;
}

/** The library prefix of a controller's composed source — what that task
 * believes the shared .st libraries say. A deployed source carries the
 * workspace prelude verbatim; an online edit may have rewritten it, so fall
 * back to peeling the task's known program body off the end, then to cutting
 * at the `PROGRAM` line every IEC language opens its program file with. A
 * source that defeats all three (fully diverged) is returned whole — the
 * diff then shows everything that task runs, which is the honest picture. */
export function controllerPrelude(source: string, wsPrelude: string, wsBody?: string): string {
  if (splitProgram(source, wsPrelude) !== undefined) return wsPrelude;
  if (wsBody) {
    const trimmed = wsBody.replace(/\n+$/, "");
    if (source.endsWith(wsBody)) return source.slice(0, source.length - wsBody.length);
    if (trimmed && source.endsWith(trimmed)) return source.slice(0, source.length - trimmed.length);
  }
  const m = /^[ \t]*PROGRAM\b/m.exec(source);
  if (m) return source.slice(0, m.index);
  return source;
}

/** Whitespace-insensitive comparison: embed order and blank lines differ
 * between a binary's embed composition and the editor's, but the logic
 * doesn't. */
export function normalize(src: string): string {
  return src
    .replace(/\r/g, "")
    .split("\n")
    .map((l) => l.replace(/\s+$/g, ""))
    .filter((l) => l.length > 0)
    .join("\n");
}
