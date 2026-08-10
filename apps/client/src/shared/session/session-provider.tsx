import {
  clearMobileSession,
  getMobileSession,
  revokeCurrentSessionSilently,
  setMobileSessionInvalidationHandler,
} from "@/api/fetch";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import {
  createContext,
  type PropsWithChildren,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
} from "react";
import { StyleSheet, View } from "react-native";
import { Platform } from "react-native";

import { getTranslator } from "@/shared/i18n/locale";
import { LoadingTransition } from "@/shared/ui";
import { restoreWebSession } from "./api";

type SessionUser = { id: string; username: string };
export type SessionReplacementDestination = "/" | "/account" | "/create-tournament";
type SessionContextValue = {
  isRestoring: boolean;
  replacementDestination: SessionReplacementDestination;
  revision: number;
  user: SessionUser | null;
  transition: "idle" | "confirming" | "resetting" | "signing-out";
  beginSessionReplacement: () => void;
  completeSessionReplacement: (
    user: SessionUser,
    destination?: SessionReplacementDestination,
  ) => void;
  cancelSessionReplacement: () => void;
  finishSessionReplacement: () => void;
  signOut: () => Promise<void>;
  completeAccountDeletion: () => Promise<void>;
};

const SessionContext = createContext<SessionContextValue | null>(null);

/** Estado de sesión mínimo mientras se implementan lectura y cierre de sesión. */
export function SessionProvider({ children }: PropsWithChildren) {
  const t = getTranslator();
  const { show } = useFeedback();
  const [user, setUser] = useState<SessionUser | null>(null);
  const [isRestoring, setIsRestoring] = useState(true);
  const [revision, setRevision] = useState(0);
  const [replacementDestination, setReplacementDestination] =
    useState<SessionReplacementDestination>("/");
  const [transition, setTransition] = useState<SessionContextValue["transition"]>("idle");
  const hasInvalidatedSession = useRef(false);

  const beginSessionReplacement = useCallback(() => setTransition("confirming"), []);
  const completeSessionReplacement = useCallback(
    (nextUser: SessionUser, destination: SessionReplacementDestination = "/") => {
      hasInvalidatedSession.current = false;
      setUser(nextUser);
      setRevision((current) => current + 1);
      setReplacementDestination(destination);
      setTransition("resetting");
    },
    [],
  );
  const cancelSessionReplacement = useCallback(() => setTransition("idle"), []);
  const finishSessionReplacement = useCallback(() => setTransition("idle"), []);
  const signOut = useCallback(async () => {
    const session = await getMobileSession();
    void revokeCurrentSessionSilently(session);
    await clearMobileSession();
    setUser(null);
    setRevision((current) => current + 1);
    setTransition("signing-out");
  }, []);
  const completeAccountDeletion = useCallback(async () => {
    await clearMobileSession();
    setUser(null);
    setRevision((current) => current + 1);
    setTransition("resetting");
  }, []);
  const resetInvalidSession = useCallback(async () => {
    if (hasInvalidatedSession.current) return;
    hasInvalidatedSession.current = true;
    await clearMobileSession();
    setUser(null);
    setRevision((current) => current + 1);
    setTransition("resetting");
    show({ kind: "generic-error", message: t("session_expired") });
  }, [show, t]);

  useEffect(() => {
    if (Platform.OS === "web") {
      void restoreWebSession()
        .then((session) => {
          if (session) setUser(session);
        })
        .catch(() => undefined)
        .finally(() => setIsRestoring(false));
      return;
    }
    void getMobileSession()
      .then(async (session) => {
        if (!session) return;
        const refreshExpiresAt = Date.parse(session.refreshExpiresAt);
        if (!Number.isFinite(refreshExpiresAt) || refreshExpiresAt <= Date.now()) {
          await clearMobileSession();
          return;
        }
        setUser(session.user);
      })
      .finally(() => setIsRestoring(false));
  }, []);

  useEffect(() => setMobileSessionInvalidationHandler(resetInvalidSession), [resetInvalidSession]);

  return (
    <SessionContext.Provider
      value={{
        revision,
        user,
        isRestoring,
        replacementDestination,
        transition,
        beginSessionReplacement,
        completeSessionReplacement,
        cancelSessionReplacement,
        finishSessionReplacement,
        signOut,
        completeAccountDeletion,
      }}
    >
      <View style={styles.root}>
        {children}
        <LoadingTransition
          active={transition !== "idle"}
          message={t("link_confirmation_loading")}
        />
      </View>
    </SessionContext.Provider>
  );
}

export function useSession() {
  const value = useContext(SessionContext);
  if (!value) throw new Error("useSession debe usarse dentro de SessionProvider");
  return value;
}

const styles = StyleSheet.create({ root: { flex: 1 } });
