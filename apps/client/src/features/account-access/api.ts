import {
  createReauthenticationTicket,
  getAccessMethods,
  putLocalCredential,
} from "@/api/generated/session/session";
import { createCurrentAccountGoogleIdentity } from "@/api/generated/federated-identity/federated-identity";
import { apiFetch, APIUnexpectedResponseError } from "@/api/fetch";

export async function getAccountAccessMethods() {
  const response = await getAccessMethods(undefined, apiFetch);
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  return response.data;
}

export async function setAccountPassword(currentPassword: string, password: string) {
  const ticket = await createReauthenticationTicket(
    { password: currentPassword },
    undefined,
    apiFetch,
  );
  if (ticket.status !== 201) throw new APIUnexpectedResponseError(ticket.status);
  const response = await putLocalCredential(
    { password, ticket: ticket.data.ticket },
    undefined,
    apiFetch,
  );
  if (response.status !== 204) throw new APIUnexpectedResponseError(response.status);
}

export class GoogleLinkError extends Error {
  constructor(readonly reason: "conflict" | "expired") {
    super(reason);
  }
}

export async function reauthenticateWithPassword(password: string) {
  const response = await createReauthenticationTicket({ password }, undefined, apiFetch);
  if (response.status !== 201) throw new GoogleLinkError("expired");
  return response.data.ticket;
}

export async function reauthenticateWithGoogle(challengeId: string, idToken: string) {
  const response = await createReauthenticationTicket(
    { challengeId, idToken },
    undefined,
    apiFetch,
  );
  if (response.status !== 201) throw new GoogleLinkError("expired");
  return response.data.ticket;
}

export async function linkGoogle(ticket: string, challengeId: string, idToken: string) {
  const response = await createCurrentAccountGoogleIdentity(
    { ticket, challengeId, idToken },
    undefined,
    apiFetch,
  );
  if (response.status === 204) return;
  if (response.status === 409) throw new GoogleLinkError("conflict");
  throw new GoogleLinkError("expired");
}
