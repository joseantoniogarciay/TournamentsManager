import { createContext, type PropsWithChildren, useCallback, useContext, useState } from "react";
import { StyleSheet, View } from "react-native";

import { SessionTransition } from "./session-transition";

type SessionUser = { id: string; username: string };
type SessionContextValue = {
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
  const [user, setUser] = useState<SessionUser | null>(null);
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

  return (
    <SessionContext.Provider
      value={{
        revision,
        user,
        transition,
        beginSessionReplacement,
        completeSessionReplacement,
        cancelSessionReplacement,
        finishSessionReplacement,
      }}
    >
      <View style={styles.root}>
        {children}
        <SessionTransition active={transition !== "idle"} />
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
