import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { runInNewContext } from "node:vm";
import test from "node:test";
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

class APIUnexpectedResponseError extends Error {}
class APISessionInvalidatedError extends Error {}
class TransportError extends Error {}
const transport = () => {};
function adapter(respond) {
  return load("apps/client/src/features/league-creation/api.ts", {
    "@/api/fetch": {
      APIUnexpectedResponseError,
      APISessionInvalidatedError,
      authenticatedApiFetch: transport,
    },
    "@/api/generated/tournaments/tournaments": {
      followTournament: respond,
      unfollowTournament: respond,
    },
    "@/api/generated/users/users": {},
    "./response-parser": {},
  });
}
test("follow and unfollow use authenticated generated operations and accept only 204", async () => {
  const api = adapter(async (id, options, fetch) => {
    assert.equal(id, "fixture");
    assert.equal(options, undefined);
    assert.equal(fetch, transport);
    return { status: 204 };
  });
  await api.followTournamentRequest("fixture");
  await api.unfollowTournamentRequest("fixture");
  for (const status of [200, 400, 401, 403, 429, 500, 502]) {
    const failed = adapter(async () => ({ status, data: { detail: "private diagnostic" } }));
    await assert.rejects(failed.followTournamentRequest("fixture"), APIUnexpectedResponseError);
    await assert.rejects(failed.unfollowTournamentRequest("fixture"), APIUnexpectedResponseError);
  }
  const missing = adapter(async () => ({ status: 404 }));
  await assert.rejects(
    missing.followTournamentRequest("fixture"),
    missing.TournamentUnavailableError,
  );
  await assert.rejects(missing.unfollowTournamentRequest("fixture"), APIUnexpectedResponseError);
  const failed = adapter(async () => {
    throw new TransportError();
  });
  await assert.rejects(failed.followTournamentRequest("fixture"), TransportError);
});
class TournamentUnavailableError extends Error {}
function harness(respond) {
  const slots = [];
  let index = 0;
  let user = { id: "player" };
  let id = "fixture";
  let relationship = null;
  const effects = [];
  const feedback = [];
  let invalidations = 0;
  const calls = [];
  const react = {
    useState(initial) {
      const i = index++;
      if (!(i in slots)) slots[i] = initial;
      return [
        slots[i],
        (v) => {
          slots[i] = v;
        },
      ];
    },
    useRef(initial) {
      const i = index++;
      return (slots[i] ??= { current: initial });
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
  const module = load("apps/client/src/features/league-creation/use-tournament-follow.ts", {
    react,
    "@/api/fetch": { APISessionInvalidatedError },
    "@/shared/feedback/feedback-provider": {
      useFeedback: () => ({ show: (v) => feedback.push(v) }),
    },
    "@/shared/feedback/request-failure": {
      getRequestFailure: (error) => ({
        kind: "error",
        messageKey:
          error instanceof TransportError ? "common_network_error" : "common_request_error",
      }),
    },
    "@/shared/i18n/locale": { getTranslator: () => (key) => key },
    "@/shared/session/session-provider": { useSession: () => ({ user }) },
    "./api": {
      TournamentUnavailableError,
      followTournamentRequest: async (id) => {
        calls.push("follow");
        return respond(id);
      },
      unfollowTournamentRequest: async (id) => {
        calls.push("unfollow");
        return respond(id);
      },
    },
    "./tournament-library-revision": { invalidateTournamentLibrary: () => invalidations++ },
  });
  return {
    calls,
    feedback,
    get invalidations() {
      return invalidations;
    },
    get relationship() {
      return relationship;
    },
    account: (value) => {
      user = value;
    },
    relation: (value) => {
      relationship = value;
    },
    route: (value) => {
      id = value;
    },
    unmount: () => slots.forEach((slot) => slot?.cleanup?.()),
    render() {
      index = 0;
      const state = module.useTournamentFollow(id, relationship, (v) => {
        relationship = v;
      });
      effects.splice(0).forEach((fn) => fn());
      return state;
    },
  };
}
function deferred() {
  let resolve, reject;
  const promise = new Promise((a, b) => {
    resolve = a;
    reject = b;
  });
  return { promise, resolve, reject };
}
test("double taps send one request; only confirmed success changes relationship and invalidates library", async () => {
  const pending = deferred();
  const h = harness(() => pending.promise);
  let state = h.render();
  const request = state.toggleFollow();
  await state.toggleFollow();
  assert.equal(h.calls.length, 1);
  assert.equal(h.relationship, null);
  assert.equal(h.render().isSaving, true);
  pending.resolve();
  await request;
  state = h.render();
  assert.equal(state.isFollowed, true);
  assert.equal(state.isSaving, false);
  assert.equal(h.invalidations, 1);
  await state.toggleFollow();
  assert.equal(h.relationship, null);
  assert.deepEqual(h.calls, ["follow", "unfollow"]);
  assert.equal(h.invalidations, 2);
});
test("failure preserves the relationship and uses safe localized recovery", async () => {
  for (const [error, message] of [
    [new TransportError("private"), "common_network_error"],
    [new Error("private"), "common_request_error"],
    [new TournamentUnavailableError(), "league_unavailable"],
  ]) {
    const h = harness(async () => {
      throw error;
    });
    h.relation("follower");
    await h.render().toggleFollow();
    assert.equal(h.relationship, "follower");
    assert.equal(h.invalidations, 0);
    assert.equal(h.feedback[0].message, message);
    assert.equal(h.render().isSaving, false);
  }
});
test("controls are absent for administrator, delegate, unknown relationship and anonymous account", async () => {
  const h = harness(async () => {});
  for (const role of ["organizer", "delegated", undefined]) {
    h.relation(role);
    const state = h.render();
    assert.equal(state.canFollow, false);
    await state.toggleFollow();
  }
  h.relation(null);
  h.account(null);
  assert.equal(h.render().canFollow, false);
  await h.render().toggleFollow();
  assert.equal(h.calls.length, 0);
});
test("logout, route replacement and unmount ignore late updates but invalidate a confirmed server mutation", async () => {
  for (const cleanup of [
    (h) => {
      h.account(null);
      h.render();
    },
    (h) => {
      h.route("another");
      h.render();
    },
    (h) => h.unmount(),
  ]) {
    const pending = deferred();
    const h = harness(() => pending.promise);
    const request = h.render().toggleFollow();
    cleanup(h);
    pending.resolve();
    await request;
    assert.equal(h.relationship, null);
    assert.equal(h.feedback.length, 0);
    assert.equal(h.invalidations, 1);
  }
});
test("invalidated sessions and stale failures do not display feedback", async () => {
  const invalid = harness(async () => {
    throw new APISessionInvalidatedError();
  });
  await invalid.render().toggleFollow();
  assert.equal(invalid.feedback.length, 0);
  const pending = deferred();
  const h = harness(() => pending.promise);
  const request = h.render().toggleFollow();
  h.unmount();
  pending.reject(new Error("private"));
  await request;
  assert.equal(h.feedback.length, 0);
});
