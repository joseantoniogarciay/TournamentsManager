import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { Buffer } from "node:buffer";
import { copyFile, mkdir, mkdtemp, readFile, rm, writeFile } from "node:fs/promises";
import { createRequire } from "node:module";
import { tmpdir } from "node:os";
import { dirname, join } from "node:path";
import test from "node:test";
import { runInNewContext } from "node:vm";
import { fileURLToPath, pathToFileURL, URLSearchParams } from "node:url";

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

test("Expo code signing retains PEM Buffer input and verifies certificates and payloads", () => {
  const expoRequire = createRequire(clientRequire.resolve("expo/package.json"));
  const expoCLIRequire = createRequire(expoRequire.resolve("@expo/cli/package.json"));
  const signing = expoCLIRequire("@expo/code-signing-certificates");
  const signingRequire = createRequire(expoCLIRequire.resolve("@expo/code-signing-certificates"));
  const forge = signingRequire("node-forge");
  const pair = signing.generateKeyPair();
  const certificate = signing.generateSelfSignedCodeSigningCertificate({
    keyPair: pair,
    validityNotBefore: new Date(Date.now() - 60_000),
    validityNotAfter: new Date(Date.now() + 60_000),
    commonName: "Dependency compatibility fixture",
  });
  const pem = signing.convertCertificateToCertificatePEM(certificate);
  assert.deepEqual(forge.pem.decode(Buffer.from(pem)), forge.pem.decode(pem));
  const restored = signing.convertCertificatePEMToCertificate(pem);
  const restoredPair = signing.convertKeyPairPEMToKeyPair(signing.convertKeyPairToPEM(pair));
  signing.validateSelfSignedCertificate(restored, restoredPair);
  const payload = Buffer.from("local test payload");
  const signature = signing.signBufferRSASHA256AndVerify(
    restoredPair.privateKey,
    restored,
    payload,
  );
  const digest = (value) => forge.md.sha256.create().update(value).digest().getBytes();
  const signatureBytes = forge.util.decode64(signature);
  assert.equal(restored.publicKey.verify(digest(payload.toString("binary")), signatureBytes), true);
  assert.equal(restored.publicKey.verify(digest("altered payload"), signatureBytes), false);
});

test("Spectral retains filename and path filters with modern minimatch", () => {
  const rootRequire = createRequire(join(root, "package.json"));
  const scalarRequire = createRequire(rootRequire.resolve("@scalar/cli/package.json"));
  const bundlerRequire = createRequire(
    scalarRequire.resolve("@stoplight/spectral-ruleset-bundler"),
  );
  for (const consumer of [scalarRequire, bundlerRequire]) {
    const spectralRequire = createRequire(consumer.resolve("@stoplight/spectral-core"));
    const { minimatch } = spectralRequire("./ruleset/utils/minimatch.js");
    assert.equal(minimatch("contracts/openapi/public.yaml", "*.yaml"), true);
    assert.equal(minimatch("contracts/openapi/public.json", "**/*.{yaml,json}"), true);
    assert.equal(minimatch("contracts/openapi/public.yaml", "**/private/**"), false);
    assert.equal(minimatch("contracts/openapi/.hidden.yaml", "**/*.yaml"), false);
  }
});

