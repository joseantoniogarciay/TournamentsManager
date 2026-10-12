import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { runInThisContext } from "node:vm";
import test from "node:test";
import ts from "typescript";

const root = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const source = join(root, "apps/client/src");
const modules = new Map();
class UnexpectedResponse extends Error {}
const fetchStub = { APIUnexpectedResponseError: UnexpectedResponse, authenticatedApiFetch() {} };
let response;
let failure;
let argumentsReceived;
const operation = async (...args) => {
  argumentsReceived = args;
  if (failure) throw failure;
  return response;
};
const generated = {
  listCurrentAccountSessions: operation,
  revokeCurrentAccountSession: operation,
  revokeOtherCurrentAccountSessions: operation,
};
function load(path) {
  const file = path.endsWith(".ts") ? path : `${path}.ts`;
  if (file.endsWith("/api/fetch.ts")) return fetchStub;
  if (file.endsWith("/api/generated/session/session.ts")) return generated;
  if (modules.has(file)) return modules.get(file).exports;
  const module = { exports: {} };
  modules.set(file, module);
  const code = ts.transpileModule(readFileSync(file, "utf8"), {
    compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 },
  }).outputText;
  runInThisContext(`(function(require,module,exports){${code}\n})`, { filename: file })(
    (specifier) =>
      load(
        specifier.startsWith("@/")
          ? join(source, specifier.slice(2))
          : resolve(dirname(file), specifier),
      ),
    module,
    module.exports,
  );
  return module.exports;
}
const api = load(join(source, "features/session-management/api"));
const id = "019abcde-1111-7111-8111-111111111111";
const page = () => ({
  items: [
    {
      id,
      createdAt: "2026-10-12T12:00:00Z",
      lastSeenAt: "2026-10-12T13:00:00Z",
      platform: "ios",
      client: "app",
      current: true,
    },
  ],
  nextCursor: null,
});

test("session parser validates the entire page and drops undeclared backend data", () => {
  assert.deepEqual(api.parseSessionPage(page()), page());
  const extra = page();
  extra.items[0].detail = "private backend error";
  assert.deepEqual(api.parseSessionPage(extra), page());
  for (const mutate of [
    (p) => delete p.nextCursor,
    (p) => (p.items[0].id = "private"),
    (p) => (p.items[0].createdAt = "yesterday"),
    (p) => (p.items[0].lastSeenAt = "2026-10-11T12:00:00Z"),
    (p) => (p.items[0].current = "true"),
    (p) => (p.items[0].platform = "private machine name"),
    (p) => (p.items[0].client = "private user-agent"),
    (p) => p.items.push(p.items[0]),
    (p) => (p.nextCursor = id),
  ]) {
    const value = page();
    mutate(value);
    assert.equal(api.parseSessionPage(value), null);
  }
  assert.equal(
    api.parseSessionPage({ items: Array(51).fill(page().items[0]), nextCursor: null }),
    null,
  );
  assert.deepEqual(api.parseSessionPage({ items: [], nextCursor: null }), {
    items: [],
    nextCursor: null,
  });
});

test("session adapter passes generated operations the shared authenticated transport and cancellation", async () => {
  const controller = new globalThis.AbortController();
  response = { status: 200, data: page() };
  assert.deepEqual(await api.listSessions(id, controller.signal), page());
  assert.deepEqual(argumentsReceived[0], { cursor: id });
  assert.equal(argumentsReceived[1].signal, controller.signal);
  assert.equal(argumentsReceived[2], fetchStub.authenticatedApiFetch);
  response = { status: 204 };
  await api.revokeSession(id, controller.signal);
  assert.equal(argumentsReceived[0], id);
  assert.equal(argumentsReceived[1].signal, controller.signal);
  assert.equal(argumentsReceived[2], fetchStub.authenticatedApiFetch);
  await api.revokeOthers(controller.signal);
  assert.equal(argumentsReceived[0].signal, controller.signal);
  assert.equal(argumentsReceived[1], fetchStub.authenticatedApiFetch);
});

test("unknown HTTP statuses, 5xx and malformed success bodies never become session data or successful revocations", async () => {
  for (const status of [400, 401, 403, 409, 429, 500, 503]) {
    response = { status, data: { title: "private", detail: "private" } };
    await assert.rejects(api.listSessions(), UnexpectedResponse);
    await assert.rejects(api.revokeSession(id), UnexpectedResponse);
    await assert.rejects(api.revokeOthers(), UnexpectedResponse);
  }
  response = { status: 200, data: { items: "private" } };
  await assert.rejects(api.listSessions(), UnexpectedResponse);
});

test("connection errors and cancellation remain intact for the shared feedback policy", async () => {
  for (const err of [
    new Error("network"),
    Object.assign(new Error("cancelled"), { name: "AbortError" }),
  ]) {
    failure = err;
    await assert.rejects(api.listSessions(), (error) => error === err);
    await assert.rejects(api.revokeSession(id), (error) => error === err);
    await assert.rejects(api.revokeOthers(), (error) => error === err);
  }
  failure = undefined;
});
