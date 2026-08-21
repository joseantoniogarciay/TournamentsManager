import type { AccountLeague } from "@/api/generated/models/accountLeague";
import { AccountLeagueRelationship } from "@/api/generated/models/accountLeagueRelationship";
import { AccountLeagueState } from "@/api/generated/models/accountLeagueState";
import type { LeagueStanding } from "@/api/generated/models/leagueStanding";
import type { LeagueTeam } from "@/api/generated/models/leagueTeam";
import type { Match } from "@/api/generated/models/match";
import { MatchState } from "@/api/generated/models/matchState";
import type { PublicLeague } from "@/api/generated/models/publicLeague";
import { PublicLeagueFormat } from "@/api/generated/models/publicLeagueFormat";
import { PublicLeagueSport } from "@/api/generated/models/publicLeagueSport";
import { PublicLeagueState } from "@/api/generated/models/publicLeagueState";
import type { PublishedLeague } from "@/api/generated/models/publishedLeague";
import { PublishedLeagueState } from "@/api/generated/models/publishedLeagueState";
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

function isAccountLeagueState(value: unknown): value is AccountLeague["state"] {
  return Object.values(AccountLeagueState).includes(value as AccountLeague["state"]);
}

function isAccountLeagueRelationship(value: unknown): value is AccountLeague["relationship"] {
  return Object.values(AccountLeagueRelationship).includes(value as AccountLeague["relationship"]);
}

function parseAccountLeague(value: unknown): AccountLeague | null {
  if (!isRecord(value)) return null;
  if (
    !isUUID(value.id) ||
    typeof value.name !== "string" ||
    !isAccountLeagueState(value.state) ||
    !isDateTime(value.createdAt) ||
    !isDateTime(value.lastActivityAt) ||
    !isAccountLeagueRelationship(value.relationship)
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

function parseAccountLeagues(value: unknown): AccountLeague[] | null {
  if (!Array.isArray(value)) return null;
  return value.flatMap((item) => {
    const league = parseAccountLeague(item);
    return league ? [league] : [];
  });
}

/** Devuelve `null` cuando el contenedor paginado no cumple el contrato. */
export function parseAccountLeaguePageItems(value: unknown): AccountLeague[] | null {
  if (!isRecord(value)) return null;
  return parseAccountLeagues(value.items);
}

/** Devuelve `null` cuando la respuesta de lista no cumple el contrato. */
export function parseRecentAccountLeagues(value: unknown): AccountLeague[] | null {
  return parseAccountLeagues(value);
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

export function parseLeagueTeam(value: unknown): LeagueTeam | null {
  if (!isRecord(value)) return null;
  if (!isUUID(value.id) || typeof value.name !== "string" || typeof value.withdrawn !== "boolean") {
    return null;
  }
  return { id: value.id, name: value.name, withdrawn: value.withdrawn };
}

function parseLeagueTeams(value: unknown): LeagueTeam[] | null {
  if (!Array.isArray(value)) return null;
  return value.flatMap((item) => {
    const team = parseLeagueTeam(item);
    return team ? [team] : [];
  });
}

function parseMatch(value: unknown): Match | null {
  if (!isRecord(value)) return null;
  if (
    !isUUID(value.id) ||
    !isIntegerAtLeast(value.round, 1) ||
    !isIntegerAtLeast(value.sequence, 1) ||
    !isUUID(value.homeTeamId) ||
    !isUUID(value.awayTeamId) ||
    !Object.values(MatchState).includes(value.state as Match["state"])
  ) {
    return null;
  }
  if (
    ("homeScore" in value && !isIntegerAtLeast(value.homeScore, 0)) ||
    ("awayScore" in value && !isIntegerAtLeast(value.awayScore, 0))
  ) {
    return null;
  }

  return {
    id: value.id,
    round: value.round,
    sequence: value.sequence,
    homeTeamId: value.homeTeamId,
    awayTeamId: value.awayTeamId,
    state: value.state as Match["state"],
    ...(isIntegerAtLeast(value.homeScore, 0) ? { homeScore: value.homeScore } : {}),
    ...(isIntegerAtLeast(value.awayScore, 0) ? { awayScore: value.awayScore } : {}),
  };
}

function parseMatches(value: unknown): Match[] | null {
  if (!Array.isArray(value)) return null;
  return value.flatMap((item) => {
    const match = parseMatch(item);
    return match ? [match] : [];
  });
}

function parseLeagueStanding(value: unknown): LeagueStanding | null {
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

function parseLeagueStandings(value: unknown): LeagueStanding[] | null {
  if (!Array.isArray(value)) return null;
  return value.flatMap((item) => {
    const standing = parseLeagueStanding(item);
    return standing ? [standing] : [];
  });
}

function parseUUIDs(value: unknown): string[] | null {
  return Array.isArray(value) ? value.filter(isUUID) : null;
}

/** Valida el contenedor y filtra de forma independiente sus colecciones internas. */
export function parsePublicLeague(value: unknown): PublicLeague | null {
  if (!isRecord(value)) return null;
  const teams = parseLeagueTeams(value.teams);
  const matches = parseMatches(value.matches);
  const standings = parseLeagueStandings(value.standings);
  const championTeamIds = parseUUIDs(value.championTeamIds);
  if (
    !isUUID(value.id) ||
    typeof value.name !== "string" ||
    !Object.values(PublicLeagueSport).includes(value.sport as PublicLeague["sport"]) ||
    !Object.values(PublicLeagueFormat).includes(value.format as PublicLeague["format"]) ||
    !Object.values(PublicLeagueState).includes(value.state as PublicLeague["state"]) ||
    (value.roundRobinLegs !== 1 && value.roundRobinLegs !== 2) ||
    !teams ||
    !matches ||
    !standings ||
    !championTeamIds
  ) {
    return null;
  }
  return {
    id: value.id,
    name: value.name,
    sport: value.sport as PublicLeague["sport"],
    format: value.format as PublicLeague["format"],
    state: value.state as PublicLeague["state"],
    roundRobinLegs: value.roundRobinLegs,
    teams,
    matches,
    standings,
    championTeamIds,
  };
}

/** Valida el contenedor y filtra de forma independiente equipos y partidos. */
export function parsePublishedLeague(value: unknown): PublishedLeague | null {
  if (!isRecord(value)) return null;
  const teams = parseLeagueTeams(value.teams);
  const matches = parseMatches(value.matches);
  if (
    !isUUID(value.id) ||
    typeof value.name !== "string" ||
    !Object.values(PublishedLeagueState).includes(value.state as PublishedLeague["state"]) ||
    !teams ||
    !matches
  ) {
    return null;
  }
  return {
    id: value.id,
    name: value.name,
    state: value.state as PublishedLeague["state"],
    teams,
    matches,
  };
}
