import {
  captureProductOutcome,
  APIUnexpectedResponseError,
  authenticatedApiFetch,
} from "@/api/fetch";
import { recordMatchResult } from "@/api/generated/leagues/leagues";
import type { MatchResultInput } from "@/api/generated/models";

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
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  captureProductOutcome("match_result_recorded", response.headers);
  return response.data;
}
