import Constants from "expo-constants";
import * as Crypto from "expo-crypto";
import * as WebBrowser from "expo-web-browser";
import { useCallback, useEffect, useRef, useState } from "react";
import { Platform } from "react-native";

import type {
  AppleLoginChallenge,
  Locale,
  TournamentDraftInput,
  User,
  Username,
} from "@/api/generated/models";
import {
  beginAppleAuthentication,
  finishAppleAuthentication,
} from "@/features/federated-apple/api";
import { appleReturnStatus } from "@/features/federated-apple/return";
import { configuredAppleServiceID } from "@/features/social-authentication/config";
import { androidAuthBrowserOptions } from "@/features/social-authentication/browser";

type Prepared = { challenge: AppleLoginChallenge; proof: string };
const serviceID = configuredAppleServiceID(process.env.EXPO_PUBLIC_APPLE_SERVICE_ID);

export function useAppleAuthentication({
  draft,
  locale,
  onSession,
}: {
  draft?: TournamentDraftInput;
  locale: Locale;
  onSession: (user: User, createdTournament: boolean) => void;
}) {
  const [prepared, setPrepared] = useState<Prepared | null>(null);
  const [pendingAccount, setPendingAccount] = useState<Prepared | null>(null);
  const [error, setError] = useState<unknown>(null);
  const [isPreparing, setIsPreparing] = useState(false);
  const [isAuthenticating, setIsAuthenticating] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const busy = useRef(false);
  const preparing = useRef(false);
  const environment = Constants.expoConfig?.extra?.appEnvironment;
  const isConfigured = Boolean(serviceID) && environment !== "local";

  const prepare = useCallback(
    async ({ reportFailure = false }: { reportFailure?: boolean } = {}) => {
      if (!isConfigured || preparing.current || busy.current) return;
      preparing.current = true;
      setIsPreparing(true);
      try {
        const bytes = await Crypto.getRandomBytesAsync(32);
        const proof = btoa(String.fromCharCode(...bytes))
          .replace(/\+/g, "-")
          .replace(/\//g, "_")
          .replace(/=+$/, "");
        const proofChallenge = await Crypto.digestStringAsync(
          Crypto.CryptoDigestAlgorithm.SHA256,
          `apple-login-proof:${proof}`,
        );
        const platform =
          Platform.OS === "ios" ? "ios" : Platform.OS === "android" ? "android" : "web";
        const challenge = await beginAppleAuthentication({ platform, proofChallenge });
        const expectedReturnURL =
          Platform.OS === "web"
            ? `${window.location.origin}/oauth/apple-complete`
            : `${environment === "production" ? "fasttourney" : "fasttourney-dev"}://oauth/apple-complete`;
        const authorization = new URL(challenge.authorizationUrl);
        if (
          challenge.returnUrl !== expectedReturnURL ||
          authorization.origin !== "https://appleid.apple.com" ||
          authorization.pathname !== "/auth/authorize" ||
          authorization.searchParams.get("client_id") !== serviceID
        ) {
          throw new Error("Apple authorization configuration mismatch");
        }
        setPrepared({ challenge, proof });
      } catch (nextError) {
        setPrepared(null);
        if (reportFailure) setError(nextError);
      } finally {
        preparing.current = false;
        setIsPreparing(false);
      }
    },
    [environment, isConfigured],
  );

  useEffect(() => {
    if (!prepared || pendingAccount || isAuthenticating) return;
    const timeout = setTimeout(
      () => {
        setPrepared(null);
      },
      Math.max(0, Date.parse(prepared.challenge.expiresAt) - Date.now() - 30_000),
    );
    return () => clearTimeout(timeout);
  }, [isAuthenticating, pendingAccount, prepared]);

  const establish = useCallback(
    async (input: Prepared, username?: Username) => {
      const result = await finishAppleAuthentication({
        challengeId: input.challenge.id,
        proof: input.proof,
        sessionTransport: Platform.OS === "web" ? "cookie" : "bearer",
        draft,
        username,
        locale: username ? locale : undefined,
        termsVersion: username ? "2026-08-22" : undefined,
      });
      if (result.kind === "username-required") {
        setPendingAccount(input);
        return;
      }
      setPrepared(null);
      setPendingAccount(null);
      onSession(result.session.user, Boolean(draft));
    },
    [draft, locale, onSession],
  );

  const start = useCallback(async () => {
    if (!isConfigured || preparing.current || busy.current) return;
    setError(null);
    if (!prepared || Date.parse(prepared.challenge.expiresAt) <= Date.now()) {
      await prepare({ reportFailure: true });
      return;
    }
    busy.current = true;
    setIsAuthenticating(true);
    try {
      // Invoke directly from the gesture: web popup, ASWebAuthenticationSession
      // on iOS and Custom Tabs when available on Android, otherwise its browser.
      const result = await WebBrowser.openAuthSessionAsync(
        prepared.challenge.authorizationUrl,
        prepared.challenge.returnUrl,
        Platform.OS === "android" ? await androidAuthBrowserOptions() : undefined,
      );
      if (result.type === "cancel" || result.type === "dismiss") return;
      if (result.type !== "success") throw new Error("Apple authentication failed");
      const status = appleReturnStatus(
        result.url,
        prepared.challenge.returnUrl,
        prepared.challenge.id,
      );
      if (status === "cancelled") return;
      if (status !== "ready") throw new Error("Apple authentication failed");
      await establish(prepared);
    } catch (nextError) {
      setError(nextError);
    } finally {
      busy.current = false;
      setIsAuthenticating(false);
      setPrepared(null);
    }
  }, [establish, isConfigured, prepare, prepared]);

  const chooseUsername = useCallback(
    async (username: Username) => {
      if (!pendingAccount || busy.current) return;
      busy.current = true;
      setIsSubmitting(true);
      try {
        await establish(pendingAccount, username);
      } catch (nextError) {
        setError(nextError);
      } finally {
        busy.current = false;
        setIsSubmitting(false);
      }
    },
    [establish, pendingAccount],
  );

  return {
    chooseUsername,
    dismissPendingAccount: () => {
      setPendingAccount(null);
      setPrepared(null);
    },
    dismissError: () => setError(null),
    error,
    isConfigured,
    isPreparing,
    isAuthenticating,
    isSubmitting,
    prepare,
    requiresUsername: Boolean(pendingAccount),
    start,
  };
}
