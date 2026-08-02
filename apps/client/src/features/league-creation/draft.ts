import AsyncStorage from "@react-native-async-storage/async-storage";

const key = "tm-league-draft";
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
