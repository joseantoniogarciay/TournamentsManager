import AsyncStorage from "@react-native-async-storage/async-storage";

import type { TournamentInput } from "@/api/generated/models";

const key = "tm-league-draft";
export const maximumTournamentTeams = 64;
export const maximumTournamentNameLength = 56;
export type LocalTournamentDraft = { name: string; teams: string[] };

export async function getLocalTournamentDraft(): Promise<LocalTournamentDraft | null> {
  const serialized = await AsyncStorage.getItem(key);
  if (!serialized) return null;
  try {
    const value: unknown = JSON.parse(serialized);
    if (!value || typeof value !== "object") return null;
    const draft = value as Partial<LocalTournamentDraft>;
    return typeof draft.name === "string" &&
      Array.isArray(draft.teams) &&
      draft.teams.every((team) => typeof team === "string")
      ? { name: draft.name, teams: draft.teams }
      : null;
  } catch {
    return null;
  }
}
export function saveLocalTournamentDraft(draft: LocalTournamentDraft) {
  return AsyncStorage.setItem(key, JSON.stringify(draft));
}
export function clearLocalTournamentDraft() {
  return AsyncStorage.removeItem(key);
}

/** Convierte exclusivamente un borrador completo al contrato de alta/publicación. */
export function toTournamentInput(draft: LocalTournamentDraft | null): TournamentInput | undefined {
  if (!draft) return undefined;
  const name = draft.name.trim();
  const teams = draft.teams.map((team) => team.trim()).filter(Boolean);
  if (
    !name ||
    name.length > maximumTournamentNameLength ||
    teams.length < 2 ||
    teams.length > maximumTournamentTeams ||
    new Set(teams.map((team) => team.toLowerCase())).size !== teams.length ||
    teams.some((team) => team.length > 100)
  ) {
    return undefined;
  }
  return { name, teams: teams.map((name) => ({ name })) };
}
