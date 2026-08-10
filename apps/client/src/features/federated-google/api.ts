import { APIUnexpectedResponseError, apiFetch, saveMobileSession } from "@/api/fetch";
import {
  createGoogleLoginChallenge,
  createGoogleSession,
} from "@/api/generated/federated-identity/federated-identity";
import type {
  GoogleLoginChallenge,
  LeagueInput,
  Locale,
  Transport,
  Username,
} from "@/api/generated/models";

export type GoogleAuthenticationFailure = "conflict" | "rate-limited";

/** Error de negocio recuperable del flujo de inicio de Google. */
export class GoogleAuthenticationError extends Error {
  constructor(readonly failure: GoogleAuthenticationFailure) {
    super(`Autenticación Google: ${failure}`);
    this.name = "GoogleAuthenticationError";
  }
}

/** Pide un nonce de un solo uso antes de abrir el proveedor externo. */
export async function beginGoogleAuthentication() {
  const response = await createGoogleLoginChallenge(undefined, apiFetch);
  if (response.status === 201) return response.data;
  if (response.status === 429) throw new GoogleAuthenticationError("rate-limited");
  throw new APIUnexpectedResponseError(response.status);
}

/** Entrega la prueba de Google al backend, que decide sesión, alta o conflicto. */
export async function finishGoogleAuthentication(input: {
  challenge: GoogleLoginChallenge;
  draft?: LeagueInput;
  idToken: string;
  locale?: Locale;
  sessionTransport: Transport;
  username?: Username;
}) {
  const response = await createGoogleSession(
    {
      challengeId: input.challenge.id,
      draft: input.draft,
      idToken: input.idToken,
      locale: input.locale,
      sessionTransport: input.sessionTransport,
      username: input.username,
    },
    undefined,
    apiFetch,
  );

  if (response.status === 202) return { kind: "username-required" as const };
  if (response.status === 409) throw new GoogleAuthenticationError("conflict");
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  if (input.sessionTransport === "bearer") await saveMobileSession(response.data);
  return { kind: "session" as const, session: response.data };
}
