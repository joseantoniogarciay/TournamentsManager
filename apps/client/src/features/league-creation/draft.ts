import AsyncStorage from "@react-native-async-storage/async-storage";

import type { LeagueInput } from "@/api/generated/models";

const key = "tm-league-draft";
export const maximumLeagueTeams = 64;
export type LocalLeagueDraft = { name: string; teams: string[] };

export async function getLocalLeagueDraft(): Promise<LocalLeagueDraft | null> {
  const serialized = await AsyncStorage.getItem(key);
  if (!serialized) return null;
  try {
    const value: unknown = JSON.parse(serialized);
    if (!value || typeof value !== "object") return null;
    const draft = value as Partial<LocalLeagueDraft>;
    return typeof draft.name === "string" &&
      Array.isArray(draft.teams) &&
      draft.teams.every((team) => typeof team === "string")
      ? { name: draft.name, teams: draft.teams }
      : null;
  } catch {
    return null;
  }
}
export function saveLocalLeagueDraft(draft: LocalLeagueDraft) {
  return AsyncStorage.setItem(key, JSON.stringify(draft));
}
export function clearLocalLeagueDraft() {
  return AsyncStorage.removeItem(key);
}

/** Convierte exclusivamente un borrador completo al contrato de alta/publicación. */
export function toLeagueInput(draft: LocalLeagueDraft | null): LeagueInput | undefined {
  if (!draft) return undefined;
  const name = draft.name.trim();
  const teams = draft.teams.map((team) => team.trim()).filter(Boolean);
  if (
    !name ||
    name.length > 140 ||
    teams.length < 2 ||
    teams.length > maximumLeagueTeams ||
    new Set(teams.map((team) => team.toLowerCase())).size !== teams.length ||
    teams.some((team) => team.length > 100)
  ) {
    return undefined;
  }
  return { name, teams: teams.map((name) => ({ name })) };
}
