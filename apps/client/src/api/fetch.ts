import * as SecureStore from "expo-secure-store";
import { Platform } from "react-native";

import type { SessionEstablishment, User } from "./generated/models";
import { refreshSession } from "./generated/session/session";

const defaultAPIBaseURL = "http://127.0.0.1:8080/v1";
const mobileSessionKey = "tm-mobile-session";
const refreshThresholdMilliseconds = 60 * 60 * 1000;

type MobileSession = {
  accessToken: string;
  expiresAt: string;
  refreshExpiresAt: string;
  refreshToken: string;
  user: User;
};

let refreshingSession: Promise<MobileSession | null> | null = null;
let invalidateMobileSession: (() => Promise<void>) | null = null;

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

function getRequestURL(input: RequestInfo | URL) {
  return typeof input !== "string" || isAbsoluteURL(input)
    ? input
    : `${getAPIBaseURL()}${input.startsWith("/") ? input : `/${input}`}`;
}

function isDateAfter(date: string, milliseconds: number) {
  const timestamp = Date.parse(date);
  return Number.isFinite(timestamp) && timestamp > milliseconds;
}

function isMobileSession(value: unknown): value is MobileSession {
  if (!value || typeof value !== "object") return false;
  const session = value as Partial<MobileSession>;
  return (
    typeof session.accessToken === "string" &&
    typeof session.refreshToken === "string" &&
    typeof session.expiresAt === "string" &&
    typeof session.refreshExpiresAt === "string" &&
    !!session.user &&
    typeof session.user.id === "string" &&
    typeof session.user.username === "string"
  );
}

async function fetchWithAPIBase(input: RequestInfo | URL, init?: RequestInit) {
  try {
    return await globalThis.fetch(getRequestURL(input), { ...init, credentials: "include" });
  } catch (error: unknown) {
    if (error instanceof Error && error.name === "AbortError") throw error;
    throw new APIConnectionError(error);
  }
}

/** Lee la sesión local; el perfil y las fechas son solo estado de arranque. */
export async function getMobileSession() {
  if (Platform.OS === "web") return null;
  const serialized = await SecureStore.getItemAsync(mobileSessionKey);
  if (!serialized) return null;

  try {
    const session: unknown = JSON.parse(serialized);
    return isMobileSession(session) ? session : null;
  } catch {
    return null;
  }
}

/** Guarda los dos secretos y sus metadatos juntos para no restaurar una sesión parcial. */
export async function saveMobileSession(session: SessionEstablishment) {
  if (Platform.OS === "web") return;
  if (
    session.delivery !== "bearer" ||
    !session.sessionToken ||
    !session.refreshToken ||
    !isDateAfter(session.expiresAt, Date.now()) ||
    !isDateAfter(session.refreshExpiresAt, Date.now())
  ) {
    throw new APIUnexpectedResponseError(200);
  }

  const mobileSession: MobileSession = {
    accessToken: session.sessionToken,
    expiresAt: session.expiresAt,
    refreshExpiresAt: session.refreshExpiresAt,
    refreshToken: session.refreshToken,
    user: session.user,
  };
  await SecureStore.setItemAsync(mobileSessionKey, JSON.stringify(mobileSession));
}

export async function clearMobileSession() {
  if (Platform.OS !== "web") await SecureStore.deleteItemAsync(mobileSessionKey);
}

/** Registra el único coordinador autorizado para borrar una sesión inválida. */
export function setMobileSessionInvalidationHandler(handler: () => Promise<void>) {
  invalidateMobileSession = handler;
  return () => {
    if (invalidateMobileSession === handler) invalidateMobileSession = null;
  };
}

async function expireMobileSession() {
  if (invalidateMobileSession) await invalidateMobileSession();
}

async function refreshMobileSession(session: MobileSession) {
  if (!isDateAfter(session.refreshExpiresAt, Date.now())) {
    await expireMobileSession();
    return null;
  }

  if (!refreshingSession) {
    refreshingSession = (async () => {
      const response = await refreshSession(undefined, (input, init) => {
        const headers = new Headers(init?.headers);
        headers.set("Authorization", `Bearer ${session.refreshToken}`);
        return fetchWithAPIBase(input, { ...init, headers });
      });
      if (response.status === 401) {
        await expireMobileSession();
        return null;
      }
      await saveMobileSession(response.data);
      return getMobileSession();
    })().finally(() => {
      refreshingSession = null;
    });
  }

  return refreshingSession;
}

async function getMobileAccessToken() {
  const session = await getMobileSession();
  if (!session) return null;
  if (isDateAfter(session.expiresAt, Date.now() + refreshThresholdMilliseconds)) {
    return session.accessToken;
  }
  const refreshedSession = await refreshMobileSession(session);
  return refreshedSession?.accessToken ?? null;
}

/**
 * Transporte común del cliente OpenAPI. Las operaciones generadas aportan paths
 * relativos; este borde añade el origen y la credencial móvil vigente.
 */
export const apiFetch: typeof globalThis.fetch = async (input, init) => {
  const headers = new Headers(init?.headers);
  if (Platform.OS !== "web") {
    const accessToken = await getMobileAccessToken();
    if (accessToken) headers.set("Authorization", `Bearer ${accessToken}`);
  }
  return fetchWithAPIBase(input, { ...init, headers });
};
