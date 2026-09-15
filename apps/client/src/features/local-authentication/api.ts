import {
  captureProductOutcome,
  APIUnexpectedResponseError,
  apiFetch,
  saveMobileSession,
} from "@/api/fetch";
import { createSession } from "@/api/generated/session/session";
import type { TournamentDraftInput, Transport } from "@/api/generated/models";

export type LocalAuthenticationResult =
  | { kind: "pending-verification" }
  | {
      kind: "session";
      createdTournament: boolean;
      user: { id: string; username: string };
    };

/** Error recuperable: el contrato confirma que la autenticación fue rechazada. */
export class LocalAuthenticationError extends Error {
  constructor() {
    super("Credenciales locales no válidas");
    this.name = "LocalAuthenticationError";
  }
}

/** Autentica una cuenta local sin exponer detalles de una credencial rechazada. */
export async function authenticateLocalAccount(input: {
  email: string;
  password: string;
  sessionTransport: Transport;
  draft?: TournamentDraftInput;
}): Promise<LocalAuthenticationResult> {
  const response = await createSession(input, undefined, apiFetch);
  if (response.status === 202) return { kind: "pending-verification" };
  if (response.status === 401) throw new LocalAuthenticationError();
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  if (input.sessionTransport === "bearer") await saveMobileSession(response.data);
  captureProductOutcome("account_signed_in", response.headers, { method: "password" });
  return { kind: "session", createdTournament: Boolean(input.draft), user: response.data.user };
}
