import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { runInNewContext } from "node:vm";
import test from "node:test";
import { setImmediate } from "node:timers";
import ts from "typescript";

const { Headers, Response } = globalThis;
const path = "apps/client/src/api/fetch.ts";
const code = ts.transpileModule(readFileSync(path, "utf8"), {
  compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 },
}).outputText;

function fixture(refreshResponse) {
  let authorized = false;
  let refreshes = 0;
  let invalidations = 0;
  const module = { exports: {} };
  const context = {
    exports: module.exports,
    Headers,
    Date,
    WeakMap,
    process: { env: {} },
    fetch: async (url) => {
      if (url.endsWith("/sessions/refresh")) {
        refreshes++;
        const response = await refreshResponse();
        if (response.status === 200) authorized = true;
        return response;
      }
      return new Response("{}", { status: authorized ? 200 : 401 });
    },
    require: (name) => {
      if (name === "react-native") return { Platform: { OS: "web" } };
      if (name === "./generated/session/session") {
        return {
          refreshSession: async (_, fetchFn) => {
            const response = await fetchFn("/sessions/refresh", { method: "POST" });
            return { status: response.status, data: JSON.parse(await response.text()) };
          },
        };
      }
      if (name === "@/shared/analytics/posthog-client") {
        return { isProductAnalyticsActive: () => false };
      }
      return {};
    },
  };
  runInNewContext(code, context, { filename: path });
  const api = module.exports;
  api.setMobileSessionInvalidationHandler(async () => {
    invalidations++;
  });
  return { api, refreshes: () => refreshes, invalidations: () => invalidations };
}

test("web: a network failure during refresh keeps the session and permits retry", async () => {
  let offline = true;
  const f = fixture(() => {
    if (offline) throw new TypeError("offline");
    return new Response("{}", { status: 200 });
  });
  await assert.rejects(f.api.authenticatedApiFetch("/me/tournaments"), f.api.APIConnectionError);
  assert.equal(f.invalidations(), 0);
  offline = false;
  assert.equal((await f.api.authenticatedApiFetch("/me/tournaments")).status, 200);
  assert.equal(f.refreshes(), 2);
  assert.equal(f.invalidations(), 0);
});

for (const status of [429, 500]) {
  test(`web: refresh ${status} is a safe request error without logout`, async () => {
    const f = fixture(() => new Response("{}", { status }));
    await assert.rejects(f.api.authenticatedApiFetch("/me/tournaments"), (error) => {
      assert.ok(error instanceof f.api.APIUnexpectedResponseError);
      assert.equal(error.status, status);
      return true;
    });
    assert.equal(f.invalidations(), 0);
  });
}

test("web: invalid refresh bodies do not invalidate a session", async () => {
  const f = fixture(() => new Response("invalid", { status: 200 }));
  await assert.rejects(f.api.authenticatedApiFetch("/me/tournaments"));
  assert.equal(f.invalidations(), 0);
});

test("web: rejected refresh invalidates the session once", async () => {
  const f = fixture(() => new Response("{}", { status: 401 }));
  await assert.rejects(
    f.api.authenticatedApiFetch("/me/tournaments"),
    f.api.APISessionInvalidatedError,
  );
  assert.equal(f.invalidations(), 1);
});

test("web: concurrent protected reads share a single successful refresh", async () => {
  let release;
  const barrier = new Promise((resolve) => {
    release = resolve;
  });
  const f = fixture(async () => {
    await barrier;
    return new Response("{}", { status: 200 });
  });
  const reads = ["/me/tournaments", "/me/access-methods"].map((route) =>
    f.api.authenticatedApiFetch(route),
  );
  await new Promise((resolve) => setImmediate(resolve));
  assert.equal(f.refreshes(), 1);
  release();
  assert.deepEqual(
    (await Promise.all(reads)).map((response) => response.status),
    [200, 200],
  );
  assert.equal(f.invalidations(), 0);
});
