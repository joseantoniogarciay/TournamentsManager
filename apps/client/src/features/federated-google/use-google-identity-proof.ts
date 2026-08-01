import * as Google from "expo-auth-session/providers/google";
import { makeRedirectUri } from "expo-auth-session";
import * as WebBrowser from "expo-web-browser";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Platform } from "react-native";

import type { GoogleLoginChallenge } from "@/api/generated/models";
import { beginGoogleAuthentication } from "@/features/federated-google/api";

WebBrowser.maybeCompleteAuthSession();

const clientIDs = {
  android: process.env.EXPO_PUBLIC_GOOGLE_ANDROID_CLIENT_ID,
  ios: process.env.EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID,
  web: process.env.EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID,
} as const;

function clientID() {
  return Platform.OS === "android"
    ? clientIDs.android
    : Platform.OS === "ios"
      ? clientIDs.ios
      : clientIDs.web;
}

/** Obtiene una única prueba OIDC para un flujo ya autenticado. */
export function useGoogleIdentityProof(
  onProof: (challenge: GoogleLoginChallenge, idToken: string) => Promise<void>,
) {
  const [challenge, setChallenge] = useState<GoogleLoginChallenge | null>(null);
  const [error, setError] = useState<unknown>(null);
  const [preparing, setPreparing] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const handled = useRef<unknown>(null);
  const id = clientID();
  const redirectUri =
    id && Platform.OS !== "web"
      ? makeRedirectUri({
          native: `com.googleusercontent.apps.${id.replace(/\.apps\.googleusercontent\.com$/, "")}:/oauthredirect`,
        })
      : undefined;
  const config = useMemo(
    () => ({
      androidClientId: clientIDs.android,
      clientId: id ?? "google-client-not-configured",
      extraParams: challenge ? { nonce: challenge.nonce } : undefined,
      iosClientId: clientIDs.ios,
      redirectUri,
      selectAccount: true,
      webClientId: clientIDs.web,
    }),
    [challenge, id, redirectUri],
  );
  const [request, response, promptAsync] = Google.useIdTokenAuthRequest(config);

  const prepare = useCallback(async () => {
    setPreparing(true);
    setError(null);
    try {
      setChallenge(await beginGoogleAuthentication());
    } catch (nextError) {
      setError(nextError);
    } finally {
      setPreparing(false);
    }
  }, []);

  useEffect(() => {
    if (response?.type !== "success" || !challenge || response === handled.current) return;
    handled.current = response;
    const token = response.params.id_token ?? response.authentication?.idToken;
    if (!token) {
      setError(new Error("missing Google ID token"));
      return;
    }
    setSubmitting(true);
    void onProof(challenge, token)
      .catch(setError)
      .finally(() => {
        setSubmitting(false);
        setChallenge(null);
      });
  }, [challenge, onProof, response]);

  const start = useCallback(async () => {
    if (!id || preparing || submitting) return;
    if (!challenge || !request) {
      await prepare();
      return;
    }
    try {
      const result = await promptAsync();
      if (result.type !== "success") setError(new Error("Google authentication was cancelled"));
    } catch (nextError) {
      setError(nextError);
    }
  }, [challenge, id, prepare, preparing, promptAsync, request, submitting]);

  return { error, isConfigured: Boolean(id), isLoading: preparing || submitting, prepare, start };
}
