import AsyncStorage from "@react-native-async-storage/async-storage";
import { randomUUID } from "expo-crypto";

import type { TournamentDraftInput } from "@/api/generated/models";
import { TournamentInputSport } from "@/api/generated/models/tournamentInputSport";

const key = "tm-league-draft";
export const maximumTournamentTeams = 64;
export const minimumTournamentTeamsToStart = 2;
export const maximumTournamentNameLength = 56;
export const maximumTeamNameLength = 100;
export type TournamentSport = TournamentInputSport;
export type LocalTournamentDraft = {
  draftId: string;
  name: string;
  sport: TournamentSport;
  teams: string[];
};

const uuidPattern = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;

export async function getLocalTournamentDraft(): Promise<LocalTournamentDraft | null> {
  const serialized = await AsyncStorage.getItem(key);
  if (!serialized) return null;
  try {
    const value: unknown = JSON.parse(serialized);
    if (!value || typeof value !== "object") return null;
    const draft = value as Partial<LocalTournamentDraft>;
    if (
      typeof draft.name !== "string" ||
      !Array.isArray(draft.teams) ||
      !draft.teams.every((team) => typeof team === "string")
    ) {
      return null;
    }
    const normalized = {
      draftId:
        typeof draft.draftId === "string" && uuidPattern.test(draft.draftId)
          ? draft.draftId
          : randomUUID(),
      name: draft.name,
      sport:
        draft.sport === TournamentInputSport.basketball ||
        draft.sport === TournamentInputSport.handball
          ? draft.sport
          : TournamentInputSport.football,
      teams: draft.teams,
    };
    if (draft.draftId !== normalized.draftId) {
      await AsyncStorage.setItem(key, JSON.stringify(normalized));
    }
    return normalized;
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
export function toTournamentDraftInput(
  draft: LocalTournamentDraft | null,
): TournamentDraftInput | undefined {
  if (!draft) return undefined;
  const name = draft.name.trim();
  const teams = draft.teams.map((team) => team.trim()).filter(Boolean);
  if (
    !name ||
    name.length > maximumTournamentNameLength ||
    teams.length < 1 ||
    teams.length > maximumTournamentTeams ||
    new Set(teams.map((team) => team.toLowerCase())).size !== teams.length ||
    teams.some((team) => team.length > maximumTeamNameLength)
  ) {
    return undefined;
  }
  return {
    draftId: draft.draftId,
    name,
    sport: draft.sport,
    teams: teams.map((name) => ({ name })),
  };
}
