import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { runInNewContext } from "node:vm";
import test from "node:test";
import ts from "typescript";

const path = "apps/client/src/app/_layout.tsx";
const source = ts.createSourceFile(path, readFileSync(path, "utf8"), ts.ScriptTarget.Latest, true);
const navigator = source.statements.find(
  (node) => ts.isFunctionDeclaration(node) && node.name?.text === "SessionNavigator",
);
assert.ok(navigator);
const code = ts.transpileModule(
  `${navigator.getText(source)}\nexports.render = SessionNavigator;`,
  {
    compilerOptions: {
      module: ts.ModuleKind.CommonJS,
      target: ts.ScriptTarget.ES2022,
      jsx: ts.JsxEmit.ReactJSX,
    },
  },
).outputText;

function fixture(platform, transition = "resetting", canDismiss = true) {
  const events = [];
  const frames = new Map();
  let nextFrame = 0;
  let cleanup;
  const module = { exports: {} };
  const context = {
    exports: module.exports,
    require: () => ({
      jsx: (type, props, key) => ({ type, props, key }),
      jsxs: (type, props, key) => ({ type, props, key }),
    }),
    Platform: { OS: platform },
    Stack: { Screen: "screen" },
    typography: { family: { semibold: "font" } },
    useSession: () => ({
      revision: 1,
      transition,
      replacementDestination: "/join-team",
      finishSessionReplacement: () => events.push("finished"),
    }),
    useEffect: (effect) => {
      cleanup = effect();
    },
    router: {
      canDismiss: () => canDismiss,
      dismissAll: () => events.push("dismiss"),
      replace: (destination) => events.push(destination),
    },
    requestAnimationFrame: (callback) => {
      frames.set(++nextFrame, callback);
      return nextFrame;
    },
    cancelAnimationFrame: (id) => frames.delete(id),
  };
  runInNewContext(code, context, { filename: path });
  const tree = module.exports.render();
  const advance = () => {
    const pending = [...frames.values()];
    frames.clear();
    for (const callback of pending) callback();
  };
  return { events, frames, advance, tree, cleanup: () => cleanup?.() };
}

for (const platform of ["ios", "web", "android"]) {
  test(`${platform}: clears previous modal history before restoring invitation and finishes once`, () => {
    const f = fixture(platform);
    assert.equal(f.tree.key, undefined, "session changes must preserve the native root container");
    if (platform === "android") assert.deepEqual(f.events, []);
    for (let i = 0; i < 5; i++) f.advance();
    assert.deepEqual(f.events, ["dismiss", "/join-team", "finished"]);
  });
  test(`${platform}: logout targets account and idle session never navigates`, () => {
    const f = fixture(platform, "signing-out", false);
    for (let i = 0; i < 5; i++) f.advance();
    assert.deepEqual(f.events, ["/account", "finished"]);
    const idle = fixture(platform, "idle");
    idle.advance();
    assert.deepEqual(idle.events, []);
  });
}

for (const boundary of [0, 1, 2, 3]) {
  test(`android: cleanup after ${boundary} frames prevents stale navigation or completion`, () => {
    const f = fixture("android");
    for (let i = 0; i < boundary; i++) f.advance();
    const previous = [...f.events];
    f.cleanup();
    for (let i = 0; i < 5; i++) f.advance();
    assert.deepEqual(f.events, previous);
    assert.equal(f.frames.size, 0);
    assert.ok(!f.events.includes("finished"));
  });
}
