import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { runInThisContext } from "node:vm";
import test from "node:test";
import ts from "typescript";

const root = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const clientSource = join(root, "apps/client/src");
const modules = new Map();
// Load the actual pure client modules, preserving their aliases and generated enums.
function loadModule(path) {
  const file = path.endsWith(".ts") ? path : `${path.replace(/\.js$/, "")}.ts`;
  if (modules.has(file)) return modules.get(file).exports;
  if (file.endsWith("/api/fetch.ts")) return fetchStub;
  if (file.endsWith("/api/generated/tournaments/tournaments.ts")) return generatedStub;
  const module = { exports: {} };
  modules.set(file, module);
  const code = ts.transpileModule(readFileSync(file, "utf8"), {
    compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 },
  }).outputText;
  const requireSource = (specifier) =>
    loadModule(
      specifier.startsWith("@/")
        ? join(clientSource, specifier.slice(2))
        : resolve(dirname(file), specifier),
    );
  runInThisContext(`(function(require,module,exports){${code}\n})`, { filename: file })(
    requireSource,
    module,
    module.exports,
  );
  return module.exports;
}

class UnexpectedResponse extends Error {
  constructor(status) {
    super(String(status));
  }
}
const fetchStub = {
  APIUnexpectedResponseError: UnexpectedResponse,
  authenticatedApiFetch() {},
  apiFetch() {},
};
let response;
let failure;
let lastArguments;
const operation = async (...args) => {
  lastArguments = args;
  if (failure) throw failure;
  return response;
};
const generatedStub = { listPersonalMatches: operation, listTournamentEditions: operation };
const api = loadModule(join(clientSource, "features/tournament-browsing/api"));
const id = "019abcde-1111-7111-8111-111111111111";
const away = "019abcde-2222-7222-8222-222222222222";
function matchPage() {
  return {
    hasMatches: true,
    items: [
      {
        tournamentId: id,
        tournamentName: "Friends",
        sport: "football",
        stageType: "league",
        homeTeamName: "A",
        awayTeamName: "B",
        match: {
          id,
          stageId: id,
          round: 1,
          sequence: 1,
          homeTeamId: id,
          awayTeamId: away,
          homeSourceKind: "seeded_team",
          awaySourceKind: "seeded_team",
          state: "pending",
          sets: [],
        },
      },
    ],
  };
}
test("personal collection rejects malformed matches and preserves evidence of other view", () => {
  assert.ok(api.parsePersonalMatchPage(matchPage()));
  assert.deepEqual(api.parsePersonalMatchPage({ items: [], hasMatches: true }), {
    items: [],
    hasMatches: true,
  });
  const malformed = matchPage();
  malformed.items[0].match.scheduledDate = "2026-02-31";
  assert.equal(api.parsePersonalMatchPage(malformed), null);
  assert.equal(api.parsePersonalMatchPage({ items: [], hasMatches: "true" }), null);
  assert.equal(api.parsePersonalMatchPage({ ...matchPage(), nextCursor: "private SQL" }), null);
});
test("edition collection validates all rows without silently dropping broken historical data", () => {
  const value = {
    items: [
      {
        id,
        name: "Edition",
        state: "completed",
        createdAt: "2026-10-11T12:00:00Z",
        champions: ["A", "B"],
      },
    ],
  };
  assert.ok(api.parseEditionPage(value));
  assert.equal(api.parseEditionPage({ items: [{ ...value.items[0], champions: [5] }] }), null);
  assert.equal(
    api.parseEditionPage({ items: [{ ...value.items[0], createdAt: "invalid" }] }),
    null,
  );
});
test("adapters use generated operations, propagate cancellation and fallback for unknown problems", async () => {
  failure = undefined;
  response = { status: 200, data: matchPage() };
  const controller = new globalThis.AbortController();
  await api.getPersonalMatches("pending", undefined, controller.signal);
  assert.equal(lastArguments.at(-1), fetchStub.authenticatedApiFetch);
  assert.equal(lastArguments[1].signal, controller.signal);
  response = {
    status: 200,
    data: {
      items: [
        {
          id,
          name: "Edition",
          state: "published",
          createdAt: "2026-10-11T12:00:00Z",
          champions: [],
        },
      ],
    },
  };
  await api.getEditions(id, undefined, controller.signal);
  assert.equal(lastArguments.at(-1), fetchStub.apiFetch);
  for (const status of [400, 404, 429, 500, 503]) {
    response = {
      status,
      data: { type: "private:unknown", title: "private SQL", detail: "secret" },
    };
    await assert.rejects(api.getEditions(id, undefined, controller.signal), UnexpectedResponse);
  }
  response = { status: 200, data: { items: [], hasMatches: "invalid" } };
  await assert.rejects(
    api.getPersonalMatches("results", undefined, controller.signal),
    UnexpectedResponse,
  );
  const canceled = new Error("Canceled");
  canceled.name = "AbortError";
  failure = canceled;
  await assert.rejects(
    api.getEditions(id, undefined, controller.signal),
    (error) => error === canceled,
  );
  failure = undefined;
});

