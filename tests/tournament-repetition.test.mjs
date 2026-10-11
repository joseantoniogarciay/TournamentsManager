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
  if (file.endsWith("/api/generated/users/users.ts")) return {};
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
const generatedStub = { repeatTournament: operation, joinTournamentWithTeamInvitation: operation };
const values = loadModule(join(clientSource, "features/tournament-repetition/values"));
const { parsePublicTournament, parseTournamentTeamInvitation } = loadModule(
  join(clientSource, "features/league-creation/response-parser"),
);
const api = loadModule(join(clientSource, "features/tournament-repetition/api"));
const invitationApi = loadModule(join(clientSource, "features/league-creation/api"));
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
test("only current organizers can repeat completed sources", () => {
  for (const relation of [undefined, null, "followed", "delegated", "organizer"])
    for (const state of ["published", "in_progress", "completed", "cancelled"])
      assert.equal(
        values.canRepeatTournament(relation, state),
        relation === "organizer" && state === "completed",
      );
});
test("review validates unique names and places the selected own team first", () => {
  const teams = [
    { key: "first", name: " Blue " },
    { key: "second", name: "Red" },
  ];
  assert.deepEqual(values.reviewedRepetitionInput(id, " New edition ", teams, "second"), {
    requestId: id,
    name: "New edition",
    teams: [{ name: "Red" }, { name: "Blue" }],
  });
  assert.equal(teams[0].name, " Blue ");
  for (const [name, entries, selected] of [
    ["", teams, "first"],
    ["a".repeat(57), teams, "first"],
    ["New", [], "first"],
    ["New", teams, "missing"],
    [
      "New",
      [
        { key: "x", name: "Blue" },
        { key: "y", name: " blue " },
      ],
      "x",
    ],
    ["New", [{ key: "x", name: "a".repeat(101) }], "x"],
    ["New", Array.from({ length: 65 }, (_, i) => ({ key: String(i), name: String(i) })), "0"],
  ])
    assert.equal(values.reviewedRepetitionInput(id, name, entries, selected), null);
});
test("inherited configuration parser rejects invalid or closed suggestions", () => {
  for (const suggestion of [
    { format: "league", roundRobinLegs: 2 },
    { format: "single_elimination" },
    {
      format: "league_then_single_elimination",
      roundRobinLegs: 1,
      leagueStructure: "single_table",
      qualifierCount: 4,
    },
    {
      format: "league_then_single_elimination",
      roundRobinLegs: 2,
      leagueStructure: "groups",
      groupCount: 4,
      qualifiersPerGroup: 2,
    },
  ]) {
    const valid = { ...fixture(), state: "published", matches: [], suggestedStart: suggestion };
    assert.deepEqual(parsePublicTournament(valid)?.suggestedStart, suggestion);
    assert.equal(parsePublicTournament({ ...valid, state: "completed" }), null);
    assert.equal(
      parsePublicTournament({ ...valid, suggestedStart: { ...suggestion, private: "field" } }),
      null,
    );
  }
  for (const suggestedStart of [
    null,
    {},
    { format: "league" },
    { format: "league", roundRobinLegs: 3 },
    {
      format: "league_then_single_elimination",
      roundRobinLegs: 1,
      leagueStructure: "groups",
      groupCount: 1,
      qualifiersPerGroup: 2,
    },
  ])
    assert.equal(parsePublicTournament({ ...fixture(), state: "published", suggestedStart }), null);
});
test("invitation choices contain only valid names and preserve backward compatibility", () => {
  const base = { tournamentId: id, tournamentName: "Friends" };
  assert.deepEqual(parseTournamentTeamInvitation(base), base);
  assert.deepEqual(
    parseTournamentTeamInvitation({
      ...base,
      allowsExistingTeams: true,
      availableTeamNames: ["Red", "Blue"],
    }).availableTeamNames,
    ["Red", "Blue"],
  );
  assert.deepEqual(
    parseTournamentTeamInvitation({ ...base, allowsExistingTeams: true }).availableTeamNames,
    [],
  );
  for (const bad of [
    { availableTeamNames: ["Red"] },
    { allowsExistingTeams: true, availableTeamNames: [""] },
    { allowsExistingTeams: "true" },
    { allowsExistingTeams: true, availableTeamNames: [17] },
  ])
    assert.equal(parseTournamentTeamInvitation({ ...base, ...bad }), null);
});
test("repetition uses generated transport, safe business feedback and common fallback", async () => {
  const signal = new globalThis.AbortController().signal;
  const input = { requestId: id, name: "Friends", teams: [{ name: "A" }] };
  response = { status: 201, data: { ...fixture(), state: "published", matches: [] } };
  assert.equal((await api.repeatTournamentRequest(id, input, signal)).id, id);
  assert.deepEqual(lastArguments, [id, input, { signal }, fetchStub.authenticatedApiFetch]);
  for (const [status, reason] of [
    [403, "forbidden"],
    [404, "unavailable"],
    [409, "not_completed"],
  ]) {
    response = { status, data: { type: "about:blank", title: "private", status } };
    await assert.rejects(
      api.repeatTournamentRequest(id, input, signal),
      (error) =>
        error instanceof api.RepetitionRejectedError &&
        error.reason === reason &&
        !error.message.includes("private"),
    );
  }
  for (const status of [400, 401, 429, 500, 503]) {
    response = { status, data: { title: "private", detail: "SQL" } };
    await assert.rejects(api.repeatTournamentRequest(id, input, signal), UnexpectedResponse);
  }
  response = { status: 201, data: {} };
  await assert.rejects(api.repeatTournamentRequest(id, input, signal), UnexpectedResponse);
  for (const error of [
    new Error("transport"),
    new globalThis.DOMException("Aborted", "AbortError"),
  ]) {
    failure = error;
    await assert.rejects(
      api.repeatTournamentRequest(id, input, signal),
      (received) => received === error,
    );
  }
  failure = undefined;
});

test("occupied team has specific safe feedback; malformed and unknown problems fall back", async () => {
  response = {
    status: 409,
    data: {
      type: "urn:tournaments-manager:problem:team-occupied",
      status: 409,
      title: "private",
      detail: "SQL",
    },
  };
  await assert.rejects(
    invitationApi.joinTournamentWithTeamInvitationRequest("token", "Red"),
    invitationApi.TournamentTeamOccupiedError,
  );
  response = { status: 409, data: { type: "about:blank", status: 409, title: "private" } };
  await assert.rejects(
    invitationApi.joinTournamentWithTeamInvitationRequest("token", "Red"),
    invitationApi.TournamentTeamInvitationConflictError,
  );
  for (const data of [
    {},
    { type: "urn:unknown", status: 409, title: "private" },
    { type: "urn:tournaments-manager:problem:team-occupied", status: 500, title: "private" },
  ]) {
    response = { status: 409, data };
    await assert.rejects(
      invitationApi.joinTournamentWithTeamInvitationRequest("token", "Red"),
      UnexpectedResponse,
    );
    await assert.rejects(
      api.repeatTournamentRequest(
        id,
        { requestId: id, name: "Friends", teams: [{ name: "A" }] },
        new globalThis.AbortController().signal,
      ),
      UnexpectedResponse,
    );
  }
});
