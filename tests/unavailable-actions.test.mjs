import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { createRequire } from "node:module";
import { resolve } from "node:path";
import { runInNewContext } from "node:vm";
import test from "node:test";
import { transpileForTest } from "./helpers/transpile-typescript.mjs";

const clientRequire = createRequire(resolve("apps/client/package.json"));
const React = clientRequire("react");
const { renderToStaticMarkup } = clientRequire("react-dom/server");
const web = clientRequire("react-native-web");
const tokenModule = { exports: {} };
runInNewContext(transpileForTest(readFileSync("packages/design-tokens/src/tokens.ts", "utf8")), {
  module: tokenModule,
  exports: tokenModule.exports,
});
const tokens = tokenModule.exports;

function fixture(platform = "web", theme = "light") {
  const messages = [];
  const colors = tokens.themeColors[theme];
  const modules = {
    "react-native": { ...web, Platform: { OS: platform } },
    "@tournaments-manager/design-tokens": tokens,
    "@/shared/i18n/locale": { getTranslator: () => (key) => key },
    "@/shared/feedback/feedback-provider": {
      useFeedback: () => ({ show: (value) => messages.push(value) }),
    },
    "@/shared/preferences/preferences-provider": {
      usePreferences: () => ({ colors, resolvedTheme: theme }),
    },
    "./text": { Text: web.Text },
    "expo-linear-gradient": { LinearGradient: "gradient" },
  };
  function load(file) {
    const module = { exports: {} };
    runInNewContext(transpileForTest(readFileSync(file, "utf8"), { jsx: true }), {
      module,
      exports: module.exports,
      require: (name) => modules[name] ?? clientRequire(name),
    });
    return module.exports;
  }
  const { ActionPressable } = load("apps/client/src/shared/ui/action-pressable.tsx");
  modules["./action-pressable"] = { ActionPressable };
  const { Button } = load("apps/client/src/shared/ui/button.tsx");
  const { ConfigurationOption } = load("apps/client/src/shared/ui/configuration-option.tsx");
  const pressable = (props) => {
    const tree = ActionPressable(props);
    return platform === "web" ? tree.props.children : tree;
  };
  return { ActionPressable, pressable, Button, ConfigurationOption, messages, colors };
}

for (const platform of ["web", "ios", "android"]) {
  test(`${platform}: unavailable activation explains the current reason without invoking the action`, () => {
    const f = fixture(platform);
    let calls = 0;
    const tree = f.pressable({
      disabled: true,
      disabledReason: "Add two teams first.",
      onPress: () => calls++,
    });
    assert.equal(tree.props.disabled, undefined, "the explanation must remain reachable");
    assert.equal(tree.props.accessibilityState.disabled, true);
    assert.equal(tree.props.accessibilityHint, "Add two teams first.");
    tree.props.onPress({});
    tree.props.onPress({});
    assert.equal(calls, 0);
    assert.equal(f.messages.length, 2);
    assert.equal(f.messages[0].message, "Add two teams first.");
    const enabled = f.pressable({ disabled: false, onPress: () => calls++ });
    enabled.props.onPress({});
    assert.equal(calls, 1, "an action becomes available again without stale blocking");
  });
  test(`${platform}: busy activation cannot repeat an operation or invoke a long press`, () => {
    const f = fixture(platform);
    let calls = 0;
    const tree = f.pressable({
      accessibilityState: { busy: true },
      onPress: () => calls++,
      onLongPress: () => calls++,
    });
    tree.props.onPress({});
    assert.equal(tree.props.onLongPress, undefined);
    assert.equal(calls, 0);
    assert.equal(f.messages[0].message, "common_action_in_progress");
  });
}

test("web: unavailable control inherits aria-disabled and stays keyboard-focusable without HTML disabled", () => {
  const { ActionPressable } = fixture();
  const html = renderToStaticMarkup(
    React.createElement(
      ActionPressable,
      {
        accessibilityRole: "button",
        disabled: true,
        disabledReason: "Add two teams first.",
      },
      React.createElement(web.Text, null, "Start tournament"),
    ),
  );
  assert.match(html, /aria-disabled="true"/);
  assert.match(html, /tabindex="0"/);
  assert.doesNotMatch(html, /(?:\sdisabled=|pointer-events:none)/);
});

for (const theme of ["light", "dark"]) {
  for (const variant of ["primary", "secondary", "ghost", "destructive"]) {
    test(`${theme}/${variant}: blocked button has a solid surface, readable label and opacity 1`, () => {
      const f = fixture("web", theme);
      const button = f.Button({
        variant,
        label: "Start",
        disabled: true,
        disabledReason: "Add two teams first.",
        onPress: () => assert.fail("business action"),
      });
      const surface = web.StyleSheet.flatten(button.props.children.props.style);
      assert.match(surface.backgroundColor, /^#[0-9a-f]{6}$/i);
      assert.equal(surface.backgroundColor, f.colors.surface.subtle);
      assert.equal(button.props.children.props.children.props.children[1].props.color, "secondary");
      const pressable = f.pressable(button.props);
      const style = web.StyleSheet.flatten(pressable.props.style({ pressed: true }));
      assert.equal(style.opacity, 1);
      assert.match(style.backgroundColor, /^#[0-9a-f]{6}$/i);
      pressable.props.onPress({});
      assert.equal(f.messages[0].message, "Add two teams first.");
    });
  }
  test(`${theme}: selected configuration stays opaque while blocked and explains the reason`, () => {
    const f = fixture("web", theme);
    const option = f.ConfigurationOption({
      label: "League",
      selected: true,
      disabled: true,
      disabledReason: "Saving.",
      onPress: () => assert.fail("configuration change"),
    });
    const style = web.StyleSheet.flatten(option.props.style);
    assert.equal(style.backgroundColor, f.colors.surface.subtle);
    f.pressable(option.props).props.onPress({});
    assert.equal(f.messages[0].message, "Saving.");
  });
}
