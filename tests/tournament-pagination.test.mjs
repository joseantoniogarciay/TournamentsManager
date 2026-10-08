import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { runInNewContext } from "node:vm";
import test from "node:test";
import { setImmediate } from "node:timers";
import ts from "typescript";

function load(path, imports) {
  const code = ts.transpileModule(readFileSync(resolve(path), "utf8"), {
    compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 },
  }).outputText;
  const module = { exports: {} };
  runInNewContext(`(function(require,module,exports){${code}\n})`, {})(
    (name) => {
      if (name in imports) return imports[name];
      if (name.startsWith("@/api/generated/models/")) {
        const key = name.split("/").at(-1);
        const values = {
          accountTournamentRelationship: ["organizer", "delegated", "follower"],
          accountTournamentState: ["published", "in_progress", "completed", "cancelled"],
        };
        return {
          [key[0].toUpperCase() + key.slice(1)]: Object.fromEntries(
            (values[key] ?? []).map((v) => [v, v]),
          ),
        };
      }
      throw new Error(`Unexpected import: ${name}`);
    },
    module,
    module.exports,
  );
  return module.exports;
}
const parser = load("apps/client/src/features/league-creation/response-parser.ts", {
  "@/shared/tournaments/sport": {},
});
class APIUnexpectedResponseError extends Error {}
class APISessionInvalidatedError extends Error {}
const fetchImports = {
  APIUnexpectedResponseError,
  APISessionInvalidatedError,
  authenticatedApiFetch: () => {},
  apiFetch: () => {},
};
const id = (n) => `01a10743-eafc-713f-9dc4-${String(n).padStart(12, "0")}`;
const item = (n, relationship = "organizer") => ({
  id: id(n),
  name: `Tournament ${n}`,
  relationship,
  state: "published",
  createdAt: "2026-10-01T10:00:00Z",
  lastActivityAt: "2026-10-01T10:00:00Z",
});
const first = { items: Array.from({ length: 50 }, (_, i) => item(100 - i)), nextCursor: id(51) };
const second = { items: [item(50, "delegated")] };
function apiFor(respond) {
  return load("apps/client/src/features/league-creation/api.ts", {
    "@/api/fetch": fetchImports,
    "@/api/generated/tournaments/tournaments": { listCurrentAccountTournaments: respond },
    "@/api/generated/users/users": {},
    "./response-parser": parser,
  });
}

test("one library request returns only its requested page; delegated relationship beyond 50 is found", async () => {
  const calls = [];
  const api = apiFor(async (params, options, transport) => {
    calls.push(params);
    assert.equal(transport, fetchImports.authenticatedApiFetch);
    return { status: 200, data: params.cursor ? second : first };
  });
  assert.equal((await api.listRelatedTournaments("administered")).items.length, 50);
  assert.equal(calls.length, 1);
  assert.equal(await api.getTournamentRelationship(id(50)), "delegated");
  assert.equal(calls.length, 3);
  assert.equal(calls[2].cursor, first.nextCursor);
  assert.equal(await api.getTournamentRelationship(id(1)), null);
  assert.equal(calls.length, 5);
});

test("a full terminal page can be followed by an empty page; missing relationship is only returned after exhaustion", async () => {
  let calls = 0;
  const api = apiFor(async (params) => {
    calls++;
    return { status: 200, data: params.cursor ? { items: [] } : first };
  });
  assert.equal(await api.getTournamentRelationship(id(1)), null);
  assert.equal(calls, 2);
});

test("malformed, duplicated, unordered or mismatched cursor pages fail closed", () => {
  for (const value of [
    null,
    {},
    { items: [null] },
    { items: [], nextCursor: id(1) },
    { ...first, nextCursor: id(9) },
    { items: [item(1), item(1)] },
    { items: [item(1), item(2)] },
  ]) {
    assert.equal(parser.parseAccountTournamentPage(value), null);
  }
  assert.equal(parser.parseAccountTournamentPage(first).items.length, 50);
});

test("cursor regressions and HTTP failures never become absent permissions", async () => {
  const cyclic = apiFor(async () => ({ status: 200, data: first }));
  await assert.rejects(cyclic.getTournamentRelationship(id(1)), APIUnexpectedResponseError);
  for (const status of [400, 401, 403, 429, 500]) {
    const api = apiFor(async (params) => ({ status: params.cursor ? status : 200, data: first }));
    await assert.rejects(api.getTournamentRelationship(id(1)), APIUnexpectedResponseError);
  }
});