test("Rollup and Expo retain file discovery, CommonJS bundling and cancellation", async (context) => {
  const rootRequire = createRequire(join(root, "package.json"));
  const scalarRequire = createRequire(rootRequire.resolve("@scalar/cli/package.json"));
  const bundlerRequire = createRequire(
    scalarRequire.resolve("@stoplight/spectral-ruleset-bundler"),
  );
  const commonjsRequire = createRequire(bundlerRequire.resolve("@rollup/plugin-commonjs"));
  const rollupGlob = commonjsRequire("glob");
  const expoRequire = createRequire(clientRequire.resolve("expo/package.json"));
  const configRequire = createRequire(expoRequire.resolve("@expo/config/package.json"));
  const current = configRequire("glob");
  const directory = await mkdtemp(join(tmpdir(), "fasttourney-glob-"));
  context.after(() => rm(directory, { recursive: true, force: true }));
  await mkdir(join(directory, "src", "nested"), { recursive: true });
  for (const name of ["a.js", "b.ts", "skip.txt", ".hidden.js"]) {
    await writeFile(join(directory, "src", "nested", name), 'module.exports = "fixture";');
  }
  const pattern = "**/*.{js,ts}";
  const options = { cwd: directory, nodir: true };
  const expected = ["src/nested/a.js", "src/nested/b.ts"];
  assert.deepEqual(rollupGlob.sync(pattern, options).sort(), expected);
  const entry = join(directory, "entry.cjs");
  await writeFile(entry, 'module.exports = require("./src/nested/a.js");');
  const { rollup } = bundlerRequire("rollup");
  const commonjs = bundlerRequire("@rollup/plugin-commonjs");
  const bundle = await rollup({
    input: entry,
    plugins: [commonjs({ dynamicRequireTargets: join(directory, "src/**/*.js") })],
  });
  try {
    const { output } = await bundle.generate({ format: "cjs", exports: "default" });
    const fixtureModule = { exports: {} };
    runInNewContext(output[0].code, { module: fixtureModule, exports: fixtureModule.exports });
    assert.equal(fixtureModule.exports, "fixture");
  } finally {
    await bundle.close();
  }
  assert.deepEqual(current.globSync(pattern, options).sort(), expected);
  assert.deepEqual((await current.glob(pattern, options)).sort(), expected);
  const controller = new globalThis.AbortController();
  const reason = new Error("intentional cancellation fixture");
  controller.abort(reason);
  await assert.rejects(
    current.glob("**/*", { ...options, signal: controller.signal }),
    (error) => error === reason,
  );
});

test("Expo fingerprint keeps positive and negated ignore patterns", () => {
  const expoRequire = createRequire(clientRequire.resolve("expo/package.json"));
  const fingerprintRequire = createRequire(expoRequire.resolve("@expo/fingerprint/package.json"));
  const { isIgnoredPath } = fingerprintRequire("./build/utils/Path.js");
  assert.equal(isIgnoredPath("assets/icon.png", ["assets/**"]), true);
  assert.equal(isIgnoredPath("src/app.tsx", ["assets/**"]), false);
  assert.equal(isIgnoredPath("assets/keep.png", ["assets/**", "!assets/keep.png"]), false);
});

test("TypeDoc expands nested glob entry points and excludes ignored files", async (context) => {
  const rootRequire = createRequire(join(root, "package.json"));
  const orvalRequire = createRequire(rootRequire.resolve("orval/package.json"));
  const typedoc = await import(pathToFileURL(orvalRequire.resolve("typedoc")).href);
  const directory = await mkdtemp(join(tmpdir(), "fasttourney-typedoc-"));
  context.after(() => rm(directory, { recursive: true, force: true }));
  await mkdir(join(directory, "nested"));
  await writeFile(
    join(directory, "nested", "visible.ts"),
    "export interface Visible { id: string }\n",
  );
  await writeFile(
    join(directory, "nested", "ignored.ts"),
    "export interface Ignored { id: string }\n",
  );
  await writeFile(
    join(directory, "tsconfig.json"),
    JSON.stringify({
      compilerOptions: {
        strict: true,
        target: "ES2022",
        module: "NodeNext",
        moduleResolution: "NodeNext",
        types: [],
      },
      include: ["**/*.ts"],
    }),
  );
  const application = await typedoc.Application.bootstrap({
    entryPoints: [join(directory, "**/*.ts")],
    entryPointStrategy: "expand",
    exclude: ["**/ignored.ts"],
    tsconfig: join(directory, "tsconfig.json"),
    logLevel: "Error",
  });
  const project = await application.convert();
  assert.ok(project);
  assert.deepEqual(
    project.getReflectionsByKind(typedoc.ReflectionKind.Interface).map((item) => item.name),
    ["Visible"],
  );
});
