import { APIUnexpectedResponseError, apiFetch, saveMobileSession } from "@/api/fetch";
import {
  registerLocalAccount,
  verifyRegistration,
} from "@/api/generated/registration/registration";
import { getUsernameAvailability } from "@/api/generated/usernames/usernames";
import type { RegisterRequest } from "@/api/generated/models";

export type RegistrationVerificationFailure = "already-used" | "expired";

/** Error de negocio que el contrato permite recuperar durante una verificación. */
export class RegistrationVerificationError extends Error {
  constructor(readonly failure: RegistrationVerificationFailure) {
    super(`Verificación de registro: ${failure}`);
    this.name = "RegistrationVerificationError";
  }
}

/** Adapta el cliente OpenAPI a las necesidades de la feature de registro. */
export function getRegistrationUsernameAvailability(username: string, signal: AbortSignal) {
  return getUsernameAvailability(username, { signal }, apiFetch);
}

export async function registerLocalAccountRequest(input: RegisterRequest) {
  const response = await registerLocalAccount(input, undefined, apiFetch);
  if (response.status !== 202) throw new APIUnexpectedResponseError(response.status);
}

export async function confirmRegistration(
  token: string,
  sessionTransport: "bearer" | "cookie",
  signal?: AbortSignal,
) {
  const response = await verifyRegistration({ token, sessionTransport }, { signal }, apiFetch);
  if (response.status === 409) throw new RegistrationVerificationError("already-used");
  if (response.status === 410) throw new RegistrationVerificationError("expired");
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  if (sessionTransport === "bearer") {
    await saveMobileSession(response.data);
  }
  return response.data;
}
