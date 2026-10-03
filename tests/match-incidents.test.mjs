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
const { getIncidentInput, editableScoreFromMatch, parseSetScores } = loadModule(
  join(clientSource, "features/match-results/result-input"),
);
const { parsePublishedTournament } = loadModule(
  join(clientSource, "features/league-creation/response-parser"),
);
function score(type = "retirement", sets = []) {
  return {
    incidentType: type,
    incidentSide: "home",
    home: "",
    away: "",
    homePenalties: "",
    awayPenalties: "",
    sets: sets.map(([home, away]) => ({ home: String(home), away: String(away) })),
  };
}

test("incident form requires a side and distinguishes omitted, partial and final scores", () => {
  assert.equal(getIncidentInput({ ...score(), incidentSide: undefined }, "football", 0), null);
  assert.deepEqual(
    getIncidentInput({ ...score("no_show"), home: "52", away: "48" }, "basketball", 0),
    { type: "no_show", side: "home" },
  );
  assert.deepEqual(getIncidentInput(score(), "basketball", 0), {
    type: "retirement",
    side: "home",
  });
  assert.equal(getIncidentInput({ ...score(), home: "52" }, "basketball", 0), null);
  assert.deepEqual(getIncidentInput({ ...score(), home: "52", away: "48" }, "basketball", 0), {
    type: "retirement",
    side: "home",
    partialHomeScore: 52,
    partialAwayScore: 48,
  });
  assert.equal(
    getIncidentInput({ ...score(), home: "2147483648", away: "1" }, "football", 0),
    null,
  );
});

test("partial sets stop at the unfinished set and cannot describe an already finished match", () => {
  assert.ok(
    getIncidentInput(
      score("retirement", [
        [6, 4],
        [2, 1],
      ]),
      "padel",
      3,
    ),
  );
  assert.equal(
    getIncidentInput(
      score("retirement", [
        [2, 1],
        [6, 4],
      ]),
      "padel",
      3,
    ),
    null,
  );
  assert.equal(
    getIncidentInput(
      score("retirement", [
        [6, 4],
        [6, 0],
      ]),
      "padel",
      3,
    ),
    null,
  );
  assert.equal(
    parseSetScores(
      score("retirement", [
        [6, 4],
        [2, 1],
      ]),
      3,
      "padel",
    ),
    null,
  );
  for (const [points, limit] of [
    [15, 21],
    [21, 30],
  ]) {
    assert.ok(
      getIncidentInput(score("retirement", [[limit - 1, limit - 1]]), "badminton", 3, points),
    );
    assert.equal(
      getIncidentInput(score("retirement", [[limit + 1, limit - 1]]), "badminton", 3, points),
      null,
    );
  }
  assert.ok(
    getIncidentInput(
      score("retirement", [
        [25, 20],
        [20, 25],
        [25, 20],
        [20, 25],
        [14, 14],
      ]),
      "volleyball",
      5,
    ),
  );
});

const id = "019abcde-1111-7111-8111-111111111111",
  away = "019abcde-1111-7111-8111-111111111112";
const match = {
  id,
  stageId: id,
  round: 1,
  sequence: 1,
  homeTeamId: id,
  awayTeamId: away,
  homeSourceKind: "seeded_team",
  awaySourceKind: "seeded_team",
  state: "completed",
  resultType: "retirement",
  homeScore: 0,
  awayScore: 20,
  sets: [],
  incident: { type: "retirement", side: "home", partialHomeScore: 52, partialAwayScore: 48 },
};
function tournament(value) {
  return {
    id,
    name: "Incident",
    sport: "basketball",
    state: "published",
    teams: [],
    matches: [value],
  };
}

test("editing an incident restores actual partial scores rather than administrative scores", () => {
  const editable = editableScoreFromMatch(match, 3);
  assert.equal(editable.home, "52");
  assert.equal(editable.away, "48");
  assert.equal(editable.incidentType, "retirement");
  assert.equal(editable.incidentSide, "home");
  const normal = editableScoreFromMatch({ ...match, resultType: "played", incident: undefined }, 3);
  assert.equal(normal.home, "0");
  assert.equal(normal.away, "20");
  assert.equal(normal.incidentType, undefined);
});

test("response parser keeps incident metadata and rejects contradictory or malformed variants", () => {
  assert.deepEqual(parsePublishedTournament(tournament(match)).matches[0].incident, match.incident);
  for (const incident of [
    undefined,
    null,
    { type: "no_show", side: "home" },
    { type: "retirement", side: "both" },
    { type: "retirement", side: "home", partialHomeScore: 1 },
    { type: "retirement", side: "home", partialSets: [] },
    { type: "retirement", side: "home", partialSets: [{ homeScore: -1, awayScore: 0 }] },
  ]) {
    assert.equal(parsePublishedTournament(tournament({ ...match, incident })), null);
  }
  assert.equal(parsePublishedTournament(tournament({ ...match, resultType: "played" })), null);
  assert.equal(
    parsePublishedTournament(
      tournament({
        ...match,
        resultType: "no_show",
        incident: { type: "no_show", side: "away", partialHomeScore: 0, partialAwayScore: 0 },
      }),
    ),
    null,
  );
});
