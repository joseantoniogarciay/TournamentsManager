import assert from "node:assert/strict";
import { createRequire } from "node:module";
import test from "node:test";

const require = createRequire(import.meta.url);
const { migrateAppDelegate } = require("../apps/client/plugins/with-ios-scene-lifecycle.cjs");
const template = `class AppDelegate: ExpoAppDelegate {
  var window: UIWindow?
  var reactNativeFactory: RCTReactNativeFactory?
  func application() {
    window = UIWindow(frame: UIScreen.main.bounds)
    factory.startReactNative(
      withModuleName: "main",
      in: window,
      launchOptions: launchOptions)
  }
  // Existing Expo linking and subscriber integration must survive prebuild.
  func existingIntegration() {}
}`;

test("regeneration preserves integrations and never creates a second scene delegate", () => {
  const once = migrateAppDelegate(template);
  assert.equal(migrateAppDelegate(once), once);
  assert.equal(once.match(/class SceneDelegate:/g)?.length, 1);
  assert.ok(once.includes("func existingIntegration() {}"));
  assert.ok(!once.includes("UIWindow(frame:"));
  assert.ok(once.includes("UIWindow(windowScene: windowScene)"));
});

test("a changed Expo startup template requires review instead of a partial migration", () => {
  assert.throws(
    () => migrateAppDelegate(template.replace('withModuleName: "main"', 'withModuleName: "other"')),
    /Unsupported Expo AppDelegate/,
  );
  assert.throws(() => migrateAppDelegate("class AppDelegate {}"), /Unsupported Expo AppDelegate/);
});
