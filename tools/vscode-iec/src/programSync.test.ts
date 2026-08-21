// Plain-Node tests for the online-edit composition helpers (no vscode
// dependency). Run via `npm test` (compiles then executes with node:test).

import { test } from "node:test";
import * as assert from "node:assert/strict";
import { controllerPrelude, normalize, pouOf, splitProgram } from "./programSync";

const PRELUDE = "FUNCTION_BLOCK RateOfChange\nVAR_INPUT IN : REAL; END_VAR\nEND_FUNCTION_BLOCK\n";
const BODY = "PROGRAM Main\nVAR x : REAL; END_VAR\nx := 1.0;\nEND_PROGRAM\n";

test("pouOf finds the PROGRAM name, in any language's program file", () => {
  assert.equal(pouOf(BODY), "Main");
  assert.equal(pouOf("(* comment *)\n  PROGRAM Interlocks\n"), "Interlocks");
  assert.equal(pouOf(PRELUDE), "");
});

test("splitProgram inverts Join for a known prelude", () => {
  assert.equal(splitProgram(PRELUDE + BODY, PRELUDE), BODY);
  // Tolerates a trailing-newline difference at the seam.
  assert.equal(splitProgram(PRELUDE.replace(/\n$/, "") + "\n" + BODY, PRELUDE), BODY);
  assert.equal(splitProgram("something else entirely\n" + BODY, PRELUDE), undefined);
});

test("controllerPrelude: deployed source carries the workspace prelude", () => {
  assert.equal(controllerPrelude(PRELUDE + BODY, PRELUDE, BODY), PRELUDE);
});

test("controllerPrelude: an online-edited prelude is peeled off the known body", () => {
  const edited = PRELUDE.replace("REAL", "LREAL");
  assert.equal(controllerPrelude(edited + BODY, PRELUDE, BODY), edited);
});

test("controllerPrelude: both halves edited — cut at the PROGRAM line", () => {
  const edited = PRELUDE.replace("REAL", "LREAL");
  const editedBody = BODY.replace("1.0", "2.0");
  assert.equal(controllerPrelude(edited + editedBody, PRELUDE, BODY), edited);
});

test("controllerPrelude: unsplittable source is returned whole", () => {
  const src = '{"nodes": []}';
  assert.equal(controllerPrelude(src, PRELUDE, BODY), src);
});

test("normalize ignores blank lines and trailing whitespace", () => {
  assert.equal(normalize("a  \n\nb\r\n"), normalize("a\nb"));
  assert.notEqual(normalize("a\nb"), normalize("a\nc"));
});
