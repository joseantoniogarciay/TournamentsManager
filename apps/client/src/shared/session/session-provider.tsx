import {
  clearMobileSession,
  getMobileSession,
  setMobileSessionInvalidationHandler,
} from "@/api/fetch";
import {
  createContext,
  type PropsWithChildren,
  useCallback,
  useContext,
  useEffect,
  useState,
} from "react";
import { StyleSheet, View } from "react-native";

import { getTranslator } from "@/shared/i18n/locale";
import { LoadingTransition } from "@/shared/ui";

type SessionUser = { id: string; username: string };
type SessionContextValue = {
  isRestoring: boolean;
  revision: number;
  user: SessionUser | null;
  transition: "idle" | "confirming" | "resetting";
  beginSessionReplacement: () => void;
  completeSessionReplacement: (user: SessionUser) => void;
  cancelSessionReplacement: () => void;
  finishSessionReplacement: () => void;
};

const SessionContext = createContext<SessionContextValue | null>(null);

/** Estado de sesión mínimo mientras se implementan lectura y cierre de sesión. */
export function SessionProvider({ children }: PropsWithChildren) {
  const t = getTranslator();
  const [user, setUser] = useState<SessionUser | null>(null);
  const [isRestoring, setIsRestoring] = useState(true);
  const [revision, setRevision] = useState(0);
  const [transition, setTransition] = useState<SessionContextValue["transition"]>("idle");

  const beginSessionReplacement = useCallback(() => setTransition("confirming"), []);
  const completeSessionReplacement = useCallback((nextUser: SessionUser) => {
    setUser(nextUser);
    setRevision((current) => current + 1);
    setTransition("resetting");
  }, []);
  const cancelSessionReplacement = useCallback(() => setTransition("idle"), []);
  const finishSessionReplacement = useCallback(() => setTransition("idle"), []);
  const resetInvalidSession = useCallback(async () => {
    await clearMobileSession();
    setUser(null);
    setRevision((current) => current + 1);
    setTransition("resetting");
  }, []);

  useEffect(() => {
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
        transition,
        beginSessionReplacement,
        completeSessionReplacement,
        cancelSessionReplacement,
        finishSessionReplacement,
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
