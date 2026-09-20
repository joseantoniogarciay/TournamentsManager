import {
  captureProductIntent,
  captureProductOutcome,
  APIUnexpectedResponseError,
  apiFetch,
  authenticatedApiFetch,
} from "@/api/fetch";
import {
  addTournamentTeam,
  assignTournamentAdministrator,
  cancelTournament,
  completeTournament,
  createTournamentTeamInvitation,
  createTournament,
  getPublicTournament,
  listTournamentAdministrators,
  listCurrentAccountTournaments,
  listRecentAccountTournaments,
  inspectTournamentTeamInvitation,
  joinTournamentWithTeamInvitation,
  removeTournamentTeam,
  removeTournamentAdministrator,
  revokeTournamentTeamInvitation,
  startTournament,
  startTournamentElimination,
  transferTournamentOwnership,
  withdrawTournamentTeam,
} from "@/api/generated/tournaments/tournaments";
import type {
  TournamentInput,
  StartTournamentRequest,
  TeamInput,
  TournamentTeamInvitationToken,
  Username,
} from "@/api/generated/models";
import { searchUsers } from "@/api/generated/users/users";
import {
  parseAccountTournamentPageItems,
  parseTournamentTeam,
  parsePublicTournament,
  parsePublishedTournament,
  parseRecentAccountTournaments,
  parseTournamentTeamInvitation,
  parseTournamentTeamInvitationToken,
  parseTournamentTeamRegistration,
  parseUsernames,
} from "./response-parser";

export class UserSearchRateLimitedError extends Error {
  constructor() {
    super("Búsqueda de usuarios limitada");
  }
}

export class TournamentAdministratorConflictError extends Error {
  constructor() {
    super("Conflicto al asignar administradora de liga");
  }
}

/** La proyección pública ya no está disponible; reintentar no puede recuperarla. */
export class TournamentUnavailableError extends Error {
  constructor() {
    super("Liga no disponible");
  }
}

export class TournamentTeamInvitationUnavailableError extends Error {
  constructor() {
    super("Invitación de equipo no disponible");
  }
}

export class TournamentTeamInvitationConflictError extends Error {
  constructor() {
    super("La invitación no puede aceptar este equipo");
  }
}

export class TournamentConfigurationConflictError extends Error {
  constructor() {
    super("La composición no satisface la configuración del torneo");
  }
}

export class TournamentStageTransitionConflictError extends Error {
  constructor() {
    super("La fase de clasificación todavía no puede cerrarse");
  }
}

