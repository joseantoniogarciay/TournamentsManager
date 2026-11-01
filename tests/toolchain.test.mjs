import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { globSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join, resolve } from "node:path";
import { execPath } from "node:process";
import test from "node:test";
import { parse } from "@babel/parser";

const root = resolve(import.meta.dirname, "..");

function fixture(context) {
  const directory = mkdtempSync(join(tmpdir(), "tm-toolchain-"));
  context.after(() => rmSync(directory, { recursive: true, force: true }));
  return (name, source) => {
    const file = join(directory, name);
    writeFileSync(file, source);
    return file;
  };
}

function lint(file) {
  const result = spawnSync(
    execPath,
    [
      join(root, "node_modules/oxlint/bin/oxlint"),
      "--config",
      join(root, ".oxlintrc.json"),
      "--no-ignore",
      "--deny-warnings",
      file,
    ],
    { cwd: root, encoding: "utf8" },
  );
  assert.ifError(result.error);
  return { status: result.status, diagnostics: result.stdout + result.stderr };
}

test("lint rejects undefined JS names, unsafe TS assertions and unused assignments", (context) => {
  const write = fixture(context);
  for (const [name, source, rules] of [
    ["undefined.mjs", "export const value = unknownValue;", ["no-undef"]],
    ["unsafe.ts", "export const value: any = null!;", ["no-explicit-any", "no-non-null-assertion"]],
    [
      "assignment.mjs",
      "export function run() { let value = 1; value = 2; return value; }",
      ["no-useless-assignment"],
    ],
  ]) {
    const result = lint(write(name, source));
    assert.equal(result.status, 1, result.diagnostics);
    for (const rule of rules) assert.ok(result.diagnostics.includes(rule), result.diagnostics);
  }
  const safe = lint(write("safe.ts", "export const value: number = 1;"));
  assert.equal(safe.status, 0, safe.diagnostics);
});

test("CommonJS keeps config loading while strict parsing rejects legacy octal and duplicate parameters", (context) => {
  const write = fixture(context);
  const safe = lint(write("config.cjs", '"use strict"; module.exports = require("node:path");'));
  assert.equal(safe.status, 0, safe.diagnostics);
  for (const [name, source] of [
    ["octal.cjs", '"use strict"; module.exports = 012;'],
    ["duplicate.cjs", '"use strict"; module.exports = function(a, a) { return a; };'],
  ]) {
    const result = lint(write(name, source));
    assert.equal(result.status, 1, result.diagnostics);
    assert.ok(result.diagnostics.length > 0);
  }
  // Oxlint has no separate no-octal rule: all authored CommonJS must opt into
  // strict parsing so that this diagnostic cannot disappear on a new config.
  for (const file of globSync("**/*.cjs", {
    cwd: root,
    exclude: [
      "**/node_modules/**",
      ".git/**",
      ".tmp/**",
      "apps/client/android/**",
      "apps/client/ios/**",
    ],
  })) {
    const ast = parse(readFileSync(join(root, file), "utf8"), { sourceType: "commonjs" });
    assert.ok(
      ast.program.directives.some((directive) => directive.value.value === "use strict"),
      file,
    );
  }
});

test("the compiler rejects unchecked indexes and explicit undefined optional properties", (context) => {
  const write = fixture(context);
  write(
    "invalid.ts",
    "const values: string[] = []; export const value: string = values[0];\n" +
      "interface Options { value?: string }; export const options: Options = { value: undefined };",
  );
  const project = write(
    "tsconfig.json",
    JSON.stringify({
      extends: join(root, "tsconfig.base.json"),
      compilerOptions: { types: [], target: "ES2022" },
      files: ["invalid.ts"],
    }),
  );
  const result = spawnSync(
    execPath,
    [join(root, "node_modules/typescript/bin/tsc"), "--project", project, "--pretty", "false"],
    { cwd: root, encoding: "utf8" },
  );
  assert.ifError(result.error);
  assert.notEqual(result.status, 0);
  assert.match(result.stdout + result.stderr, /TS2322/);
  assert.match(result.stdout + result.stderr, /TS2375/);
});
