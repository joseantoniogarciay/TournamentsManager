import assert from "node:assert/strict";
import { createRequire } from "node:module";
import test from "node:test";

const require = createRequire(import.meta.url);
const {
  configureAdaptiveWindow,
} = require("../apps/client/plugins/with-android-adaptive-window.cjs");

test("CNG permits four user-controlled orientations and multi-window without altering activity integrations", () => {
  const activity = {
    $: {
      "android:name": ".MainActivity",
      "android:screenOrientation": "portrait",
      "android:configChanges": "orientation|screenSize|screenLayout|smallestScreenSize",
      "android:windowSoftInputMode": "adjustResize",
    },
    "intent-filter": [{ action: [{ $: { "android:name": "android.intent.action.MAIN" } }] }],
  };
  const integrations = globalThis.structuredClone(activity);
  const manifest = { manifest: { application: [{ activity: [activity] }] } };
  configureAdaptiveWindow(manifest);
  assert.equal(activity.$["android:screenOrientation"], "fullUser");
  assert.equal(activity.$["android:resizeableActivity"], "true");
  assert.equal(activity.$["android:configChanges"], integrations.$["android:configChanges"]);
  assert.equal(activity.$["android:windowSoftInputMode"], "adjustResize");
  assert.deepEqual(activity["intent-filter"], integrations["intent-filter"]);
  assert.deepEqual(configureAdaptiveWindow(globalThis.structuredClone(manifest)), manifest);
});

test("a missing launch activity fails instead of silently leaving an orientation lock", () => {
  assert.throws(() => configureAdaptiveWindow({ manifest: { application: [{}] } }));
});