// Drive the real collection hook while resolving requests in a chosen order.
function collectionHarness(fetchPage) {
  const slots = [];
  let index = 0;
  let identity = "account:pending";
  const effects = [];
  let focused;
  const feedback = [];
  const react = {
    useState(initial) {
      const slot = index++;
      if (!(slot in slots)) slots[slot] = initial;
      return [
        slots[slot],
        (value) => {
          slots[slot] = typeof value === "function" ? value(slots[slot]) : value;
        },
      ];
    },
    useRef(initial) {
      const slot = index++;
      return (slots[slot] ??= { current: initial });
    },
    useCallback(fn) {
      return fn;
    },
    useEffect(fn, deps) {
      const slot = index++;
      const previous = slots[slot];
      if (!previous || deps.some((value, position) => value !== previous.deps[position]))
        effects.push(() => {
          previous?.cleanup?.();
          slots[slot] = { deps, cleanup: fn() };
        });
    },
  };
  const imports = {
    react,
    "expo-router": {
      useFocusEffect(fn) {
        focused = fn;
      },
    },
    "@/api/fetch": { APISessionInvalidatedError: class extends Error {} },
    "@/shared/feedback/feedback-provider": {
      useFeedback: () => ({ show: (value) => feedback.push(value) }),
    },
    "@/shared/feedback/request-failure": {
      getRequestFailure: () => ({ kind: "generic-error", messageKey: "common_request_error" }),
    },
    "@/shared/i18n/locale": { getTranslator: () => (key) => key },
  };
  const module = { exports: {} };
  const source = ts.transpileModule(
    readFileSync(join(clientSource, "features/tournament-browsing/use-collection.ts"), "utf8"),
    { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } },
  ).outputText;
  runInThisContext(`(function(require,module,exports){${source}\n})`)(
    (name) => imports[name],
    module,
    module.exports,
  );
  const key = (item) => item.id;
  return {
    feedback,
    identity(value) {
      identity = value;
    },
    focus() {
      focused();
    },
    render() {
      index = 0;
      const state = module.exports.useCollection(identity, fetchPage, key);
      effects.splice(0).forEach((fn) => fn());
      return state;
    },
  };
}
function deferredPage() {
  let resolve, reject;
  const promise = new Promise((success, failure) => {
    resolve = success;
    reject = failure;
  });
  return { promise, resolve, reject };
}
test("collection preserves content on refresh failure and retries pagination without duplicates", async () => {
  const first = { items: [{ id: "first" }], nextCursor: "first" };
  let request = deferredPage();
  let calls = 0;
  const h = collectionHarness(async () => {
    calls++;
    return request.promise;
  });
  let state = h.render();
  const initial = state.load();
  request.resolve(first);
  await initial;
  state = h.render();
  request = deferredPage();
  const refresh = state.load("refresh");
  await state.load("refresh");
  assert.equal(calls, 2);
  request.reject(new Error("private detail"));
  await refresh;
  state = h.render();
  assert.deepEqual(state.page, first);
  assert.equal(h.feedback[0].message, "common_request_error");
  request = deferredPage();
  const append = state.load("more");
  request.reject(new Error("private detail"));
  await append;
  state = h.render();
  assert.deepEqual(state.page, first);
  assert.equal(state.moreError, "common_request_error");
  request = deferredPage();
  const retry = state.load("more");
  request.resolve({ items: [{ id: "first" }, { id: "second" }] });
  await retry;
  state = h.render();
  assert.deepEqual(state.page.items, [{ id: "first" }, { id: "second" }]);
  assert.equal(state.moreError, null);
});
test("changing account or view aborts stale requests and cannot restore their data", async () => {
  const old = deferredPage();
  let signal;
  const h = collectionHarness(async (_cursor, currentSignal) => {
    signal = currentSignal;
    return old.promise;
  });
  let state = h.render();
  const pending = state.load();
  h.identity("other:results");
  h.render();
  assert.equal(signal.aborted, true);
  old.resolve({ items: [{ id: "private-old-account" }] });
  await pending;
  state = h.render();
  assert.equal(state.page, null);
  assert.equal(state.error, null);
  assert.deepEqual(h.feedback, []);
  h.identity(null);
  state = h.render();
  await state.load();
  assert.equal(h.render().page, null);
});

test("refresh rebuilds loaded pages atomically and removes rows no longer returned", async () => {
  const calls = [];
  let refreshed = false;
  let failContinuation = false;
  const h = collectionHarness(async (cursor) => {
    calls.push(cursor);
    if (failContinuation && cursor) throw new Error("private detail");
    if (!cursor) return { items: [{ id: "first" }], nextCursor: "first" };
    return { items: [{ id: refreshed ? "replacement" : "second" }] };
  });
  await h.render().load();
  await h.render().load("more");
  assert.deepEqual(
    h.render().page.items.map((item) => item.id),
    ["first", "second"],
  );
  refreshed = true;
  failContinuation = true;
  await h.render().load("refresh");
  assert.deepEqual(
    h.render().page.items.map((item) => item.id),
    ["first", "second"],
  );
  assert.equal(h.feedback.at(-1).message, "common_request_error");
  failContinuation = false;
  await h.render().load("refresh");
  assert.deepEqual(
    h.render().page.items.map((item) => item.id),
    ["first", "replacement"],
  );
  assert.equal(h.render().page.nextCursor, undefined);
  assert.deepEqual(calls, [undefined, "first", undefined, "first", undefined, "first"]);
});

test("changing identity during a multi-page refresh aborts and discards its continuation", async () => {
  let continuation;
  let signal;
  const h = collectionHarness(async (cursor, currentSignal) => {
    signal = currentSignal;
    if (cursor && continuation) return continuation.promise;
    return cursor
      ? { items: [{ id: "second" }] }
      : { items: [{ id: "first" }], nextCursor: "first" };
  });
  await h.render().load();
  await h.render().load("more");
  continuation = deferredPage();
  const refresh = h.render().load("refresh");
  await Promise.resolve();
  h.identity("other:results");
  h.render();
  assert.equal(signal.aborted, true);
  continuation.resolve({ items: [{ id: "old-private-second" }] });
  await refresh;
  assert.equal(h.render().page, null);
  assert.deepEqual(h.feedback, []);
});
