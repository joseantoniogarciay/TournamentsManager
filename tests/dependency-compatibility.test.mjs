import assert from "node:assert/strict";
import { Buffer } from "node:buffer";
import { copyFile, mkdtemp, readFile, rm, writeFile } from "node:fs/promises";
import { createRequire } from "node:module";
import { tmpdir } from "node:os";
import { dirname, join } from "node:path";
import test from "node:test";
import { fileURLToPath } from "node:url";

const root = dirname(dirname(fileURLToPath(import.meta.url)));
const clientRequire = createRequire(join(root, "apps/client/package.json"));
const routerRequire = createRequire(clientRequire.resolve("expo-router/package.json"));
const queryString = routerRequire("query-string");
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

test("Router query strings round-trip Unicode, repeated values and reserved characters", () => {
  const query = { name: "José 🏓", tag: ["uno", "dos"], literal: "%&+=/?" };
  assert.deepEqual({ ...queryString.parse(queryString.stringify(query)) }, query);
  const parsed = queryString.parseUrl(
    "https://fasttourney.com/account?next=%2Ftournament%2F123&name=Jos%C3%A9#section",
  );
  assert.equal(parsed.url, "https://fasttourney.com/account");
  assert.deepEqual({ ...parsed.query }, { name: "José", next: "/tournament/123" });
});

test("Router tolerates malformed percent encoding and long invalid UTF-8 input", () => {
  const malformed = "%E0%A4%A";
  assert.equal(queryString.parse(`value=${malformed}`).value, malformed);
  const repeated = "%C2".repeat(10000);
  const decoded = queryString.parse(`value=${repeated}`).value;
  assert.equal(decoded.length, 10000);
  assert.match(decoded, /^\uFFFD+$/);
  assert.equal(queryString.parse("value=a+b%2Bc").value, "a b+c");
});
