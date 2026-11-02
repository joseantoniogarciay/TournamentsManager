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
    "react-native": { ScrollView: "scroll", View: "view", StyleSheet: { create: (s) => s } },
    "expo-linear-gradient": { LinearGradient: "gradient" },
    "@tournaments-manager/design-tokens": { space: { 3: 12 } },
    "@/shared/preferences/preferences-provider": {
      usePreferences: () => ({ colors: { surface: { canvas: "#F5F1E8" } } }),
    },
  };
  const module = { exports: {} };
  runInNewContext(
    transpileForTest(readFileSync("apps/client/src/shared/ui/scroll-view.tsx", "utf8"), {
      jsx: true,
    }),
    {
      module,
      exports: module.exports,
      require: (name) => modules[name] ?? require(name),
    },
  );
  return { ScrollView: module.exports.ScrollView, edges: () => edges };
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
