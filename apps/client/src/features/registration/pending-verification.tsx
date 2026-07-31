import { createContext, type PropsWithChildren, useContext, useState } from "react";

const PendingVerificationContext = createContext<{
  token: string | null;
  setToken: (token: string | null) => void;
} | null>(null);

/** Conserva el token únicamente en memoria tras retirarlo de la URL e historial. */
export function PendingVerificationProvider({ children }: PropsWithChildren) {
  const [token, setToken] = useState<string | null>(null);
  return (
    <PendingVerificationContext.Provider value={{ token, setToken }}>
      {children}
    </PendingVerificationContext.Provider>
  );
}

export function usePendingVerification() {
  const value = useContext(PendingVerificationContext);
  if (!value)
    throw new Error("usePendingVerification debe usarse dentro de PendingVerificationProvider");
  return value;
}
