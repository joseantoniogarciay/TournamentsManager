import type { AccountTournament } from "@/api/generated/models/accountTournament";
import { AccountTournamentRelationship } from "@/api/generated/models/accountTournamentRelationship";
import { AccountTournamentState } from "@/api/generated/models/accountTournamentState";
import type { TournamentStanding } from "@/api/generated/models/tournamentStanding";
import type { TournamentTeam } from "@/api/generated/models/tournamentTeam";
import type { Match } from "@/api/generated/models/match";
import { MatchState } from "@/api/generated/models/matchState";
import type { PublicTournament } from "@/api/generated/models/publicTournament";
import { PublicTournamentFormat } from "@/api/generated/models/publicTournamentFormat";
import { PublicTournamentSport } from "@/api/generated/models/publicTournamentSport";
import { PublicTournamentState } from "@/api/generated/models/publicTournamentState";
import type { PublishedTournament } from "@/api/generated/models/publishedTournament";
import { PublishedTournamentState } from "@/api/generated/models/publishedTournamentState";
import type { Username } from "@/api/generated/models/username";

type RecordValue = Record<string, unknown>;

function isRecord(value: unknown): value is RecordValue {
  return !!value && typeof value === "object" && !Array.isArray(value);
}

function isUUID(value: unknown): value is string {
  return (
    typeof value === "string" &&
    /^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(value)
  );
}

function isDateTime(value: unknown): value is string {
  return typeof value === "string" && Number.isFinite(Date.parse(value));
}

function isAccountTournamentState(value: unknown): value is AccountTournament["state"] {
  return Object.values(AccountTournamentState).includes(value as AccountTournament["state"]);
}

function isAccountTournamentRelationship(
  value: unknown,
): value is AccountTournament["relationship"] {
  return Object.values(AccountTournamentRelationship).includes(
    value as AccountTournament["relationship"],
  );
}

function parseAccountTournament(value: unknown): AccountTournament | null {
  if (!isRecord(value)) return null;
  if (
    !isUUID(value.id) ||
    typeof value.name !== "string" ||
    !isAccountTournamentState(value.state) ||
    !isDateTime(value.createdAt) ||
    !isDateTime(value.lastActivityAt) ||
    !isAccountTournamentRelationship(value.relationship)
  ) {
    return null;
  }

  return {
    id: value.id,
    name: value.name,
    state: value.state,
    createdAt: value.createdAt,
    lastActivityAt: value.lastActivityAt,
    relationship: value.relationship,
  };
}

function parseAccountTournaments(value: unknown): AccountTournament[] | null {
  if (!Array.isArray(value)) return null;
  return value.flatMap((item) => {
    const league = parseAccountTournament(item);
    return league ? [league] : [];
  });
}

/** Devuelve `null` cuando el contenedor paginado no cumple el contrato. */
export function parseAccountTournamentPageItems(value: unknown): AccountTournament[] | null {
  if (!isRecord(value)) return null;
  return parseAccountTournaments(value.items);
}

/** Devuelve `null` cuando la respuesta de lista no cumple el contrato. */
export function parseRecentAccountTournaments(value: unknown): AccountTournament[] | null {
  return parseAccountTournaments(value);
}

function isUsername(value: unknown): value is Username {
  return typeof value === "string" && /^[a-z0-9_]{3,30}$/.test(value);
}

/** Devuelve `null` cuando el contenedor de usernames no cumple el contrato. */
export function parseUsernames(value: unknown): Username[] | null {
  if (!isRecord(value) || !Array.isArray(value.usernames)) return null;
  return value.usernames.filter(isUsername);
}

function isIntegerAtLeast(value: unknown, minimum: number): value is number {
  return typeof value === "number" && Number.isInteger(value) && value >= minimum;
}

export function parseTournamentTeam(value: unknown): TournamentTeam | null {
  if (!isRecord(value)) return null;
  if (!isUUID(value.id) || typeof value.name !== "string" || typeof value.withdrawn !== "boolean") {
    return null;
  }
  return { id: value.id, name: value.name, withdrawn: value.withdrawn };
}

function parseTournamentTeams(value: unknown): TournamentTeam[] | null {
  if (!Array.isArray(value)) return null;
  return value.flatMap((item) => {
    const team = parseTournamentTeam(item);
    return team ? [team] : [];
  });
}

function parseMatch(value: unknown): Match | null {
  if (
    !isRecord(value) ||
    !isUUID(value.id) ||
    !isUUID(value.stageId) ||
    !isIntegerAtLeast(value.round, 1) ||
    !isIntegerAtLeast(value.sequence, 1) ||
    !(value.homeTeamId === "" || isUUID(value.homeTeamId)) ||
    !(value.awayTeamId === "" || isUUID(value.awayTeamId)) ||
    !Object.values(MatchState).includes(value.state as Match["state"])
  )
    return null;
  for (const side of ["home", "away"] as const) {
    const kind = value[side + "SourceKind"];
    const source = value[side + "SourceMatchId"];
    const team = value[side + "TeamId"];
    if (!["seeded_team", "winner", "bye"].includes(String(kind))) return null;
    if (kind === "winner" ? !isUUID(source) : source !== undefined) return null;
    if (kind === "seeded_team" && !isUUID(team)) return null;
    if (kind === "bye" && team !== "") return null;
  }
  for (const key of ["homeScore", "awayScore", "homePenalties", "awayPenalties"]) {
    if (key in value && !isIntegerAtLeast(value[key], 0)) return null;
  }
  if ("winnerTeamId" in value && !isUUID(value.winnerTeamId)) return null;
  if (
    value.state === "completed" &&
    (!isIntegerAtLeast(value.homeScore, 0) ||
      !isIntegerAtLeast(value.awayScore, 0) ||
      !value.homeTeamId ||
      !value.awayTeamId)
  )
    return null;
  if (value.state === "bye" && !isUUID(value.winnerTeamId)) return null;
  return value as unknown as Match;
}

