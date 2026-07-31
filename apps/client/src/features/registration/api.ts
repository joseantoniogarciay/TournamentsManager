import { APIUnexpectedResponseError, apiFetch, saveMobileSessionToken } from "@/api/fetch";
import {
  registerLocalAccount,
  verifyRegistration,
} from "@/api/generated/registration/registration";
import { getUsernameAvailability } from "@/api/generated/usernames/usernames";
import type { RegisterRequest } from "@/api/generated/models";

/** Adapta el cliente OpenAPI a las necesidades de la feature de registro. */
export function getRegistrationUsernameAvailability(username: string, signal: AbortSignal) {
  return getUsernameAvailability(username, { signal }, apiFetch);
}

export async function registerLocalAccountRequest(input: RegisterRequest) {
  const response = await registerLocalAccount(input, undefined, apiFetch);
  if (response.status !== 202) throw new APIUnexpectedResponseError(response.status);
}

export async function confirmRegistration(token: string, sessionTransport: "bearer" | "cookie") {
  const response = await verifyRegistration({ token, sessionTransport }, undefined, apiFetch);
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  if (sessionTransport === "bearer") {
    if (!response.data.sessionToken) throw new APIUnexpectedResponseError(response.status);
    await saveMobileSessionToken(response.data.sessionToken);
  }
  return response.data;
}
