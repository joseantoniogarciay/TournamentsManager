import { APIUnexpectedResponseError, authenticatedApiFetch } from "@/api/fetch";
import { createCurrentAccountSuggestion } from "@/api/generated/suggestions/suggestions";

export class SuggestionRateLimitedError extends Error {
  constructor() {
    super("Envío de sugerencias limitado");
  }
}

export async function submitSuggestion(body: string) {
  const response = await createCurrentAccountSuggestion({ body }, undefined, authenticatedApiFetch);
  if (response.status === 201) return;
  if (response.status === 429) throw new SuggestionRateLimitedError();
  throw new APIUnexpectedResponseError(response.status);
}
