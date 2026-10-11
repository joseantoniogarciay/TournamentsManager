import type { PublicTournament, RepeatTournamentInput } from "@/api/generated/models";

export function canRepeatTournament(relationship: string | null | undefined, state: string) {
  return relationship === "organizer" && state === "completed";
}

export function repetitionConfiguration(source: PublicTournament) {
  const stage = source.stages.find((item) => item.type === "league");
  return {
    format: source.format,
    ...(source.format !== "single_elimination" ? { roundRobinLegs: source.roundRobinLegs } : {}),
    ...(source.format === "league_then_single_elimination" && stage
      ? {
          leagueStructure: stage.leagueStructure,
          qualifierCount: stage.qualifierCount,
          groupCount: stage.groupCount,
          qualifiersPerGroup: stage.qualifiersPerGroup,
        }
      : {}),
  };
}

export function reviewedRepetitionInput(
  requestId: string,
  name: string,
  teams: { key: string; name: string }[],
  ownKey: string,
): RepeatTournamentInput | null {
  const own = teams.find((team) => team.key === ownKey);
  const names = teams.map((team) => team.name.trim());
  if (
    !own ||
    !name.trim() ||
    Array.from(name.trim()).length > 56 ||
    teams.length < 1 ||
    teams.length > 64 ||
    names.some((value) => !value || Array.from(value).length > 100) ||
    new Set(names.map((value) => value.toLowerCase())).size !== names.length
  )
    return null;
  return {
    requestId,
    name: name.trim(),
    teams: [own, ...teams.filter((team) => team.key !== ownKey)].map((team) => ({
      name: team.name.trim(),
    })),
  };
}
