import * as Google from "expo-auth-session/providers/google";
import { makeRedirectUri } from "expo-auth-session";
import * as WebBrowser from "expo-web-browser";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Platform } from "react-native";

import type { GoogleLoginChallenge, Locale, Username } from "@/api/generated/models";

import {
  beginGoogleAuthentication,
  finishGoogleAuthentication,
} from "@/features/federated-google/api";

type PendingGoogleAccount = { challenge: GoogleLoginChallenge; idToken: string };

// Completa el popup de OAuth al volver a la web antes de que React monte la pantalla.
WebBrowser.maybeCompleteAuthSession();

const googleClientIDs = {
  android: process.env.EXPO_PUBLIC_GOOGLE_ANDROID_CLIENT_ID,
  ios: process.env.EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID,
  web: process.env.EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID,
} as const;

function getPlatformClientID() {
  if (Platform.OS === "android") return googleClientIDs.android;
  if (Platform.OS === "ios") return googleClientIDs.ios;
  return googleClientIDs.web;
}

function getGoogleRedirectURI(clientID: string | undefined) {
  if (!clientID || Platform.OS === "web") return undefined;
  const scheme = `com.googleusercontent.apps.${clientID.replace(/\.apps\.googleusercontent\.com$/, "")}`;
  return makeRedirectUri({ native: `${scheme}:/oauthredirect` });
}

/**
 * Coordina el nonce emitido por nuestra API con la prueba de Google de Expo.
 * La identidad externa nunca se usa para vincular cuentas por email en cliente.
 */
export function useGoogleAuthentication({
  locale,
  onSession,
}: {
  locale: Locale;
  onSession: (user: { id: string; username: string }) => void;
}) {
  const [challenge, setChallenge] = useState<GoogleLoginChallenge | null>(null);
  const [pendingAccount, setPendingAccount] = useState<PendingGoogleAccount | null>(null);
  const [error, setError] = useState<unknown>(null);
  const [isPreparing, setIsPreparing] = useState(false);
  const [isPrompting, setIsPrompting] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const handledResponse = useRef<unknown>(null);
  const configuredClientID = getPlatformClientID();
  const isConfigured = Boolean(configuredClientID);
  const redirectUri = getGoogleRedirectURI(configuredClientID);
  const sessionTransport = Platform.OS === "web" ? "cookie" : "bearer";

  const requestConfig = useMemo(
    () => ({
      androidClientId: googleClientIDs.android,
      clientId: configuredClientID ?? "google-client-not-configured",
      extraParams: challenge ? { nonce: challenge.nonce } : undefined,
      iosClientId: googleClientIDs.ios,
      redirectUri,
      selectAccount: true,
      webClientId: googleClientIDs.web,
    }),
    [challenge, configuredClientID, redirectUri],
  );
  const [request, response, promptAsync] = Google.useIdTokenAuthRequest(requestConfig);

  const loadChallenge = useCallback(async () => {
    setIsPreparing(true);
    try {
      setChallenge(await beginGoogleAuthentication());
    } catch (nextError) {
      setChallenge(null);
      setError(nextError);
    } finally {
      setIsPreparing(false);
    }
  }, []);

  // En web el popup debe abrirse desde el gesto de la persona. Precargamos el
  // nonce para que el botón pueda llamar promptAsync directamente.
  useEffect(() => {
    if (!isConfigured) return;
    void loadChallenge();
  }, [isConfigured, loadChallenge]);

  useEffect(() => {
    if (!challenge) return;
    const refreshIn = Math.max(0, Date.parse(challenge.expiresAt) - Date.now() - 30_000);
    const timeout = setTimeout(() => {
      void loadChallenge();
    }, refreshIn);
    return () => clearTimeout(timeout);
  }, [challenge, loadChallenge]);

  const establish = useCallback(
    async (input: PendingGoogleAccount, username?: Username) => {
      const result = await finishGoogleAuthentication({
        ...input,
        locale: username ? locale : undefined,
        sessionTransport,
        username,
      });
      if (result.kind === "username-required") {
        setPendingAccount(input);
        return;
      }
      setChallenge(null);
      setPendingAccount(null);
      onSession(result.session.user);
    },
    [locale, onSession, sessionTransport],
  );

  useEffect(() => {
    if (response?.type !== "success" || !challenge || response === handledResponse.current) return;
    handledResponse.current = response;
    const idToken = response.params.id_token ?? response.authentication?.idToken;
    setIsPrompting(false);
    if (!idToken) {
      setError(new Error("Google no devolvió un token de identidad."));
      return;
    }

    setIsSubmitting(true);
    void establish({ challenge, idToken })
      .catch((nextError) => {
        setError(nextError);
        setChallenge(null);
        void loadChallenge();
      })
      .finally(() => setIsSubmitting(false));
  }, [challenge, establish, loadChallenge, response]);

  const start = useCallback(async () => {
    if (!isConfigured || isPreparing || isPrompting || isSubmitting) return;
    setError(null);
    // Un popup web debe abrirse directamente desde el gesto. Si el challenge o
    // la petición aún no están listos, este toque solo los prepara y el siguiente
    // podrá abrir Google sin que el navegador lo bloquee.
    if (!challenge || !request) {
      await loadChallenge();
      return;
    }
    setIsPrompting(true);
    try {
      const result = await promptAsync();
      if (result.type !== "success") setIsPrompting(false);
    } catch (nextError) {
      setError(nextError);
      setIsPrompting(false);
    }
  }, [
    challenge,
    isConfigured,
    isPreparing,
    isPrompting,
    isSubmitting,
    loadChallenge,
    promptAsync,
    request,
  ]);

  const chooseUsername = useCallback(
    async (username: Username) => {
      if (!pendingAccount || isSubmitting) return;
      setIsSubmitting(true);
      try {
        await establish(pendingAccount, username);
      } catch (nextError) {
        setError(nextError);
      } finally {
        setIsSubmitting(false);
      }
    },
    [establish, isSubmitting, pendingAccount],
  );

  return {
    chooseUsername,
    dismissError: () => setError(null),
    error,
    isConfigured,
    isPreparing,
    isAuthenticating: isPrompting || isSubmitting,
    isSubmitting,
    requiresUsername: Boolean(pendingAccount),
    start,
  };
}
