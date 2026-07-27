// Tests for the user-components build harness. `customComponentNames` is
// pure and covered directly; `buildUserComponentsBundle` is exercised for
// real against the worked example (examples/hmi-demo/src/lib/
// HeatExchanger.svelte, which is exactly the fixture the end-to-end
// verification also compiles) — it uses that project's OWN node_modules
// (svelte + esbuild's TS transpile of its `<script lang="ts">`), proving the
// harness works against a real, unmodified user project, not a fixture
// built to fit the compiler.

import { strict as assert } from "node:assert";
import { test } from "node:test";
import * as fs from "node:fs";
import * as path from "node:path";
import * as os from "node:os";
import { buildUserComponentsBundle, customComponentNames } from "./userComponentBuild";

test("customComponentNames excludes built-ins, dedupes, and sorts", () => {
  const builtins = new Set(["Tank", "Pump", "Valve", "Gauge", "Sparkline"]);
  assert.deepEqual(
    customComponentNames(["Tank", "HeatExchanger", "Pump", "HeatExchanger", "Widget"], builtins),
    ["HeatExchanger", "Widget"]
  );
  assert.deepEqual(customComponentNames(["Tank", "Pump"], builtins), []);
  assert.deepEqual(customComponentNames([], builtins), []);
});

test("customComponentNames ignores empty/falsy names", () => {
  assert.deepEqual(customComponentNames(["", "Foo"], new Set()), ["Foo"]);
});

const HEAT_EXCHANGER = path.join(__dirname, "..", "..", "..", "examples", "hmi-demo", "src", "lib", "HeatExchanger.svelte");
const HAS_FIXTURE = fs.existsSync(HEAT_EXCHANGER) && fs.existsSync(path.join(path.dirname(HEAT_EXCHANGER), "..", "..", "node_modules", "svelte"));

test(
  "buildUserComponentsBundle compiles the HeatExchanger demo with its own project's svelte, TS script and all",
  { skip: !HAS_FIXTURE ? "examples/hmi-demo isn't checked out / npm installed here" : false },
  async () => {
    const outfile = path.join(await fs.promises.mkdtemp(path.join(os.tmpdir(), "nx-user-components-")), "user-components.js");
    const result = await buildUserComponentsBundle({ HeatExchanger: HEAT_EXCHANGER }, outfile);
    assert.deepEqual(result.diagnostics, []);
    assert.deepEqual(result.built, ["HeatExchanger"]);
    const code = await fs.promises.readFile(outfile, "utf8");
    assert.match(code, /__NX_USER_COMPONENTS__/);
    assert.match(code, /HeatExchanger/);
    // The compiled output is plain JS — no TypeScript syntax should have
    // survived the <script lang="ts"> preprocessing step.
    assert.doesNotMatch(code, /: \s*\{\s*active\?: boolean/);
  }
);

test("buildUserComponentsBundle still emits a bundle when a component is missing entirely", async () => {
  const outfile = path.join(await fs.promises.mkdtemp(path.join(os.tmpdir(), "nx-user-components-")), "user-components.js");
  const result = await buildUserComponentsBundle({ Nope: path.join(os.tmpdir(), "does-not-exist", "Nope.svelte") }, outfile);
  assert.equal(result.built.length, 0);
  assert.equal(result.diagnostics.length, 1);
  assert.equal(result.diagnostics[0].component, "Nope");
  const code = await fs.promises.readFile(outfile, "utf8");
  assert.match(code, /__NX_USER_COMPONENTS__/);
});