export async function createTournamentRequest(input: TournamentInput) {
  captureProductIntent("league_creation_submitted");
  const response = await createTournament(input, undefined, authenticatedApiFetch);
  if (response.status !== 201) throw new APIUnexpectedResponseError(response.status);
  const league = parsePublishedTournament(response.data);
  if (!league) throw new APIUnexpectedResponseError(response.status);
  captureProductOutcome("league_created", response.headers);
  return league;
}
export async function addTournamentTeamRequest(leagueID: string, input: TeamInput) {
  const response = await addTournamentTeam(leagueID, input, undefined, authenticatedApiFetch);
  if (response.status !== 201) throw new APIUnexpectedResponseError(response.status);
  const team = parseTournamentTeam(response.data);
  if (!team) throw new APIUnexpectedResponseError(response.status);
  return team;
}
export async function createTournamentTeamInvitationRequest(tournamentID: string) {
  const response = await createTournamentTeamInvitation(
    tournamentID,
    undefined,
    authenticatedApiFetch,
  );
  if (response.status !== 201) throw new APIUnexpectedResponseError(response.status);
  const token = parseTournamentTeamInvitationToken(response.data);
  if (!token) throw new APIUnexpectedResponseError(response.status);
  return token;
}
export async function revokeTournamentTeamInvitationRequest(tournamentID: string) {
  const response = await revokeTournamentTeamInvitation(
    tournamentID,
    undefined,
    authenticatedApiFetch,
  );
  if (response.status !== 204) throw new APIUnexpectedResponseError(response.status);
}
export async function inspectTournamentTeamInvitationRequest(token: string) {
  const response = await inspectTournamentTeamInvitation(
    { token: token as TournamentTeamInvitationToken },
    undefined,
    apiFetch,
  );
  if (response.status === 404) throw new TournamentTeamInvitationUnavailableError();
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  const invitation = parseTournamentTeamInvitation(response.data);
  if (!invitation) throw new APIUnexpectedResponseError(response.status);
  return invitation;
}
export async function joinTournamentWithTeamInvitationRequest(token: string, name: string) {
  const response = await joinTournamentWithTeamInvitation(
    { token: token as TournamentTeamInvitationToken, name },
    undefined,
    authenticatedApiFetch,
  );
  if (response.status === 404) throw new TournamentTeamInvitationUnavailableError();
  if (response.status === 409) throw new TournamentTeamInvitationConflictError();
  if (response.status !== 201) throw new APIUnexpectedResponseError(response.status);
  const registration = parseTournamentTeamRegistration(response.data);
  if (!registration) throw new APIUnexpectedResponseError(response.status);
  return registration;
}
export async function removeTournamentTeamRequest(leagueID: string, teamID: string) {
  const response = await removeTournamentTeam(leagueID, teamID, undefined, authenticatedApiFetch);
  if (response.status !== 204) throw new APIUnexpectedResponseError(response.status);
}
export async function withdrawTournamentTeamRequest(leagueID: string, teamID: string) {
  const response = await withdrawTournamentTeam(leagueID, teamID, undefined, authenticatedApiFetch);
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  const league = parsePublicTournament(response.data);
  if (!league) throw new APIUnexpectedResponseError(response.status);
  return league;
}
export async function startTournamentRequest(leagueID: string, input: StartTournamentRequest) {
  const response = await startTournament(leagueID, input, undefined, authenticatedApiFetch);
  if (response.status === 422) throw new TournamentConfigurationConflictError();
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  const league = parsePublicTournament(response.data);
  if (!league) throw new APIUnexpectedResponseError(response.status);
  captureProductOutcome("league_started", response.headers);
  return league;
}
export async function startTournamentEliminationRequest(tournamentID: string) {
  const response = await startTournamentElimination(tournamentID, undefined, authenticatedApiFetch);
  if (response.status === 409) throw new TournamentStageTransitionConflictError();
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  const tournament = parsePublicTournament(response.data);
  if (!tournament) throw new APIUnexpectedResponseError(response.status);
  return tournament;
}
export async function cancelTournamentRequest(leagueID: string) {
  const response = await cancelTournament(leagueID, undefined, authenticatedApiFetch);
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  const league = parsePublicTournament(response.data);
  if (!league) throw new APIUnexpectedResponseError(response.status);
  return league;
}
export async function completeTournamentRequest(leagueID: string) {
  const response = await completeTournament(leagueID, undefined, authenticatedApiFetch);
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  const league = parsePublicTournament(response.data);
  if (!league) throw new APIUnexpectedResponseError(response.status);
  captureProductOutcome("league_completed", response.headers);
  return league;
}
export async function assignTournamentAdministratorRequest(leagueID: string, username: string) {
  const response = await assignTournamentAdministrator(
    leagueID,
    username,
    undefined,
    authenticatedApiFetch,
  );
  if (response.status === 409) throw new TournamentAdministratorConflictError();
  if (response.status !== 204) throw new APIUnexpectedResponseError(response.status);
  captureProductOutcome("tournament_administrator_assigned", response.headers);
}
export async function listTournamentAdministratorUsernames(leagueID: string) {
  const response = await listTournamentAdministrators(leagueID, undefined, authenticatedApiFetch);
  if (response.status === 404) throw new TournamentUnavailableError();
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  const usernames = parseUsernames(response.data);
  if (!usernames) throw new APIUnexpectedResponseError(response.status);
  return usernames;
}
export async function removeTournamentAdministratorRequest(leagueID: string, username: string) {
  const response = await removeTournamentAdministrator(
    leagueID,
    username as Username,
    undefined,
    authenticatedApiFetch,
  );
  if (response.status !== 204) throw new APIUnexpectedResponseError(response.status);
  captureProductOutcome("tournament_administrator_removed", response.headers);
}
export async function transferTournamentOwnershipRequest(leagueID: string, username: string) {
  const response = await transferTournamentOwnership(
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
export async function getTournament(leagueID: string) {
  const response = await getPublicTournament(leagueID, undefined, apiFetch);
  const status = (response as { status: number }).status;
  if (status === 200) {
    const league = parsePublicTournament((response as { data: unknown }).data);
    if (!league) throw new APIUnexpectedResponseError(status);
    return league;
  }
  if (status === 404) throw new TournamentUnavailableError();
  throw new APIUnexpectedResponseError(status);
}
export async function getTournamentRelationship(leagueID: string) {
  const response = await listCurrentAccountTournaments(
    { relationship: "administered", limit: 50 },
    undefined,
    authenticatedApiFetch,
  );
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  const items = parseAccountTournamentPageItems(response.data);
  if (!items) throw new APIUnexpectedResponseError(response.status);
  return items.find((league) => league.id === leagueID)?.relationship ?? null;
}
export async function listRelatedTournaments(relationship: "administered" | "followed") {
  const response = await listCurrentAccountTournaments(
    { relationship, limit: 50 },
    undefined,
    authenticatedApiFetch,
  );
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  const items = parseAccountTournamentPageItems(response.data);
  if (!items) throw new APIUnexpectedResponseError(response.status);
  return items;
}

export async function listRecentRelatedTournaments() {
  const response = await listRecentAccountTournaments(undefined, authenticatedApiFetch);
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  const leagues = parseRecentAccountTournaments(response.data);
  if (!leagues) throw new APIUnexpectedResponseError(response.status);
  return leagues;
}
