import { createContext, type PropsWithChildren, useContext, useState } from "react";

type SessionUser = { id: string; username: string };
type SessionContextValue = {
  revision: number;
  user: SessionUser | null;
  replaceSession: (user: SessionUser) => void;
};

const SessionContext = createContext<SessionContextValue | null>(null);

/** Estado de sesión mínimo mientras se implementan lectura y cierre de sesión. */
export function SessionProvider({ children }: PropsWithChildren) {
  const [user, setUser] = useState<SessionUser | null>(null);
  const [revision, setRevision] = useState(0);
  const replaceSession = (nextUser: SessionUser) => {
    setUser(nextUser);
    setRevision((current) => current + 1);
  };
  return (
    <SessionContext.Provider value={{ revision, user, replaceSession }}>
      {children}
    </SessionContext.Provider>
  );
}

export function useSession() {
  const value = useContext(SessionContext);
  if (!value) throw new Error("useSession debe usarse dentro de SessionProvider");
  return value;
}
