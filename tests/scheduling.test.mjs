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
const fetchStub = { APIUnexpectedResponseError: UnexpectedResponse, authenticatedApiFetch() {} };
let response;
let failure;
let lastArguments;
const operation = async (...args) => {
  lastArguments = args;
  if (failure) throw failure;
  return response;
};
const generatedStub = { setTournamentVenue: operation, setMatchSchedule: operation };
const values = loadModule(join(clientSource, "features/scheduling/values"));
const { parsePublicTournament } = loadModule(
  join(clientSource, "features/league-creation/response-parser"),
);
const api = loadModule(join(clientSource, "features/scheduling/api"));
const id = "019abcde-1111-7111-8111-111111111111";
const away = "019abcde-2222-7222-8222-222222222222";
function fixture() {
  return {
    id,
    name: "Friends",
    sport: "football",
    format: "league",
    state: "in_progress",
    roundRobinLegs: 1,
    teams: [
      { id, name: "A", withdrawn: false },
      { id: away, name: "B", withdrawn: false },
    ],
    stages: [],
    stageTeams: [],
    tieBreakPools: [],
    standings: [],
    championTeamIds: [],
    matches: [
      {
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
    ],
  };
}
test("civil dates reject impossible days, preserve leap days, and avoid timezone conversion", () => {
  assert.equal(values.dateFromInput("29/02/2026"), null);
  assert.equal(values.dateFromInput("29/02/2028"), "2028-02-29");
  assert.equal(values.dateFromInput("31/04/2026"), null);
  assert.equal(values.dateFromInput("01/01/0000"), null);
  assert.equal(values.dateFromInput(""), "");
  assert.equal(values.dateToInput("2026-10-18"), "18/10/2026");
  for (const time of ["24:00", "1:30", "18:60"]) assert.equal(values.validCivilTime(time), false);
  assert.equal(values.validCivilTime("00:00"), true);
  assert.equal(values.validTimeZone("Europe/Madrid"), true);
  assert.equal(values.validTimeZone("Local"), false);
  assert.equal(values.validVenue("ñ".repeat(160)), true);
  assert.equal(values.validVenue("ñ".repeat(161)), false);
  assert.equal(values.validVenue("Private\nCourt"), false);
});
test("optional metadata survives parsing and malformed metadata receives the safe fallback", () => {
  const value = fixture();
  assert.ok(parsePublicTournament(value));
  value.venue = "Club";
  Object.assign(value.matches[0], {
    venue: "Pista 2",
    scheduledDate: "2026-10-18",
    scheduledTime: "18:30",
    timeZone: "Europe/Madrid",
  });
  const parsed = parsePublicTournament(value);
  assert.equal(parsed.venue, "Club");
  assert.equal(parsed.matches[0].venue, "Pista 2");
  for (const patch of [
    { scheduledDate: "2026-02-29" },
    { scheduledTime: "24:00" },
    { timeZone: "Local" },
    { timeZone: undefined },
    { venue: null },
  ]) {
    const broken = globalThis.structuredClone(value);
    Object.assign(broken.matches[0], patch);
    assert.equal(parsePublicTournament(broken), null);
  }
  delete value.matches[0].scheduledTime;
  assert.equal(parsePublicTournament(value), null);
  delete value.matches[0].timeZone;
  assert.ok(parsePublicTournament(value));
  value.venue = null;
  assert.equal(parsePublicTournament(value), null);
});
test("only creator and delegated administrators see edit controls in active tournaments", () => {
  for (const relation of ["organizer", "delegated"]) {
    assert.equal(values.canEditScheduling(relation, "published"), true);
    assert.equal(values.canEditScheduling(relation, "in_progress"), true);
    for (const state of ["completed", "cancelled"])
      assert.equal(values.canEditScheduling(relation, state), false);
  }
  for (const relation of [null, undefined, "followed", "participant"])
    assert.equal(values.canEditScheduling(relation, "in_progress"), false);
});
test("scheduling adapters use generated operations, authenticated transport and cancellation signal", async () => {
  const signal = new globalThis.AbortController().signal;
  response = { status: 200, data: fixture() };
  failure = undefined;
  assert.equal((await api.saveTournamentVenue(id, "Club", signal)).id, id);
  assert.equal(lastArguments[0], id);
  assert.deepEqual(lastArguments[1], { venue: "Club" });
  assert.equal(lastArguments[2].signal, signal);
  assert.equal(lastArguments[3], fetchStub.authenticatedApiFetch);
  const input = { scheduledDate: "", scheduledTime: "", timeZone: "", venue: "" };
  await api.saveMatchSchedule(id, away, input, signal);
  assert.deepEqual(lastArguments.slice(0, 3), [id, away, input]);
  assert.equal(lastArguments[3].signal, signal);
  assert.equal(lastArguments[4], fetchStub.authenticatedApiFetch);
});
test("only recognized business problems change recovery; other responses remain generic", async () => {
  for (const [status, kind] of [
    [403, api.SchedulingForbiddenError],
    [404, api.SchedulingUnavailableError],
    [409, api.SchedulingConflictError],
  ]) {
    response = { status, data: { type: "about:blank", title: "safe", status } };
    await assert.rejects(
      api.saveTournamentVenue(id, "Club", new globalThis.AbortController().signal),
      kind,
    );
    for (const data of [
      null,
      {},
      { type: "urn:unknown", title: "private", status },
      { type: "about:blank", title: "private", status: 500 },
    ]) {
      response = { status, data };
      await assert.rejects(
        api.saveTournamentVenue(id, "Club", new globalThis.AbortController().signal),
        UnexpectedResponse,
      );
    }
  }
  for (const status of [400, 401, 429, 500, 503]) {
    response = { status, data: { type: "about:blank", title: "private", status } };
    await assert.rejects(
      api.saveTournamentVenue(id, "Club", new globalThis.AbortController().signal),
      UnexpectedResponse,
    );
  }
  response = { status: 200, data: {} };
  await assert.rejects(
    api.saveTournamentVenue(id, "Club", new globalThis.AbortController().signal),
    UnexpectedResponse,
  );
  failure = new Error("network");
  await assert.rejects(
    api.saveTournamentVenue(id, "Club", new globalThis.AbortController().signal),
    (error) => error === failure,
  );
  failure = undefined;
});
