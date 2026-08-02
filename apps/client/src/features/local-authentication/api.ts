import { APIUnexpectedResponseError, apiFetch, saveMobileSession } from "@/api/fetch";
import { createSession } from "@/api/generated/session/session";
import type { Transport } from "@/api/generated/models";

export type LocalAuthenticationResult =
  { kind: "pending-verification" } | { kind: "session"; user: { id: string; username: string } };

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
}): Promise<LocalAuthenticationResult> {
  const response = await createSession(input, undefined, apiFetch);
  if (response.status === 202) return { kind: "pending-verification" };
  if (response.status === 401) throw new LocalAuthenticationError();
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  if (input.sessionTransport === "bearer") await saveMobileSession(response.data);
  return { kind: "session", user: response.data.user };
}
