import {
  createContext,
  type PropsWithChildren,
  useCallback,
  useContext,
  useMemo,
  useRef,
} from "react";
import { useSyncExternalStore } from "react";

import type { PublicLeague } from "@/api/generated/models";

import { getLeague } from "./api";

type Listener = () => void;

class LeagueStore {
  private readonly leagues = new Map<string, PublicLeague>();
  private readonly listeners = new Map<string, Set<Listener>>();
  private readonly loading = new Map<string, Promise<PublicLeague>>();

  get(id: string | undefined) {
    return id ? this.leagues.get(id) : undefined;
  }

  subscribe(id: string | undefined, listener: Listener) {
    if (!id) return () => undefined;
    const listeners = this.listeners.get(id) ?? new Set<Listener>();
    listeners.add(listener);
    this.listeners.set(id, listeners);
    return () => {
      listeners.delete(listener);
      if (listeners.size === 0) this.listeners.delete(id);
    };
  }

  put(league: PublicLeague) {
    this.leagues.set(league.id, league);
    this.listeners.get(league.id)?.forEach((listener) => listener());
  }

  update(id: string, updater: (league: PublicLeague) => PublicLeague) {
    const league = this.leagues.get(id);
    if (league) this.put(updater(league));
  }

  load(id: string, force = false) {
    if (!force) {
      const league = this.leagues.get(id);
      if (league) return Promise.resolve(league);
    }
    const pending = this.loading.get(id);
    if (pending) return pending;
    const request = getLeague(id)
      .then((league) => {
        this.put(league);
        return league;
      })
      .finally(() => this.loading.delete(id));
    this.loading.set(id, request);
    return request;
  }
}

type LeagueStoreValue = {
  loadLeague: (id: string) => Promise<PublicLeague>;
  putLeague: (league: PublicLeague) => void;
  refreshLeague: (id: string) => Promise<PublicLeague>;
  updateLeague: (id: string, updater: (league: PublicLeague) => PublicLeague) => void;
  store: LeagueStore;
};

const LeagueStoreContext = createContext<LeagueStoreValue | null>(null);

export function LeagueStoreProvider({ children }: PropsWithChildren) {
  const store = useRef(new LeagueStore()).current;
  const loadLeague = useCallback((id: string) => store.load(id), [store]);
  const refreshLeague = useCallback((id: string) => store.load(id, true), [store]);
  const putLeague = useCallback((league: PublicLeague) => store.put(league), [store]);
  const updateLeague = useCallback(
    (id: string, updater: (league: PublicLeague) => PublicLeague) => store.update(id, updater),
    [store],
  );
  const value = useMemo(
    () => ({ loadLeague, putLeague, refreshLeague, store, updateLeague }),
    [loadLeague, putLeague, refreshLeague, store, updateLeague],
  );
  return <LeagueStoreContext.Provider value={value}>{children}</LeagueStoreContext.Provider>;
}

function useLeagueStoreContext() {
  const context = useContext(LeagueStoreContext);
  if (!context) throw new Error("LeagueStoreProvider is required");
  return context;
}

export function useLeagueStore() {
  const { loadLeague, putLeague, refreshLeague, updateLeague } = useLeagueStoreContext();
  return { loadLeague, putLeague, refreshLeague, updateLeague };
}

export function useLeague(id: string | undefined) {
  const { store } = useLeagueStoreContext();
  return useSyncExternalStore(
    useCallback((listener: Listener) => store.subscribe(id, listener), [id, store]),
    useCallback(() => store.get(id), [id, store]),
    () => undefined,
  );
}

export function useLeagueState(id: string, fallback: PublicLeague["state"]) {
  return useLeague(id)?.state ?? fallback;
}
