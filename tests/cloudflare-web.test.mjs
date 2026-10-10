import assert from "node:assert/strict";
import { createRequire } from "node:module";
import { mkdtemp, mkdir, readFile, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { fileURLToPath, URL } from "node:url";
import test from "node:test";
import { fileInventory, stageWeb, webConfig } from "../infra/cloudflare/stage-web.mjs";

// Test with the exact runtime bundled by the pinned deployment CLI.
const wranglerRequire = createRequire(import.meta.resolve("wrangler"));
const { Miniflare, convertV4MiniflareOptions } = wranglerRequire("miniflare");
function createRuntime(options) {
  const { scriptPath, ...rest } = options;
  return new Miniflare(
    convertV4MiniflareOptions({
      ...rest,
      modules: [
        { type: "ESModule", path: scriptPath },
        { type: "ESModule", path: join(root, "infra/cloudflare/web-headers.mjs") },
      ],
    }),
  );
}
const root = fileURLToPath(new URL("../", import.meta.url));
const webSHA = "a".repeat(40);
const infrastructureSHA = "b".repeat(40);
const id = "01a0c9b0-1d38-7d47-8825-86c5af8b1237";
const path = `/tournament/${id}`;

test("Cloudflare release preserves static routing and safe dynamic tournament documents", async (context) => {
  const directory = await mkdtemp(join(tmpdir(), "fasttourney-cloudflare-"));
  context.after(() => rm(directory, { recursive: true, force: true }));
  const source = join(directory, "source");
  await mkdir(join(source, ".well-known"), { recursive: true });
  await writeFile(join(source, "deployment.json"), JSON.stringify({ commit: webSHA }));
  await writeFile(
    join(source, "index.html"),
    '<!doctype html><html><head><title>Home</title><link rel="canonical" href="https://fasttourney.com/"><meta name="robots" content="index, follow"><meta property="og:title" content="Home"></head><body><div id="root">Expo shell</div><script src="/bundle.js"></script></body></html>',
  );
  await writeFile(join(source, "account.html"), "<html><head></head><body>Account</body></html>");
  await writeFile(join(source, "bundle.js"), "/* bundle fixture */");
  await writeFile(join(source, ".well-known/apple-app-site-association"), '{"applinks":{}}');
  const release = await stageWeb(webSHA, infrastructureSHA, source, join(directory, "releases"));
  const manifest = JSON.parse(await readFile(join(release, "release.json"), "utf8"));
  assert.equal(manifest.webSHA, webSHA);
  assert.equal(manifest.infrastructureSHA, infrastructureSHA);
  assert.deepEqual(
    (await fileInventory(release)).filter(([name]) => name !== "release.json"),
    manifest.files,
  );
  await assert.rejects(stageWeb(webSHA, infrastructureSHA, source, join(directory, "releases")));

  let apiStatus = 200;
  let apiBody = { name: 'Cup <script>alert("x")</script> & 🏓', state: "published", teams: [{}] };
  let apiCalls = 0;
  const config = webConfig("test");
  const runtime = createRuntime({
    name: "web-test",
    scriptPath: join(root, "infra/cloudflare/web-worker.mjs"),
    compatibilityDate: config.compatibility_date,
    cf: false,
    assets: {
      directory: join(release, "assets"),
      binding: "ASSETS",
      run_worker_first: config.assets.run_worker_first,
      routerConfig: { has_user_worker: true },
      assetConfig: {
        html_handling: config.assets.html_handling,
        not_found_handling: config.assets.not_found_handling,
      },
    },
    outboundService: async (request) => {
      apiCalls++;
      assert.equal(request.url, `https://api.fasttourney.com/v1/tournaments/${id}`);
      assert.equal(request.headers.get("Cookie"), null);
      assert.equal(request.headers.get("Authorization"), null);
      return new globalThis.Response(
        typeof apiBody === "string" ? apiBody : JSON.stringify(apiBody),
        { status: apiStatus, headers: { "Content-Type": "application/json" } },
      );
    },
  });
  context.after(() => runtime.dispose());
  const fetch = (route, options) =>
    runtime.dispatchFetch(`https://fasttourney.com${route}`, options);

  await context.test(
    "home is indexable, files and associations are static, account is noindex",
    async () => {
      const home = await fetch("/");
      assert.equal(home.status, 200);
      assert.equal(home.headers.get("X-Robots-Tag"), null);
      assert.match(home.headers.get("Content-Security-Policy"), /frame-ancestors 'none'/);
      assert.equal(
        (await fetch("/account")).headers.get("X-Robots-Tag"),
        "noindex, nofollow, noarchive",
      );
      assert.equal(await (await fetch("/bundle.js")).text(), "/* bundle fixture */");
      const association = await fetch("/.well-known/apple-app-site-association");
      assert.equal(association.status, 200);
      assert.match(association.headers.get("Content-Type"), /application\/json/);
      assert.equal((await fetch("/.well-known/assetlinks.json")).status, 404);
      assert.equal((await fetch("/missing.js")).status, 404);
      assert.equal(apiCalls, 0);
      const previewHome = await runtime.dispatchFetch(
        "https://preview--fasttourney-web.account.workers.dev/",
      );
      assert.equal(previewHome.headers.get("X-Robots-Tag"), "noindex, nofollow, noarchive");
    },
  );
  await context.test(
    "preview uses published shell, escaped metadata and canonical without query",
    async () => {
      const response = await fetch(`${path}?tracking=1`, {
        headers: { Cookie: "private=session", Authorization: "Bearer private" },
      });
      assert.equal(apiCalls, 1, "The preview should reach the public API adapter");
      assert.equal(response.status, 200, await response.clone().text());
      assert.equal(response.headers.get("Cache-Control"), "no-store");
      assert.equal(response.headers.get("ETag"), null);
      assert.equal(response.headers.get("X-Robots-Tag"), "noindex, nofollow, noarchive");
      const html = await response.text();
      assert.match(
        html,
        /<title>Cup &lt;script&gt;alert\(&quot;x&quot;\)&lt;\/script&gt; &amp; 🏓 \| FastTourney<\/title>/,
      );
      assert.match(html, /1 team · Published/);
      assert.match(html, new RegExp(`href="https://fasttourney.com${path}"`));
      assert.match(html, /id="root"/);
      assert.equal((html.match(/<title>/g) ?? []).length, 1);
      assert.equal((html.match(/property="og:title"/g) ?? []).length, 1);
      assert.doesNotMatch(html, /tracking=1|<script>alert/);
    },
  );
  await context.test("nested tournament pages retain SPA fallback without API calls", async () => {
    const before = apiCalls;
    const response = await fetch(`${path}/standings`);
    assert.equal(response.status, 200);
    assert.match(await response.text(), /Expo shell/);
    assert.equal(response.headers.get("X-Robots-Tag"), "noindex, nofollow, noarchive");
    assert.equal(apiCalls, before);
  });
  await context.test(
    "missing tournament is real 404, other API statuses and invalid bodies are safe",
    async () => {
      apiStatus = 404;
      assert.equal((await fetch(path)).status, 404);
      for (const status of [301, 401, 429, 500]) {
        apiStatus = status;
        apiBody = "private backend trace";
        const response = await fetch(path);
        assert.equal(response.status, 503);
        assert.doesNotMatch(await response.text(), /private backend trace/);
      }
      apiStatus = 200;
      for (const body of ["not JSON", { name: "Invalid" }, "x".repeat((1 << 20) + 1)]) {
        apiBody = body;
        assert.equal((await fetch(path)).status, 503);
      }
    },
  );
  await context.test(
    "HEAD preserves status without body and POST cannot submit through the renderer",
    async () => {
      apiBody = { name: "Cup", state: "published", teams: [] };
      assert.equal(await (await fetch(path, { method: "HEAD" })).text(), "");
      const before = apiCalls;
      const response = await fetch(path, { method: "POST", body: "private input" });
      assert.equal(response.status, 405);
      assert.equal(response.headers.get("Allow"), "GET, HEAD");
      assert.equal(apiCalls, before);
    },
  );
  await context.test(
    "dedicated www worker redirects path and query without an asset origin",
    async () => {
      const redirect = createRuntime({
        scriptPath: join(root, "infra/cloudflare/web-worker.mjs"),
        compatibilityDate: config.compatibility_date,
        cf: false,
      });
      try {
        const response = await redirect.dispatchFetch(
          `https://www.fasttourney.com${path}?tracking=1`,
          { redirect: "manual" },
        );
        assert.equal(response.status, 308);
        assert.equal(response.headers.get("Location"), `https://fasttourney.com${path}?tracking=1`);
      } finally {
        await redirect.dispose();
      }
    },
  );
});

test("staging rejects source maps and a mismatched release before publication", async (context) => {
  const directory = await mkdtemp(join(tmpdir(), "fasttourney-cloudflare-invalid-"));
  context.after(() => rm(directory, { recursive: true, force: true }));
  await writeFile(join(directory, "bundle.js.map"), "private symbols");
  await assert.rejects(fileInventory(directory), /source maps/);
  await writeFile(
    join(directory, "deployment.json"),
    JSON.stringify({ commit: infrastructureSHA }),
  );
  await assert.rejects(
    stageWeb(webSHA, infrastructureSHA, directory, join(directory, "releases")),
    /manifiesto/,
  );
});
