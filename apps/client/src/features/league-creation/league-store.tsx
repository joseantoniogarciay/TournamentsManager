import {
  createContext,
  type PropsWithChildren,
  useCallback,
  useContext,
  useMemo,
  useRef,
} from "react";
import { useSyncExternalStore } from "react";

import type { PublicTournament } from "@/api/generated/models";

import { getTournament } from "./api";

type Listener = () => void;

class TournamentStore {
  private readonly leagues = new Map<string, PublicTournament>();
  private readonly listeners = new Map<string, Set<Listener>>();
  private readonly loading = new Map<string, Promise<PublicTournament>>();

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

  put(league: PublicTournament) {
    this.leagues.set(league.id, league);
    this.listeners.get(league.id)?.forEach((listener) => listener());
  }

  update(id: string, updater: (league: PublicTournament) => PublicTournament) {
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
    const request = getTournament(id)
      .then((league) => {
        this.put(league);
        return league;
      })
      .finally(() => this.loading.delete(id));
    this.loading.set(id, request);
    return request;
  }
}

type TournamentStoreValue = {
  loadTournament: (id: string) => Promise<PublicTournament>;
  putTournament: (league: PublicTournament) => void;
  refreshTournament: (id: string) => Promise<PublicTournament>;
  updateTournament: (id: string, updater: (league: PublicTournament) => PublicTournament) => void;
  store: TournamentStore;
};

const TournamentStoreContext = createContext<TournamentStoreValue | null>(null);

export function TournamentStoreProvider({ children }: PropsWithChildren) {
  const store = useRef(new TournamentStore()).current;
  const loadTournament = useCallback((id: string) => store.load(id), [store]);
  const refreshTournament = useCallback((id: string) => store.load(id, true), [store]);
  const putTournament = useCallback((league: PublicTournament) => store.put(league), [store]);
  const updateTournament = useCallback(
    (id: string, updater: (league: PublicTournament) => PublicTournament) =>
      store.update(id, updater),
    [store],
  );
  const value = useMemo(
    () => ({ loadTournament, putTournament, refreshTournament, store, updateTournament }),
    [loadTournament, putTournament, refreshTournament, store, updateTournament],
  );
  return (
    <TournamentStoreContext.Provider value={value}>{children}</TournamentStoreContext.Provider>
  );
}

function useTournamentStoreContext() {
  const context = useContext(TournamentStoreContext);
  if (!context) throw new Error("TournamentStoreProvider is required");
  return context;
}

export function useTournamentStore() {
  const { loadTournament, putTournament, refreshTournament, updateTournament } =
    useTournamentStoreContext();
  return { loadTournament, putTournament, refreshTournament, updateTournament };
}

export function useTournament(id: string | undefined) {
  const { store } = useTournamentStoreContext();
  return useSyncExternalStore(
    useCallback((listener: Listener) => store.subscribe(id, listener), [id, store]),
    useCallback(() => store.get(id), [id, store]),
    () => undefined,
  );
}

export function useTournamentState(id: string, fallback: PublicTournament["state"]) {
  return useTournament(id)?.state ?? fallback;
}
