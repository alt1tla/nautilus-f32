// Plain-Node tests for the identifier scanner (no vscode dependency).
// Run via `npm test` (compiles then executes with node:test).

import { test } from "node:test";
import * as assert from "node:assert/strict";
import { scanIdentifiers, formatValue, formatValueHover } from "./scan";

const values = new Map<string, unknown>([
  ["levelpct", 59.887482],
  ["pumprun", true],
  ["tempsp", 65],
]);

test("finds identifiers case-insensitively", () => {
  const sites = scanIdentifiers("IF levelPct < 40.0 THEN PumpRun := TRUE; END_IF;", values);
  assert.deepEqual(
    sites.map((s) => s.path.toLowerCase()),
    ["levelpct", "pumprun"]
  );
});

test("end offset lands just past the identifier", () => {
  const text = "TempSP := 65.0;";
  const [site] = scanIdentifiers(text, values);
  assert.equal(text.slice(0, site.end), "TempSP");
});

test("skips line comments, block comments, and strings", () => {
  const text = [
    "// LevelPct in a line comment",
    "(* PumpRun in a block comment *)",
    "msg := 'LevelPct in a string';",
    'msg2 := "TempSP too";',
    "LevelPct := 1.0;",
  ].join("\n");
  const sites = scanIdentifiers(text, values);
  assert.deepEqual(
    sites.map((s) => s.path.toLowerCase()),
    ["levelpct"]
  );
});

test("does not match partial words", () => {
  const sites = scanIdentifiers("LevelPctX := LevelPct2;", values);
  assert.equal(sites.length, 0);
});

test("formatValue renders compactly", () => {
  assert.equal(formatValue(59.887482), "59.887");
  assert.equal(formatValue(65), "65");
  assert.equal(formatValue(true), "TRUE");
  assert.equal(formatValue(false), "FALSE");
  assert.equal(formatValue(null), "—");
  assert.equal(formatValue(1e9), "1000000000"); // integers never go exponential
  assert.equal(formatValue(1234567.89), "1.23e+6");
});

test("formatValue renders compound values as size hints", () => {
  assert.equal(formatValue({ AI: -4, ALMDLY: 30 }), "{…}");
  assert.equal(formatValue([1, 2, 3, 4]), "[4]");
  assert.equal(formatValue([]), "[0]");
});

test("formatValueHover expands structs TypeScript-style", () => {
  const timer = { PRE: 1200000, ACC: 0, DN: false };
  assert.equal(
    formatValueHover(timer),
    "{\n  PRE: 1200000\n  ACC: 0\n  DN: FALSE\n}"
  );
  assert.equal(
    formatValueHover({ Header: { Valid: true }, Count: 12 }),
    "{\n  Header: {\n    Valid: TRUE\n  }\n  Count: 12\n}"
  );
  assert.equal(formatValueHover(65.5), "65.500");
});

test("formatValueHover elides long arrays and deep floods", () => {
  const arr = Array.from({ length: 14 }, (_, i) => i);
  const out = formatValueHover(arr);
  assert.ok(out.includes("[0]: 0"));
  assert.ok(out.includes("… (4 more elements)"));
  // A wide struct truncates by line count instead of flooding the tooltip.
  const wide: Record<string, number> = {};
  for (let i = 0; i < 60; i++) wide["m" + i] = i;
  const wideOut = formatValueHover(wide);
  assert.ok(wideOut.split("\n").length <= 41);
  assert.ok(wideOut.includes("more lines"));
});

test("resolves member and index accessors to the child value", () => {
  const vals = new Map<string, unknown>([
    ["rtu", { VALUE: -25.019, Header: { Displacement: 3.5 } }],
    ["plt", [{ Count: 10 }, { Count: 20 }]],
  ]);
  const sites = (text: string) => scanIdentifiers(text, vals);

  // A member reference shows the child value, and the chip lands past .VALUE.
  const [m] = sites("X := RTU.VALUE;");
  assert.equal(m.value, -25.019);
  assert.equal(m.path, "RTU.VALUE");

  // Nested member.
  assert.equal(sites("X := RTU.Header.Displacement;")[0].value, 3.5);

  // Case-insensitive member key.
  assert.equal(sites("X := rtu.value;")[0].value, -25.019);

  // A bare struct reference stays the whole object (hover expands it).
  const [whole] = sites("RTU : Analog;");
  assert.equal(typeof whole.value, "object");
  assert.equal(whole.path, "RTU");

  // Array index + member.
  assert.equal(sites("X := Plt[1].Count;")[0].value, 20);

  // An unknown member stops the walk at the deepest resolved value.
  const [partial] = sites("X := RTU.Nope;");
  assert.equal(typeof partial.value, "object");
  assert.equal(partial.path, "RTU");
});

test("FUNCTION_BLOCK regions and instance scoping", () => {
  const src = [
    "(* libs *)",
    "FUNCTION_BLOCK RateOfChange",
    "VAR_INPUT IN : REAL; DT : REAL; END_VAR",
    "VAR_OUTPUT OUT : REAL; END_VAR",
    "VAR prev : REAL; END_VAR",
    "OUT := (IN - prev) / DT;",
    "prev := IN;",
    "END_FUNCTION_BLOCK",
    "",
    "FUNCTION_BLOCK Other",
    "VAR x : REAL; END_VAR",
    "x := 1.0;",
    "END_FUNCTION_BLOCK",
  ].join("\n");
  const { scanFbRegions, scanInstanceDecls, instanceScope } = require("./scan");

  const regions = scanFbRegions(src);
  assert.deepEqual(
    regions.map((r: { type: string }) => r.type),
    ["RateOfChange", "Other"]
  );
  // The body slice starts below the header and covers the block.
  const body = src.slice(regions[0].bodyStart, regions[0].end);
  assert.ok(body.includes("prev := IN"));
  assert.ok(!body.includes("FUNCTION_BLOCK RateOfChange"));
  assert.equal(regions[0].headerLine, 1);

  // Instance declarations across languages: ST VAR blocks and FBD calls.
  const decls = [
    "PROGRAM Reports",
    "VAR r0 : RateOfChange; END_VAR",
    "FBD",
    "  r1 : RateOfChange(IN := TempC, DT := RepDtS)",
    "END_FBD",
    "END_PROGRAM",
  ].join("\n");
  assert.deepEqual(scanInstanceDecls(decls, "RateOfChange"), ["r0", "r1"]);
  // The FUNCTION_BLOCK header itself is not an instance.
  assert.deepEqual(scanInstanceDecls(src, "RateOfChange"), []);

  // Scoping: bare members resolve through the instance struct.
  const vals = new Map<string, unknown>([
    ["r1", { IN: 54.2, DT: 1.0, OUT: -0.4, prev: 54.6 }],
    ["tempc", 54.2],
  ]);
  const scoped = instanceScope(vals, "r1")!;
  const sites = scanIdentifiers("OUT := (IN - prev) / DT;", scoped);
  assert.deepEqual(
    sites.map((s) => [s.path, s.value]),
    [
      ["OUT", -0.4],
      ["IN", 54.2],
      ["prev", 54.6],
      ["DT", 1.0],
    ]
  );
  // A non-struct value scopes to nothing.
  assert.equal(instanceScope(vals, "tempc"), undefined);
});
