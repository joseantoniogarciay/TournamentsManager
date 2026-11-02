import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { createRequire } from "node:module";
import { resolve } from "node:path";
import { runInNewContext } from "node:vm";
import test from "node:test";
import { transpileForTest } from "./helpers/transpile-typescript.mjs";

const require = createRequire(resolve("apps/client/package.json"));

function fixture() {
  let edges;
  const React = require("react");
  const modules = {
    react: {
      ...React,
      useRef: (current) => ({ current }),
      useEffect: () => {},
      useCallback: (callback) => callback,
      useState: (initial) => {
        edges = initial;
        return [
          edges,
          (next) => {
            edges = next(edges);
          },
        ];
      },
    },
    "react-native": {
      ScrollView: "scroll",
      View: "view",
      StyleSheet: { create: (s) => s },
      Platform: { OS: "web" },
      TextInput: { State: { currentlyFocusedInput: () => null } },
    },
    "react-native-safe-area-context": {
      useSafeAreaInsets: () => ({ top: 0, bottom: 0, left: 0, right: 0 }),
    },
    "expo-linear-gradient": { LinearGradient: "gradient" },
    "@tournaments-manager/design-tokens": { space: { 3: 12 } },
    "@/shared/preferences/preferences-provider": {
      usePreferences: () => ({ colors: { surface: { canvas: "#F5F1E8" } } }),
    },
  };
  function load(file) {
    const module = { exports: {} };
    runInNewContext(transpileForTest(readFileSync(file, "utf8"), { jsx: true }), {
      module,
      exports: module.exports,
      require: (name) => modules[name] ?? require(name),
    });
    return module.exports;
  }
  const shared = load("apps/client/src/shared/ui/scroll-view.tsx");
  modules["@/shared/ui/scroll-view"] = shared;
  const keyboard = load("apps/client/src/shared/ui/keyboard-aware-scroll-view.tsx");
  return { ...shared, ...keyboard, edges: () => edges };
}

for (const platform of ["web", "native"]) {
  test(`${platform}: scroll updates both fades and forwards events at start, middle and end`, () => {
    const f = fixture();
    const received = [];
    const tree = f.ScrollView({ onScroll: (e) => received.push(e) });
    const scroll = tree.props.children[0];
    for (const [offset, top, bottom] of [
      [0, 0, 1],
      [6, 0.5, 1],
      [100, 1, 1],
      [394, 1, 0.5],
      [400, 1, 0],
    ]) {
      const event = {
        nativeEvent: {
          contentOffset: { x: 0, y: offset },
          contentSize: { width: 320, height: 1000 },
          layoutMeasurement: { width: 320, height: 600 },
          ...(platform === "native" ? { contentInset: { top: 0, bottom: 0 } } : {}),
        },
      };
      scroll.props.onScroll(event);
      assert.equal(f.edges().top, top);
      assert.equal(f.edges().bottom, bottom);
      assert.equal(received.at(-1), event);
    }
    assert.equal(received.length, 5);
  });
}

test("content that fits has no fade; native adjusted top inset is respected", () => {
  const f = fixture();
  const scroll = f.ScrollView({}).props.children[0];
  scroll.props.onScroll({
    nativeEvent: {
      contentOffset: { y: 0 },
      contentSize: { height: 300 },
      layoutMeasurement: { height: 600 },
    },
  });
  assert.equal(f.edges().top, 0);
  assert.equal(f.edges().bottom, 0);
  scroll.props.onScroll({
    nativeEvent: {
      contentOffset: { y: -20 },
      contentInset: { top: 20, bottom: 0 },
      contentSize: { height: 1000 },
      layoutMeasurement: { height: 600 },
    },
  });
  assert.equal(f.edges().top, 0);
  assert.equal(f.edges().bottom, 1);
});

test("web forms: the keyboard wrapper forwards a scroll without insets through the shared fade", () => {
  const f = fixture();
  let received;
  const form = f.KeyboardAwareScrollView({
    onScroll: (event) => {
      received = event;
    },
  });
  assert.equal(form.type, f.ScrollView);
  const scroll = form.type(form.props).props.children[0];
  const event = {
    nativeEvent: {
      contentOffset: { y: 30 },
      contentSize: { height: 1000 },
      layoutMeasurement: { height: 600 },
    },
  };
  scroll.props.onScroll(event);
  assert.equal(received, event);
  assert.equal(f.edges().top, 1);
  assert.equal(f.edges().bottom, 1);
});

test("horizontal scroll forwards the original callback and ref without adding overlays", () => {
  const f = fixture();
  const ref = { current: null };
  const onScroll = () => {};
  const tree = f.ScrollView({ horizontal: true, ref, onScroll });
  assert.equal(tree.type, "scroll");
  assert.equal(tree.props.ref, ref);
  assert.equal(tree.props.onScroll, onScroll);
});
