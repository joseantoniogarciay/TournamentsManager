import { apiFetch, authenticatedApiFetch, APIUnexpectedResponseError } from "@/api/fetch";
import {
  listPersonalMatches,
  listTournamentEditions,
} from "@/api/generated/tournaments/tournaments";
import type { PersonalMatchPage, TournamentEditionPage } from "@/api/generated/models";
import { parseMatch } from "@/features/league-creation/response-parser";

const uuid = (value: unknown): value is string =>
  typeof value === "string" &&
  /^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(value);
const record = (value: unknown): value is Record<string, unknown> =>
  !!value && typeof value === "object" && !Array.isArray(value);
const sports = [
  "football",
  "basketball",
  "handball",
  "tennis",
  "padel",
  "table_tennis",
  "volleyball",
  "badminton",
];
export function parsePersonalMatchPage(data: unknown): PersonalMatchPage | null {
  if (
    !record(data) ||
    !Array.isArray(data.items) ||
    typeof data.hasMatches !== "boolean" ||
    (data.nextCursor !== undefined && !uuid(data.nextCursor))
  )
    return null;
  for (const item of data.items) {
    if (
      !record(item) ||
      !uuid(item.tournamentId) ||
      typeof item.tournamentName !== "string" ||
      !sports.includes(String(item.sport)) ||
      !["league", "single_elimination", "qualification_tiebreak"].includes(
        String(item.stageType),
      ) ||
      typeof item.homeTeamName !== "string" ||
      typeof item.awayTeamName !== "string" ||
      !parseMatch(item.match)
    )
      return null;
  }
  return data as unknown as PersonalMatchPage;
}
export function parseEditionPage(data: unknown): TournamentEditionPage | null {
  if (
    !record(data) ||
    !Array.isArray(data.items) ||
    (data.nextCursor !== undefined && !uuid(data.nextCursor))
  )
    return null;
  for (const item of data.items) {
    if (
      !record(item) ||
      !uuid(item.id) ||
      typeof item.name !== "string" ||
      !["published", "in_progress", "completed", "cancelled"].includes(String(item.state)) ||
      typeof item.createdAt !== "string" ||
      !Number.isFinite(Date.parse(item.createdAt)) ||
      !Array.isArray(item.champions) ||
      item.champions.some((v) => typeof v !== "string")
    )
      return null;
  }
  return data as unknown as TournamentEditionPage;
}
export async function getPersonalMatches(
  view: "pending" | "results",
  cursor: string | undefined,
  signal: AbortSignal,
) {
  const response = await listPersonalMatches({ view, cursor }, { signal }, authenticatedApiFetch);
  const page = response.status === 200 ? parsePersonalMatchPage(response.data) : null;
  if (!page) throw new APIUnexpectedResponseError(response.status);
  return page;
}
export async function getEditions(id: string, cursor: string | undefined, signal: AbortSignal) {
  const response = await listTournamentEditions(id, { cursor }, { signal }, apiFetch);
  const page = response.status === 200 ? parseEditionPage(response.data) : null;
  if (!page) throw new APIUnexpectedResponseError(response.status);
  return page;
}