function parseMatches(value: unknown): Match[] | null {
  if (!Array.isArray(value)) return null;
  const matches = value.map(parseMatch);
  return matches.every((match): match is Match => match !== null) ? matches : null;
}

function parseStages(value: unknown): PublicTournament["stages"] | null {
  if (!Array.isArray(value)) return null;
  for (const stage of value) {
    if (
      !isRecord(stage) ||
      !isUUID(stage.id) ||
      !isIntegerAtLeast(stage.position, 1) ||
      !["league", "single_elimination"].includes(String(stage.type)) ||
      !["pending", "in_progress", "completed", "cancelled"].includes(String(stage.state))
    )
      return null;
    if (
      stage.type === "league"
        ? stage.roundRobinLegs !== 1 && stage.roundRobinLegs !== 2
        : stage.roundRobinLegs !== undefined
    )
      return null;
  }
  return value as PublicTournament["stages"];
}

function parseTournamentStanding(value: unknown): TournamentStanding | null {
  if (!isRecord(value)) return null;
  if (
    !isIntegerAtLeast(value.position, 1) ||
    !isUUID(value.teamId) ||
    !isIntegerAtLeast(value.played, 0) ||
    !isIntegerAtLeast(value.won, 0) ||
    !isIntegerAtLeast(value.drawn, 0) ||
    !isIntegerAtLeast(value.lost, 0) ||
    !isIntegerAtLeast(value.goalsFor, 0) ||
    !isIntegerAtLeast(value.goalsAgainst, 0) ||
    !isIntegerAtLeast(value.points, 0) ||
    typeof value.goalDifference !== "number" ||
    !Number.isInteger(value.goalDifference)
  ) {
    return null;
  }
  return {
    position: value.position,
    teamId: value.teamId,
    played: value.played,
    won: value.won,
    drawn: value.drawn,
    lost: value.lost,
    goalsFor: value.goalsFor,
    goalsAgainst: value.goalsAgainst,
    goalDifference: value.goalDifference,
    points: value.points,
  };
}

function parseTournamentStandings(value: unknown): TournamentStanding[] | null {
  if (!Array.isArray(value)) return null;
  return value.flatMap((item) => {
    const standing = parseTournamentStanding(item);
    return standing ? [standing] : [];
  });
}

function parseUUIDs(value: unknown): string[] | null {
  return Array.isArray(value) ? value.filter(isUUID) : null;
}

/** Valida el contenedor y filtra de forma independiente sus colecciones internas. */
export function parsePublicTournament(value: unknown): PublicTournament | null {
  if (!isRecord(value)) return null;
  const stages = parseStages(value.stages);
  const teams = parseTournamentTeams(value.teams);
  const matches = parseMatches(value.matches);
  const standings = parseTournamentStandings(value.standings);
  const championTeamIds = parseUUIDs(value.championTeamIds);
  if (
    !isUUID(value.id) ||
    typeof value.name !== "string" ||
    !Object.values(PublicTournamentSport).includes(value.sport as PublicTournament["sport"]) ||
    !Object.values(PublicTournamentFormat).includes(value.format as PublicTournament["format"]) ||
    !Object.values(PublicTournamentState).includes(value.state as PublicTournament["state"]) ||
    (value.roundRobinLegs !== 1 && value.roundRobinLegs !== 2) ||
    !teams ||
    !matches ||
    !standings ||
    !championTeamIds ||
    !stages
  ) {
    return null;
  }
  return {
    id: value.id,
    name: value.name,
    sport: value.sport as PublicTournament["sport"],
    format: value.format as PublicTournament["format"],
    state: value.state as PublicTournament["state"],
    roundRobinLegs: value.roundRobinLegs,
    teams,
    matches,
    standings,
    championTeamIds,
    stages,
  };
}

/** Valida el contenedor y filtra de forma independiente equipos y partidos. */
export function parsePublishedTournament(value: unknown): PublishedTournament | null {
  if (!isRecord(value)) return null;
  const teams = parseTournamentTeams(value.teams);
  const matches = parseMatches(value.matches);
  if (
    !isUUID(value.id) ||
    typeof value.name !== "string" ||
    !Object.values(PublishedTournamentState).includes(
      value.state as PublishedTournament["state"],
    ) ||
    !teams ||
    !matches
  ) {
    return null;
  }
  return {
    id: value.id,
    name: value.name,
    state: value.state as PublishedTournament["state"],
    teams,
    matches,
  };
}
