import { isSocialSession } from "@/features/social-authentication/session";
import {
  APIUnexpectedResponseError,
  apiFetch,
  captureProductOutcome,
  saveMobileSession,
} from "@/api/fetch";
import {
  createAppleLoginChallenge,
  createAppleSession,
} from "@/api/generated/federated-identity/federated-identity";
import type {
  AppleAuthenticationRequest,
  AppleLoginChallengeRequest,
} from "@/api/generated/models";

export class AppleAuthenticationError extends Error {
  constructor(readonly failure: "conflict" | "rate-limited") {
    super(`Apple authentication: ${failure}`);
    this.name = "AppleAuthenticationError";
  }
}

export async function beginAppleAuthentication(input: AppleLoginChallengeRequest) {
  const response = await createAppleLoginChallenge(input, undefined, apiFetch);
  if (response.status === 429) throw new AppleAuthenticationError("rate-limited");
  if (response.status !== 201) throw new APIUnexpectedResponseError(response.status);
  if (
    !response.data ||
    typeof response.data.id !== "string" ||
    typeof response.data.authorizationUrl !== "string" ||
    typeof response.data.returnUrl !== "string" ||
    typeof response.data.expiresAt !== "string" ||
    !(Date.parse(response.data.expiresAt) > Date.now())
  )
    throw new APIUnexpectedResponseError(response.status);
  return response.data;
}

export async function finishAppleAuthentication(input: AppleAuthenticationRequest) {
  const response = await createAppleSession(input, undefined, apiFetch);
  if (response.status === 202) return { kind: "username-required" as const };
  if (response.status === 409) throw new AppleAuthenticationError("conflict");
  if (response.status === 429) throw new AppleAuthenticationError("rate-limited");
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  if (!isSocialSession(response.data, input.sessionTransport))
    throw new APIUnexpectedResponseError(response.status);
  if (input.sessionTransport === "bearer") await saveMobileSession(response.data);
  captureProductOutcome("account_signed_in", response.headers, { method: "apple" });
  return { kind: "session" as const, session: response.data };
}
