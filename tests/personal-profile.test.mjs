import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { runInThisContext } from "node:vm";
import test from "node:test";
import { transpileForTest } from "./helpers/transpile-typescript.mjs";

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
  const code = transpileForTest(readFileSync(file, "utf8"));
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
const generatedStub = { getPersonalProfile: operation };
const api = loadModule(join(clientSource, "features/personal-profile/api"));
const id = "019abcde-1111-7111-8111-111111111111";
const away = "019abcde-2222-7222-8222-222222222222";
function profile() {
  return {
    statistics: [
      {
        sport: "football",
        played: 2,
        wins: 1,
        draws: 1,
        losses: 0,
        drawsAllowed: true,
        winPercentage: 50,
      },
    ],
    items: [
      {
        tournamentId: id,
        tournamentName: "Friends",
        sport: "football",
        stageType: "league",
        homeTeamName: "A",
        awayTeamName: "B",
        ownSide: "home",
        match: {
          id,
          stageId: id,
          round: 1,
          sequence: 1,
          homeTeamId: id,
          awayTeamId: away,
          homeSourceKind: "seeded_team",
          awaySourceKind: "seeded_team",
          state: "completed",
          resultType: "played",
          homeScore: 1,
          awayScore: 0,
          sets: [],
        },
      },
    ],
  };
}
test("profile accepts complete totals independently of history page and handles bye-only statistics", () => {
  assert.ok(api.parsePersonalProfile(profile()));
  assert.ok(api.parsePersonalProfile({ statistics: [], items: [] }));
  const bye = profile();
  bye.statistics[0] = {
    sport: "football",
    played: 0,
    wins: 0,
    draws: 0,
    losses: 0,
    drawsAllowed: false,
    winPercentage: null,
  };
  const m = bye.items[0].match;
  Object.assign(m, { state: "bye", awayTeamId: "", awaySourceKind: "bye", winnerTeamId: id });
  delete m.homeScore;
  delete m.awayScore;
  delete m.resultType;
  bye.items[0].awayTeamName = "";
  assert.ok(api.parsePersonalProfile(bye));
});
test("profile rejects incoherent statistics, unresolved history and malformed private responses", () => {
  for (const change of [
    (v) => (v.statistics[0].wins = 3),
    (v) => (v.statistics[0].winPercentage = 100),
    (v) => (v.statistics[0].winPercentage = null),
    (v) => (v.statistics[0].drawsAllowed = false),
    (v) => v.statistics.push({ ...v.statistics[0] }),
    (v) => (v.items[0].ownSide = "other"),
    (v) => (v.items[0].match.state = "pending"),
    (v) => (v.nextCursor = "private SQL"),
    (v) => (v.items[0].match.homeScore = -1),
  ]) {
    const value = profile();
    change(value);
    assert.equal(api.parsePersonalProfile(value), null);
  }
});
test("profile adapter uses the generated operation with authenticated transport, filters and cancellation", async () => {
  failure = undefined;
  response = { status: 200, data: profile() };
  const params = { sport: "football", tournament: "Friends", cursor: id };
  const signal = new globalThis.AbortController().signal;
  await api.loadPersonalProfile(params, signal);
  assert.equal(lastArguments[0], params);
  assert.equal(lastArguments[1].signal, signal);
  assert.equal(lastArguments[2], fetchStub.authenticatedApiFetch);
  for (const status of [400, 401, 429, 500, 503]) {
    response = { status, data: { title: "private SQL" } };
    await assert.rejects(api.loadPersonalProfile({}, signal), UnexpectedResponse);
  }
  response = { status: 200, data: { private: "invalid" } };
  await assert.rejects(api.loadPersonalProfile({}, signal), UnexpectedResponse);
  failure = new globalThis.DOMException("cancelled", "AbortError");
  await assert.rejects(api.loadPersonalProfile({}, signal), (e) => e === failure);
});
