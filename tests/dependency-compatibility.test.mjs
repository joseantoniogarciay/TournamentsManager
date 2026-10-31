import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { Buffer } from "node:buffer";
import { copyFile, mkdtemp, readFile, rm, writeFile } from "node:fs/promises";
import { createRequire } from "node:module";
import { tmpdir } from "node:os";
import { dirname, join } from "node:path";
import test from "node:test";
import { fileURLToPath, URLSearchParams } from "node:url";

const root = dirname(dirname(fileURLToPath(import.meta.url)));
const clientRequire = createRequire(join(root, "apps/client/package.json"));
const routerRequire = createRequire(clientRequire.resolve("expo-router/package.json"));
const { parseUrlUsingCustomBase, safeDecodeURIComponent } = routerRequire("./build/utils/url.js");
const nativeRequire = createRequire(clientRequire.resolve("react-native/package.json"));
const cliRequire = createRequire(
  nativeRequire.resolve("@react-native/community-cli-plugin/package.json"),
);
const metroRequire = createRequire(cliRequire.resolve("metro/package.json"));
const { getAssetData, getAssetSize } = metroRequire("./src/Assets.js");
const pngFixture = join(root, "apps/client/assets/google-g.png");

test("Metro reads an on-disk PNG and retains its dimensions and content hash", async () => {
  const asset = await getAssetData(pngFixture, "assets", [], null, "/assets");
  assert.equal(asset.width, 200);
  assert.equal(asset.height, 204);
  assert.equal(asset.type, "png");
  assert.deepEqual(asset.scales, [1]);
  assert.match(asset.hash, /^[a-f0-9]{32}$/);
  assert.deepEqual(getAssetSize("png", await readFile(pngFixture), pngFixture), {
    width: 200,
    height: 204,
  });
});

test("Metro preserves scaled image dimensions and non-image assets", async (context) => {
  const directory = await mkdtemp(join(tmpdir(), "fasttourney-assets-"));
  context.after(() => rm(directory, { recursive: true, force: true }));
  const image = join(directory, "logo@2x.png");
  await copyFile(pngFixture, image);
  const asset = await getAssetData(image, "assets", [], null, "/assets");
  assert.equal(asset.width, 100);
  assert.equal(asset.height, 102);
  assert.deepEqual(asset.scales, [2]);
  const text = join(directory, "document.txt");
  await writeFile(text, "fixture");
  const document = await getAssetData(text, "assets", [], null, "/assets");
  assert.equal(document.width, undefined);
  assert.equal(document.height, undefined);
});

test("Metro rejects empty and truncated images", () => {
  assert.throws(() => getAssetSize("png", Buffer.alloc(0), "empty.png"), /empty file/);
  assert.throws(() => getAssetSize("png", Buffer.from([137, 80, 78, 71]), "truncated.png"));
});

test("Router URL parsing preserves Unicode, repeated values and reserved characters", () => {
  const params = new URLSearchParams();
  params.append("name", "José 🏓");
  params.append("tag", "uno");
  params.append("tag", "dos");
  params.append("literal", "%&+=/?");
  const parsed = parseUrlUsingCustomBase(`/account?${params}`);
  assert.equal(parsed.pathname, "/account");
  assert.equal(parsed.searchParams.get("name"), "José 🏓");
  assert.deepEqual(parsed.searchParams.getAll("tag"), ["uno", "dos"]);
  assert.equal(parsed.searchParams.get("literal"), "%&+=/?");
  assert.equal(
    parseUrlUsingCustomBase("/account?next=%2Ftournament%2F123#section").searchParams.get("next"),
    "/tournament/123",
  );
});

test("Router safely decodes malformed percent encoding and long invalid UTF-8 input", () => {
  const malformed = "%E0%A4%A";
  assert.equal(safeDecodeURIComponent(malformed), malformed);
  const repeated = "%C2".repeat(10000);
  assert.equal(safeDecodeURIComponent(repeated), repeated);
  assert.equal(
    parseUrlUsingCustomBase("/account?value=a+b%2Bc").searchParams.get("value"),
    "a b+c",
  );
});

test("Expo JSI strips private C++ extensions from both Swift interface syntaxes", async () => {
  const expoRequire = createRequire(clientRequire.resolve("expo/package.json"));
  const coreRequire = createRequire(expoRequire.resolve("expo-modules-core/package.json"));
  const jsiRoot = dirname(coreRequire.resolve("expo-modules-jsi/package.json"));
  const script = await readFile(join(jsiRoot, "apple/scripts/build-xcframework.sh"), "utf8");
  const cleanup = script.match(/sed (-E )?'([^']+)' "\$swiftinterface"/);
  const expression = cleanup?.[2];
  assert.ok(expression, "review the interface cleanup when Expo changes its build script");
  const publicAPI = "public struct PublicValue {}\n";
  for (const privateType of ["__ObjC.expo.CppError", "__ObjC::expo.__ObjC::CppError"]) {
    const input = `${publicAPI}extension ${privateType} : Swift.Error {\n  public var message: Swift.String { get }\n}\n${publicAPI}`;
    const result = spawnSync("sed", [...(cleanup[1] ? ["-E"] : []), expression], {
      input,
      encoding: "utf8",
    });
    assert.equal(result.status, 0, result.stderr);
    assert.equal(result.stdout, publicAPI + publicAPI);
  }
});
