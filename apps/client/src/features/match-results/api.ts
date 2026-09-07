import {
  captureProductOutcome,
  APIUnexpectedResponseError,
  authenticatedApiFetch,
} from "@/api/fetch";
import { recordMatchResult } from "@/api/generated/tournaments/tournaments";
import type { MatchResultInput } from "@/api/generated/models";
import { parsePublicTournament } from "@/features/league-creation/response-parser";

export class MatchResultConflictError extends Error {}

export async function recordMatchResultRequest(
  leagueID: string,
  matchID: string,
  input: MatchResultInput,
) {
  const response = await recordMatchResult(
    leagueID,
    matchID,
    input,
    undefined,
    authenticatedApiFetch,
  );
  if (response.status === 409) throw new MatchResultConflictError();
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  const tournament = parsePublicTournament(response.data);
  if (!tournament) throw new APIUnexpectedResponseError(response.status);
  captureProductOutcome("match_result_recorded", response.headers);
  return tournament;
}
