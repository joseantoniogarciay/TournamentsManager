import assert from "node:assert/strict";
import { mkdtemp, mkdir, readFile, rm, writeFile } from "node:fs/promises";
import { createRequire } from "node:module";
import { tmpdir } from "node:os";
import { dirname, join } from "node:path";
import test from "node:test";
import { fileURLToPath } from "node:url";

const root = dirname(dirname(fileURLToPath(import.meta.url)));
const clientRequire = createRequire(join(root, "apps/client/package.json"));
const { AndroidConfig, XML } = clientRequire("expo/config-plugins");
const splashRequire = createRequire(clientRequire.resolve("expo-splash-screen/package.json"));
const { withAndroidSplashStyles } = splashRequire("./plugin/build/withAndroidSplashStyles.js");
const { setAndroidIntentFilters } = AndroidConfig.IntentFilters;
const filterConfig = (host) => ({
  android: {
    intentFilters: [
      {
        action: "VIEW",
        autoVerify: true,
        category: ["BROWSABLE", "DEFAULT"],
        data: [{ scheme: "https", host, pathPrefix: "/link/" }],
      },
    ],
  },
});
const manifest = (filters) => ({
  manifest: {
    $: { "xmlns:android": "http://schemas.android.com/apk/res/android" },
    application: [
      {
        activity: [
          {
            $: { "android:name": ".MainActivity" },
            ...(filters ? { "intent-filter": filters } : {}),
          },
        ],
      },
    ],
  },
});
const activity = (value) => AndroidConfig.Manifest.getMainActivityOrThrow(value);

test("App Links migrates legacy ownership, preserves manual filters and regenerates without duplicates", async (context) => {
  const manual = {
    $: { "android:autoVerify": "true" },
    data: [{ $: { "android:host": "manual.example" } }],
  };
  const directory = await mkdtemp(join(tmpdir(), "tm-app-links-"));
  context.after(() => rm(directory, { recursive: true, force: true }));
  const path = join(directory, "AndroidManifest.xml");
  let input = manifest([manual, { $: { "data-generated": "true" } }]);
  setAndroidIntentFilters(filterConfig("first.example"), input);
  const once = globalThis.structuredClone(input);
  setAndroidIntentFilters(filterConfig("first.example"), input);
  assert.deepEqual(input, once);
  assert.equal(activity(input)["intent-filter"].length, 2);
  assert.deepEqual(activity(input)["intent-filter"][0], manual);
  const generated = activity(input)["intent-filter"][1];
  assert.equal(generated.$["data-generated"], "true");
  assert.equal(generated.$["android:autoVerify"], "true");
  assert.deepEqual(
    generated.category.map((value) => value.$["android:name"]),
    ["android.intent.category.BROWSABLE", "android.intent.category.DEFAULT"],
  );
  assert.deepEqual(generated.action, [{ $: { "android:name": "android.intent.action.VIEW" } }]);
  assert.deepEqual(generated.data, [
    {
      $: {
        "android:scheme": "https",
        "android:host": "first.example",
        "android:pathPrefix": "/link/",
      },
    },
  ]);
  await AndroidConfig.Manifest.writeAndroidManifestAsync(path, input);
  const first = await readFile(path, "utf8");
  assert.ok(!first.includes("data-generated"));
  assert.equal(first.match(/<!-- @generated expo-intent-filter -->/g)?.length, 1);
  input = await AndroidConfig.Manifest.readAndroidManifestAsync(path);
  setAndroidIntentFilters(filterConfig("first.example"), input);
  assert.deepEqual(input, once);
  await AndroidConfig.Manifest.writeAndroidManifestAsync(path, input);
  assert.equal(await readFile(path, "utf8"), first);
  // The reserved comment also survives a Windows-style source manifest.
  await writeFile(path, first.replaceAll("\n", "\r\n"));
  input = await AndroidConfig.Manifest.readAndroidManifestAsync(path);
  setAndroidIntentFilters(filterConfig("second.example"), input);
  assert.equal(activity(input)["intent-filter"].length, 2);
  assert.equal(activity(input)["intent-filter"][1].data[0].$["android:host"], "second.example");
  setAndroidIntentFilters({ android: { intentFilters: [] } }, input);
  assert.deepEqual(activity(input)["intent-filter"], [manual]);
  await AndroidConfig.Manifest.writeAndroidManifestAsync(path, input);
  assert.ok(!(await readFile(path, "utf8")).includes("@generated expo-intent-filter"));
});

test("App Links creates its first filter when the activity has no filter array", () => {
  const input = manifest();
  setAndroidIntentFilters(filterConfig("first.example"), input);
  assert.equal(activity(input)["intent-filter"].length, 1);
});

test("splash restricts the API 33 attribute to a full v33 style, preserving other styles and dark colors", async (context) => {
  const directory = await mkdtemp(join(tmpdir(), "tm-splash-resources-"));
  context.after(() => rm(directory, { recursive: true, force: true }));
  const config = withAndroidSplashStyles(
    { name: "Fixture", slug: "fixture" },
    {
      backgroundColor: "#FFFFFF",
      dark: { backgroundColor: "#000000" },
    },
  );
  const other = {
    $: { name: "OtherTheme", parent: "Theme.AppCompat" },
    item: [{ $: { name: "android:windowNoTitle" }, _: "true" }],
  };
  const resource = {
    resources: {
      style: [
        globalThis.structuredClone(other),
        {
          $: { name: "Theme.App.SplashScreen", parent: "Theme.SplashScreen" },
          item: [{ $: { name: "android:windowSplashScreenBehavior" }, _: "icon_preferred" }],
        },
      ],
    },
  };
  const run = (name, results) =>
    config.mods.android[name]({
      ...config,
      modResults: results,
      modRequest: {
        projectRoot: directory,
        platformProjectRoot: directory,
        platform: "android",
        modName: name,
      },
    });
  const styles = (await run("styles", resource)).modResults;
  const base = styles.resources.style.find((style) => style.$.name === "Theme.App.SplashScreen");
  assert.equal(base.item.length, 3);
  assert.ok(!base.item.some((item) => item.$.name === "android:windowSplashScreenBehavior"));
  assert.deepEqual(
    styles.resources.style.find((style) => style.$.name === "OtherTheme"),
    other,
  );
  assert.deepEqual((await run("styles", globalThis.structuredClone(styles))).modResults, styles);
  const path = join(directory, "app/src/main/res/values-v33/styles.xml");
  await mkdir(dirname(path), { recursive: true });
  await writeFile(path, XML.format({ resources: { style: [other] } }));
  await run("dangerous", {});
  const first = await readFile(path, "utf8");
  await run("dangerous", {});
  assert.equal(await readFile(path, "utf8"), first);
  const v33 = await AndroidConfig.Resources.readResourcesXMLAsync({ path });
  assert.equal(v33.resources.style.length, 2);
  const variant = v33.resources.style.find((style) => style.$.name === "Theme.App.SplashScreen");
  assert.equal(variant.item.length, 4);
  assert.deepEqual(
    variant.item.filter((item) => item.$.name !== "android:windowSplashScreenBehavior"),
    base.item,
  );
  assert.equal(
    variant.item.find((item) => item.$.name === "android:windowSplashScreenBehavior")._,
    "icon_preferred",
  );
  for (const [name, expected] of [
    ["colors", "#FFFFFF"],
    ["colorsNight", "#000000"],
  ]) {
    const result = (await run(name, { resources: {} })).modResults;
    assert.equal(
      result.resources.color.find((color) => color.$.name === "splashscreen_background")._,
      expected,
    );
  }
});
