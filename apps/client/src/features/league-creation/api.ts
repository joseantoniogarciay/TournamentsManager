import { APIUnexpectedResponseError, apiFetch, authenticatedApiFetch } from "@/api/fetch";
import {
  assignLeagueAdministrator,
  cancelLeague,
  createLeague,
  getPublicLeague,
  listCurrentAccountLeagues,
  listRecentAccountLeagues,
  startLeague,
} from "@/api/generated/leagues/leagues";
import type { LeagueInput, StartLeagueRequest } from "@/api/generated/models";

export async function createLeagueRequest(input: LeagueInput) {
  const response = await createLeague(input, undefined, authenticatedApiFetch);
  if (response.status !== 201) throw new APIUnexpectedResponseError(response.status);
  return response.data;
}
export async function startLeagueRequest(leagueID: string, input: StartLeagueRequest) {
  const response = await startLeague(leagueID, input, undefined, authenticatedApiFetch);
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  return response.data;
}
export async function cancelLeagueRequest(leagueID: string) {
  const response = await cancelLeague(leagueID, undefined, authenticatedApiFetch);
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  return response.data;
}
export async function assignLeagueAdministratorRequest(leagueID: string, username: string) {
  const response = await assignLeagueAdministrator(
    leagueID,
    username,
    undefined,
    authenticatedApiFetch,
  );
  if (response.status !== 204) throw new APIUnexpectedResponseError(response.status);
}
export async function getLeague(leagueID: string) {
  const response = await getPublicLeague(leagueID, undefined, apiFetch);
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  return response.data;
}
export async function getLeagueRelationship(leagueID: string) {
  const response = await listCurrentAccountLeagues(
    { relationship: "administered", limit: 50 },
    undefined,
    authenticatedApiFetch,
  );
  if (response.status !== 200) return undefined;
  return response.data.items.find((league) => league.id === leagueID)?.relationship;
}
export async function listRelatedLeagues(relationship: "administered" | "followed") {
  const response = await listCurrentAccountLeagues(
    { relationship, limit: 50 },
    undefined,
    authenticatedApiFetch,
  );
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  return response.data.items;
}

export async function listRecentRelatedLeagues() {
  const response = await listRecentAccountLeagues(undefined, authenticatedApiFetch);
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  return response.data;
}
