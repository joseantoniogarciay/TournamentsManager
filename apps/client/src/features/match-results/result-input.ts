import type { Match, MatchIncident, SetResult } from "@/api/generated/models";
import { isSetSport } from "@/shared/tournaments/sport";

export type EditableScore = {
  incidentType?: MatchIncident["type"];
  incidentSide?: MatchIncident["side"];
  home: string;
  away: string;
  homePenalties: string;
  awayPenalties: string;
  sets: Array<{ home: string; away: string }>;
};

export function parseSetScores(
  score: EditableScore,
  bestOfSets: number,
  sport: string,
  pointsPerGame?: 15 | 21,
  partial = false,
): SetResult[] | null {
  let lastPlayedIndex = -1;
  score.sets.forEach((set, index) => {
    if (set.home !== "" || set.away !== "") lastPlayedIndex = index;
  });
  if (lastPlayedIndex < 0) return partial ? [] : null;
  const played = score.sets.slice(0, lastPlayedIndex + 1);
  if (played.some((set) => !/^\d+$/.test(set.home) || !/^\d+$/.test(set.away))) return null;
  const sets = played.map((set) => ({ homeScore: Number(set.home), awayScore: Number(set.away) }));
  const needed = Math.floor(bestOfSets / 2) + 1;
  let homeWins = 0;
  let awayWins = 0;
  for (const [index, set] of sets.entries()) {
    if (homeWins === needed || awayWins === needed) return null;
    const winner = Math.max(set.homeScore, set.awayScore);
    const loser = Math.min(set.homeScore, set.awayScore);
    const pointTarget =
      sport === "badminton" ? pointsPerGame : sport === "volleyball" ? (index === 4 ? 15 : 25) : 11;
    if (pointTarget === undefined) return null;
    const pointCap = pointsPerGame === 15 ? 21 : 30;
    const valid =
      winner <= 32767 &&
      (sport === "badminton"
        ? winner <= pointCap &&
          ((winner === pointTarget && loser <= pointTarget - 2) ||
            (winner > pointTarget && winner - loser === 2) ||
            (winner === pointCap && loser === pointCap - 1))
        : sport === "volleyball" || sport === "table_tennis"
          ? (winner === pointTarget && loser <= pointTarget - 2) ||
            (winner > pointTarget && winner - loser === 2)
          : (winner === 6 && loser <= 4) || (winner === 7 && (loser === 5 || loser === 6)));
    if (!valid) {
      const unfinished =
        sport === "tennis" || sport === "padel"
          ? winner <= 5 || (winner === 6 && loser >= 5)
          : winner < (sport === "badminton" ? pointCap : 32767) &&
            (winner < pointTarget || winner - loser <= 1);
      if (!partial || index !== sets.length - 1 || !unfinished || winner > 32767) return null;
      continue;
    }
    if (set.homeScore > set.awayScore) homeWins += 1;
    else awayWins += 1;
  }
  return partial
    ? homeWins < needed && awayWins < needed
      ? sets
      : null
    : homeWins === needed || awayWins === needed
      ? sets
      : null;
}

export function getIncidentInput(
  score: EditableScore,
  sport: string,
  bestOfSets: number,
  pointsPerGame?: 15 | 21,
): MatchIncident | null {
  if (!score.incidentType || !score.incidentSide) return null;
  const incident: MatchIncident = { type: score.incidentType, side: score.incidentSide };
  if (score.incidentType === "no_show") return incident;
  if (isSetSport(sport)) {
    const sets = parseSetScores(score, bestOfSets, sport, pointsPerGame, true);
    return sets === null ? null : { ...incident, ...(sets.length ? { partialSets: sets } : {}) };
  }
  if (score.home === "" && score.away === "") return incident;
  if (
    !/^\d+$/.test(score.home) ||
    !/^\d+$/.test(score.away) ||
    Number(score.home) > 2147483647 ||
    Number(score.away) > 2147483647
  )
    return null;
  return {
    ...incident,
    partialHomeScore: Number(score.home),
    partialAwayScore: Number(score.away),
  };
}

export function editableScoreFromMatch(match: Match, bestOfSets: number): EditableScore {
  const incident = match.incident;
  const sets = incident ? (incident.partialSets ?? []) : match.sets;
  return {
    incidentType: incident?.type,
    incidentSide: incident?.side,
    home: (incident ? incident.partialHomeScore : match.homeScore)?.toString() ?? "",
    away: (incident ? incident.partialAwayScore : match.awayScore)?.toString() ?? "",
    homePenalties: match.homePenalties?.toString() ?? "",
    awayPenalties: match.awayPenalties?.toString() ?? "",
    sets: Array.from({ length: bestOfSets }, (_, index) => ({
      home: sets[index]?.homeScore.toString() ?? "",
      away: sets[index]?.awayScore.toString() ?? "",
    })),
  };
}
