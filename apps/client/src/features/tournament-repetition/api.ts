import { APIUnexpectedResponseError, authenticatedApiFetch } from "@/api/fetch";
import { repeatTournament } from "@/api/generated/tournaments/tournaments";
import type { RepeatTournamentInput } from "@/api/generated/models";
import { parsePublicTournament } from "@/features/league-creation/response-parser";

export class RepetitionRejectedError extends Error {
  constructor(readonly reason: "forbidden" | "unavailable" | "not_completed") {
    super("Tournament repetition rejected");
  }
}

export async function repeatTournamentRequest(
  id: string,
  input: RepeatTournamentInput,
  signal: AbortSignal,
) {
  const response = await repeatTournament(id, input, { signal }, authenticatedApiFetch);
  if (response.status !== 201) {
    const problem = response.data as unknown;
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
  if (response.status === 403) throw new RepetitionRejectedError("forbidden");
  if (response.status === 404) throw new RepetitionRejectedError("unavailable");
  if (response.status === 409) throw new RepetitionRejectedError("not_completed");
  if (response.status !== 201) throw new APIUnexpectedResponseError(response.status);
  const value = parsePublicTournament(response.data);
  if (!value) throw new APIUnexpectedResponseError(response.status);
  return value;
}
