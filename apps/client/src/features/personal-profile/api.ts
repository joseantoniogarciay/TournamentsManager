import { authenticatedApiFetch, APIUnexpectedResponseError } from "@/api/fetch";
import { getPersonalProfile } from "@/api/generated/tournaments/tournaments";
import type { GetPersonalProfileParams, PersonalProfile } from "@/api/generated/models";
import { parsePersonalMatchPage } from "@/features/tournament-browsing/api";
import { SportStatisticsSport } from "@/api/generated/models/sportStatisticsSport";

const record = (v: unknown): v is Record<string, unknown> =>
  !!v && typeof v === "object" && !Array.isArray(v);
export function parsePersonalProfile(value: unknown): PersonalProfile | null {
  if (
    !record(value) ||
    !Array.isArray(value.statistics) ||
    !Array.isArray(value.items) ||
    !parsePersonalMatchPage({ ...value, hasMatches: value.items.length > 0 })
  )
    return null;
  const seen = new Set<string>();
  for (const stat of value.statistics) {
    if (
      !record(stat) ||
      !Object.values(SportStatisticsSport).includes(stat.sport as SportStatisticsSport) ||
      seen.has(String(stat.sport)) ||
      typeof stat.drawsAllowed !== "boolean"
    )
      return null;
    seen.add(String(stat.sport));
    for (const key of ["played", "wins", "draws", "losses"])
      if (!Number.isSafeInteger(stat[key]) || (stat[key] as number) < 0) return null;
    const { played, wins, draws, losses, winPercentage } =
      stat as unknown as PersonalProfile["statistics"][number];
    if (
      played !== wins + draws + losses ||
      (!stat.drawsAllowed && draws !== 0) ||
      (played === 0
        ? winPercentage !== null
        : typeof winPercentage !== "number" ||
          !Number.isFinite(winPercentage) ||
          Math.abs(winPercentage - Math.round((wins * 1000) / played) / 10) > 0.001)
    )
      return null;
  }
  for (const item of value.items)
    if (
      !record(item) ||
      !["home", "away"].includes(String(item.ownSide)) ||
      !record(item.match) ||
      !["completed", "bye"].includes(String(item.match.state))
    )
      return null;
  return value as unknown as PersonalProfile;
}
export async function loadPersonalProfile(params: GetPersonalProfileParams, signal: AbortSignal) {
  const response = await getPersonalProfile(params, { signal }, authenticatedApiFetch);
  const value = response.status === 200 ? parsePersonalProfile(response.data) : null;
  if (!value) throw new APIUnexpectedResponseError(response.status);
  return value;
}
