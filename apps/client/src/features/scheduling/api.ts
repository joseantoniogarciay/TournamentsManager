import { authenticatedApiFetch, APIUnexpectedResponseError } from "@/api/fetch";
import { setMatchSchedule, setTournamentVenue } from "@/api/generated/tournaments/tournaments";
import type { MatchScheduleInput } from "@/api/generated/models";
import { parsePublicTournament } from "@/features/league-creation/response-parser";

export class SchedulingConflictError extends Error {}
export class SchedulingForbiddenError extends Error {}
export class SchedulingUnavailableError extends Error {}

function parseResponse(response: { status: number; data: unknown }) {
  if (response.status !== 200) {
    const problem = response.data;
    if (
      !problem ||
      typeof problem !== "object" ||
      !("type" in problem) ||
      problem.type !== "about:blank" ||
      !("status" in problem) ||
      problem.status !== response.status ||
      !("title" in problem) ||
      typeof problem.title !== "string"
    )
      throw new APIUnexpectedResponseError(response.status);
  }
  if (response.status === 403) throw new SchedulingForbiddenError();
  if (response.status === 404) throw new SchedulingUnavailableError();
  if (response.status === 409) throw new SchedulingConflictError();
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  const value = parsePublicTournament(response.data);
  if (!value) throw new APIUnexpectedResponseError(response.status);
  return value;
}

export async function saveTournamentVenue(id: string, venue: string, signal: AbortSignal) {
  return parseResponse(await setTournamentVenue(id, { venue }, { signal }, authenticatedApiFetch));
}

export async function saveMatchSchedule(
  id: string,
  matchId: string,
  input: MatchScheduleInput,
  signal: AbortSignal,
) {
  return parseResponse(
    await setMatchSchedule(id, matchId, input, { signal }, authenticatedApiFetch),
  );
}
