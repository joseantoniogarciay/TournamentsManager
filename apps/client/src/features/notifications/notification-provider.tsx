import {
  createContext,
  type PropsWithChildren,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
} from "react";
import { AppState } from "react-native";

import { useSession } from "@/shared/session/session-provider";

import { unreadNotificationCount } from "./api";

const refreshInterval = 5 * 60 * 1000;
type Value = {
  count: number;
  refresh: (force?: boolean) => Promise<void>;
  refreshFromPush: () => Promise<void>;
};
const NotificationContext = createContext<Value | null>(null);

export function NotificationProvider({ children }: PropsWithChildren) {
  const { user, revision } = useSession();
  const [count, setCount] = useState(0);
  const lastRefreshAt = useRef(0);
  const refresh = useCallback(
    async (force = false) => {
      if (!user) {
        setCount(0);
        lastRefreshAt.current = 0;
        return;
      }
      if (!force && Date.now() - lastRefreshAt.current < refreshInterval) return;
      const next = await unreadNotificationCount();
      setCount(next);
      lastRefreshAt.current = Date.now();
    },
    [user],
  );
  useEffect(() => {
    void refresh(true).catch(() => setCount(0));
  }, [refresh, revision]);
  useEffect(() => {
    const subscription = AppState.addEventListener("change", (state) => {
      if (state === "active") void refresh(false).catch(() => undefined);
    });
    return () => subscription.remove();
  }, [refresh]);
  const refreshFromPush = useCallback(() => refresh(true), [refresh]);
  return (
    <NotificationContext.Provider value={{ count, refresh, refreshFromPush }}>
      {children}
    </NotificationContext.Provider>
  );
}
export function useNotifications() {
  const value = useContext(NotificationContext);
  if (!value) throw new Error("useNotifications debe usarse dentro de NotificationProvider");
  return value;
}
