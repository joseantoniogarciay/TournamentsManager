import {
  captureProductIntent,
  captureProductOutcome,
  APIUnexpectedResponseError,
  apiFetch,
  authenticatedApiFetch,
} from "@/api/fetch";
import {
  addLeagueTeam,
  assignLeagueAdministrator,
  cancelLeague,
  completeLeague,
  createLeague,
  getPublicLeague,
  listLeagueAdministrators,
  listCurrentAccountLeagues,
  listRecentAccountLeagues,
  removeLeagueTeam,
  removeLeagueAdministrator,
  startLeague,
  transferLeagueOwnership,
  withdrawLeagueTeam,
} from "@/api/generated/leagues/leagues";
import type { LeagueInput, StartLeagueRequest, TeamInput, Username } from "@/api/generated/models";
import { searchUsers } from "@/api/generated/users/users";
import {
  parseAccountLeaguePageItems,
  parseLeagueTeam,
  parsePublicLeague,
  parsePublishedLeague,
  parseRecentAccountLeagues,
  parseUsernames,
} from "./response-parser";

export class UserSearchRateLimitedError extends Error {
  constructor() {
    super("Búsqueda de usuarios limitada");
  }
}

export class LeagueAdministratorConflictError extends Error {
  constructor() {
    super("Conflicto al asignar administradora de liga");
  }
}

/** La proyección pública ya no está disponible; reintentar no puede recuperarla. */
export class LeagueUnavailableError extends Error {
  constructor() {
    super("Liga no disponible");
  }
}

export async function createLeagueRequest(input: LeagueInput) {
  captureProductIntent("league_creation_submitted");
  const response = await createLeague(input, undefined, authenticatedApiFetch);
  if (response.status !== 201) throw new APIUnexpectedResponseError(response.status);
  const league = parsePublishedLeague(response.data);
  if (!league) throw new APIUnexpectedResponseError(response.status);
  captureProductOutcome("league_created", response.headers);
  return league;
}
export async function addLeagueTeamRequest(leagueID: string, input: TeamInput) {
  const response = await addLeagueTeam(leagueID, input, undefined, authenticatedApiFetch);
  if (response.status !== 201) throw new APIUnexpectedResponseError(response.status);
  const team = parseLeagueTeam(response.data);
  if (!team) throw new APIUnexpectedResponseError(response.status);
  return team;
}
export async function removeLeagueTeamRequest(leagueID: string, teamID: string) {
  const response = await removeLeagueTeam(leagueID, teamID, undefined, authenticatedApiFetch);
  if (response.status !== 204) throw new APIUnexpectedResponseError(response.status);
}
export async function withdrawLeagueTeamRequest(leagueID: string, teamID: string) {
  const response = await withdrawLeagueTeam(leagueID, teamID, undefined, authenticatedApiFetch);
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  const league = parsePublicLeague(response.data);
  if (!league) throw new APIUnexpectedResponseError(response.status);
  return league;
}
export async function startLeagueRequest(leagueID: string, input: StartLeagueRequest) {
  const response = await startLeague(leagueID, input, undefined, authenticatedApiFetch);
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  const league = parsePublicLeague(response.data);
  if (!league) throw new APIUnexpectedResponseError(response.status);
  captureProductOutcome("league_started", response.headers);
  return league;
}
export async function cancelLeagueRequest(leagueID: string) {
  const response = await cancelLeague(leagueID, undefined, authenticatedApiFetch);
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  const league = parsePublicLeague(response.data);
  if (!league) throw new APIUnexpectedResponseError(response.status);
  return league;
}
export async function completeLeagueRequest(leagueID: string) {
  const response = await completeLeague(leagueID, undefined, authenticatedApiFetch);
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  const league = parsePublicLeague(response.data);
  if (!league) throw new APIUnexpectedResponseError(response.status);
  captureProductOutcome("league_completed", response.headers);
  return league;
}
export async function assignLeagueAdministratorRequest(leagueID: string, username: string) {
  const response = await assignLeagueAdministrator(
    leagueID,
    username,
    undefined,
    authenticatedApiFetch,
  );
  if (response.status === 409) throw new LeagueAdministratorConflictError();
  if (response.status !== 204) throw new APIUnexpectedResponseError(response.status);
  captureProductOutcome("league_administrator_assigned", response.headers);
}
export async function listLeagueAdministratorUsernames(leagueID: string) {
  const response = await listLeagueAdministrators(leagueID, undefined, authenticatedApiFetch);
  if (response.status === 404) throw new LeagueUnavailableError();
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  const usernames = parseUsernames(response.data);
  if (!usernames) throw new APIUnexpectedResponseError(response.status);
  return usernames;
}
export async function removeLeagueAdministratorRequest(leagueID: string, username: string) {
  const response = await removeLeagueAdministrator(
    leagueID,
    username as Username,
    undefined,
    authenticatedApiFetch,
  );
  if (response.status !== 204) throw new APIUnexpectedResponseError(response.status);
  captureProductOutcome("league_administrator_removed", response.headers);
}
export async function transferLeagueOwnershipRequest(leagueID: string, username: string) {
  const response = await transferLeagueOwnership(
    leagueID,
    { username: username as Username },
    undefined,
    authenticatedApiFetch,
  );
  if (response.status !== 204) throw new APIUnexpectedResponseError(response.status);
}
export async function searchPublicUsernames(query: string, signal: AbortSignal) {
  const response = await searchUsers({ query: query as Username }, { signal }, apiFetch);
  if (response.status === 200) {
    const usernames = parseUsernames(response.data);
    if (!usernames) throw new APIUnexpectedResponseError(response.status);
    return usernames;
  }
  if (response.status === 429) throw new UserSearchRateLimitedError();
  throw new APIUnexpectedResponseError(response.status);
}
export async function getLeague(leagueID: string) {
  const response = await getPublicLeague(leagueID, undefined, apiFetch);
  const status = (response as { status: number }).status;
  if (status === 200) {
    const league = parsePublicLeague((response as { data: unknown }).data);
    if (!league) throw new APIUnexpectedResponseError(status);
    return league;
  }
  if (status === 404) throw new LeagueUnavailableError();
  throw new APIUnexpectedResponseError(status);
}
export async function getLeagueRelationship(leagueID: string) {
  const response = await listCurrentAccountLeagues(
    { relationship: "administered", limit: 50 },
    undefined,
    authenticatedApiFetch,
  );
  if (response.status !== 200) return undefined;
  const items = parseAccountLeaguePageItems(response.data);
  if (!items) return undefined;
  return items.find((league) => league.id === leagueID)?.relationship;
}
export async function listRelatedLeagues(relationship: "administered" | "followed") {
  const response = await listCurrentAccountLeagues(
    { relationship, limit: 50 },
    undefined,
    authenticatedApiFetch,
  );
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  const items = parseAccountLeaguePageItems(response.data);
  if (!items) throw new APIUnexpectedResponseError(response.status);
  return items;
}

export async function listRecentRelatedLeagues() {
  const response = await listRecentAccountLeagues(undefined, authenticatedApiFetch);
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  const leagues = parseRecentAccountLeagues(response.data);
  if (!leagues) throw new APIUnexpectedResponseError(response.status);
  return leagues;
}
