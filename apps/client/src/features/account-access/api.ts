import {
  createReauthenticationTicket,
  deleteCurrentAccountLocalCredential,
  getAccessMethods,
  putLocalCredential,
  scheduleAccountDeletion,
} from "@/api/generated/session/session";
import {
  createCurrentAccountGoogleIdentity,
  deleteCurrentAccountGoogleIdentity,
} from "@/api/generated/federated-identity/federated-identity";
import { apiFetch, APIUnexpectedResponseError } from "@/api/fetch";

export async function getAccountAccessMethods() {
  const response = await getAccessMethods(undefined, apiFetch);
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  return response.data;
}

export class AccountDeletionError extends Error {
  constructor(readonly reason: "owned-leagues" | "unexpected") {
    super(reason);
  }
}

export async function deleteAccount() {
  const response = await scheduleAccountDeletion(undefined, apiFetch);
  if (response.status === 200) return response.data;
  if (response.status === 409) throw new AccountDeletionError("owned-leagues");
  throw new AccountDeletionError("unexpected");
}

export type ReauthenticationPurpose =
  "set-local-password" | "link-google" | "unlink-google" | "remove-local-password";

export async function setAccountPassword(ticket: string, password: string) {
  const response = await putLocalCredential({ password, ticket }, undefined, apiFetch);
  if (response.status !== 204) throw new APIUnexpectedResponseError(response.status);
}

export async function reauthenticateWithPassword(
  password: string,
  purpose: ReauthenticationPurpose,
) {
  const ticket = await createReauthenticationTicket({ password, purpose }, undefined, apiFetch);
  if (ticket.status !== 201) throw new APIUnexpectedResponseError(ticket.status);
  return ticket.data.ticket;
}

export class GoogleLinkError extends Error {
  constructor(readonly reason: "conflict" | "expired") {
    super(reason);
  }
}

export async function reauthenticateWithGoogle(
  challengeId: string,
  idToken: string,
  purpose: ReauthenticationPurpose,
) {
  const response = await createReauthenticationTicket(
    { challengeId, idToken, purpose },
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

export async function unlinkGoogle(ticket: string) {
  const response = await deleteCurrentAccountGoogleIdentity({ ticket }, undefined, apiFetch);
  if (response.status !== 204) throw new GoogleLinkError("expired");
}

export async function removeAccountPassword(ticket: string) {
  const response = await deleteCurrentAccountLocalCredential({ ticket }, undefined, apiFetch);
  if (response.status !== 204) throw new GoogleLinkError("expired");
}