// A small React hook harness lets requests resolve in deliberately different orders.
function harness(respond, getRevision = () => 0) {
  const slots = [];
  let index = 0;
  let user = { id: "owner" };
  const effects = [];
  const feedback = [];
  const react = {
    useState(initial) {
      const i = index++;
      if (!(i in slots)) slots[i] = initial;
      return [
        slots[i],
        (value) => {
          slots[i] = typeof value === "function" ? value(slots[i]) : value;
        },
      ];
    },
    useRef(initial) {
      const i = index++;
      return (slots[i] ??= { current: initial });
    },
    useCallback(fn) {
      return fn;
    },
    useEffect(fn, deps) {
      const i = index++;
      const previous = slots[i];
      if (!previous || deps.some((v, n) => v !== previous.deps[n]))
        effects.push(() => {
          previous?.cleanup?.();
          slots[i] = { deps, cleanup: fn() };
        });
    },
  };
  const focus = [];
  const module = load("apps/client/src/features/league-creation/use-tournament-library.ts", {
    react,
    "expo-router": { useFocusEffect: (fn) => focus.push(fn) },
    "@/api/fetch": fetchImports,
    "@/shared/feedback/request-failure": {
      getRequestFailure: () => ({ kind: "error", messageKey: "common_request_error" }),
    },
    "@/shared/feedback/feedback-provider": {
      useFeedback: () => ({ show: (value) => feedback.push(value) }),
    },
    "@/shared/i18n/locale": { getTranslator: () => (key) => key },
    "@/shared/session/session-provider": { useSession: () => ({ user }) },
    "./api": { listRelatedTournaments: respond },
    "./tournament-library-revision": { getTournamentLibraryRevision: getRevision },
  });
  return {
    feedback,
    account(value) {
      user = value;
    },
    render() {
      index = 0;
      const state = module.useTournamentLibrary();
      effects.splice(0).forEach((fn) => fn());
      focus.splice(0).forEach((fn) => fn());
      return state;
    },
  };
}
const settle = () => new Promise((resolve) => setImmediate(resolve));
const deferred = () => {
  let resolve;
  let reject;
  const promise = new Promise((a, b) => {
    resolve = a;
    reject = b;
  });
  return { promise, resolve, reject };
};

test("append failure preserves items and cursor, duplicate taps are blocked, retry succeeds", async () => {
  let pending = deferred();
  let appends = 0;
  const h = harness(async (relationship, cursor) => {
    if (cursor) {
      appends++;
      return pending.promise;
    }
    return relationship === "administered" ? first : { items: [] };
  });
  h.render();
  await settle();
  let state = h.render();
  const request = state.loadMore();
  await state.loadMore();
  assert.equal(appends, 1);
  pending.reject(new Error("unrecognized private backend detail"));
  await request;
  state = h.render();
  assert.equal(state.administered.items.length, 50);
  assert.equal(state.administered.nextCursor, first.nextCursor);
  assert.equal(h.feedback[0].message, "common_request_error");
  pending = deferred();
  const retry = state.loadMore();
  pending.resolve(second);
  await retry;
  state = h.render();
  assert.equal(state.administered.items.length, 51);
  assert.equal(state.administered.nextCursor, undefined);
});

test("refresh wins over an older append and retains the selected relationship", async () => {
  const pending = deferred();
  const h = harness(async (relationship, cursor) =>
    cursor ? pending.promise : relationship === "administered" ? first : { items: [] },
  );
  h.render();
  await settle();
  let state = h.render();
  const append = state.loadMore();
  state.setSelectedRelationship("followed");
  state = h.render();
  await state.loadTournaments(true);
  pending.resolve(second);
  await append;
  state = h.render();
  assert.equal(state.administered.items.length, 50);
  assert.equal(state.selectedRelationship, "followed");
  assert.equal(state.isLoadingMore, false);
});

test("an append targets the original collection after tab change and cannot leak across logout", async () => {
  let pending = deferred();
  const h = harness(async (relationship, cursor) =>
    cursor ? pending.promise : relationship === "administered" ? first : { items: [] },
  );
  h.render();
  await settle();
  let state = h.render();
  const append = state.loadMore();
  state.setSelectedRelationship("followed");
  h.render();
  pending.resolve(second);
  await append;
  state = h.render();
  assert.equal(state.administered.items.length, 51);
  assert.equal(state.followed.items.length, 0);
  await state.loadTournaments(true);
  state = h.render();
  state.setSelectedRelationship("administered");
  state = h.render();
  pending = deferred();
  const old = state.loadMore();
  h.account(null);
  h.render();
  pending.resolve(second);
  await old;
  state = h.render();
  assert.equal(state.administered.items.length, 0);
  assert.equal(state.hasLoadedTournaments, false);
});

test("initial load failure exposes safe retry state without an empty library or duplicate banner", async () => {
  let fail = true;
  const h = harness(async () => {
    if (fail) throw new Error("private backend detail");
    return { items: [] };
  });
  h.render();
  await settle();
  let state = h.render();
  assert.equal(state.loadError, "common_request_error");
  assert.equal(h.feedback.length, 0);
  fail = false;
  await state.loadTournaments();
  state = h.render();
  assert.equal(state.loadError, null);
  assert.equal(state.hasLoadedTournaments, true);
});

test("confirmed follow invalidation refreshes the library while preserving its selected segment", async () => {
  let revision = 0;
  let followed = [];
  let calls = 0;
  const h = harness(
    async (relationship) => {
      calls++;
      return { items: relationship === "followed" ? followed : [item(10)] };
    },
    () => revision,
  );
  h.render();
  await settle();
  h.render().setSelectedRelationship("followed");
  h.render();
  assert.equal(calls, 2);
  followed = [item(9, "follower")];
  revision++;
  h.render();
  await settle();
  const state = h.render();
  assert.equal(calls, 4);
  assert.equal(state.selectedRelationship, "followed");
  assert.equal(state.followed.items.length, 1);
});

test("follower lookup traverses the followed collection beyond its first page", async () => {
  const api = apiFor(async (params) => {
    assert.equal(params.relationship, "followed");
    return {
      status: 200,
      data: params.cursor
        ? { items: [item(50, "follower")] }
        : {
            items: first.items.map((value) => ({ ...value, relationship: "follower" })),
            nextCursor: first.nextCursor,
          },
    };
  });
  assert.equal(await api.getTournamentRelationship(id(50), "followed"), "follower");
});
