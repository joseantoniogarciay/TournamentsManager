import * as SecureStore from "expo-secure-store";
import { Platform } from "react-native";

const defaultAPIBaseURL = "http://127.0.0.1:8080/v1";
const mobileSessionTokenKey = "tm-session-token";

/** La petición no llegó a recibir una respuesta HTTP de la API. */
export class APIConnectionError extends Error {
  constructor(cause: unknown) {
    super("No se pudo conectar con la API", { cause });
    this.name = "APIConnectionError";
  }
}

/** La API respondió con un estado que la feature no trata como útil para la persona. */
export class APIUnexpectedResponseError extends Error {
  constructor(status: number) {
    super(`La API respondió con el estado no tratado ${status}`);
    this.name = "APIUnexpectedResponseError";
  }
}

function getAPIBaseURL() {
  return (process.env.EXPO_PUBLIC_API_BASE_URL ?? defaultAPIBaseURL).replace(/\/$/, "");
}

function isAbsoluteURL(url: string) {
  return /^[a-z][a-z\d+.-]*:\/\//i.test(url);
}

/**
 * Transporte común del cliente OpenAPI.
 *
 * Las operaciones generadas aportan paths relativos; este borde añade el
 * origen configurable. Las credenciales de sesión se incorporarán aquí cuando
 * existan, sin obligar a cada feature a reconstruir la petición.
 */
export const apiFetch: typeof globalThis.fetch = async (input, init) => {
  const request =
    typeof input !== "string" || isAbsoluteURL(input)
      ? input
      : `${getAPIBaseURL()}${input.startsWith("/") ? input : `/${input}`}`;

  const headers = new Headers(init?.headers);
  if (Platform.OS !== "web") {
    const sessionToken = await SecureStore.getItemAsync(mobileSessionTokenKey);
    if (sessionToken) headers.set("Authorization", `Bearer ${sessionToken}`);
  }

  try {
    return await globalThis.fetch(request, { ...init, credentials: "include", headers });
  } catch (error: unknown) {
    if (error instanceof Error && error.name === "AbortError") throw error;
    throw new APIConnectionError(error);
  }
};

export async function saveMobileSessionToken(sessionToken: string) {
  if (Platform.OS !== "web") await SecureStore.setItemAsync(mobileSessionTokenKey, sessionToken);
}
