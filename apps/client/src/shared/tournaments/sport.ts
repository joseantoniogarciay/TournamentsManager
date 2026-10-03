import type { PublicTournament } from "@/api/generated/models";
import type { TranslationKey } from "@/shared/i18n/locale";

type Sport = PublicTournament["sport"];

export function getSportLabelKey(sport: Sport): TranslationKey {
  switch (sport) {
    case "badminton":
      return "tournament_sport_badminton";
    case "basketball":
      return "tournament_sport_basketball";
    case "handball":
      return "tournament_sport_handball";
    case "tennis":
      return "tournament_sport_tennis";
    case "volleyball":
      return "tournament_sport_volleyball";
    case "table_tennis":
      return "tournament_sport_table_tennis";
    case "padel":
      return "tournament_sport_padel";
    default:
      return "tournament_sport_football";
  }
}

export function isRacketSport(sport: string) {
  return (
    sport === "tennis" || sport === "padel" || sport === "table_tennis" || sport === "badminton"
  );
}

export function isSetSport(sport: string) {
  return isRacketSport(sport) || sport === "volleyball";
}

export function sportAllowsTiedLeagueResult(sport: Sport) {
  return sport !== "basketball" && !isSetSport(sport);
}

export function sportUsesShootout(sport: Sport) {
  return sport === "football" || sport === "handball";
}

export function getShootoutKeys(sport: Sport): {
  help: TranslationKey;
  home: TranslationKey;
  away: TranslationKey;
} {
  return sport === "handball"
    ? {
        help: "handball_shootout_help",
        home: "handball_home_shootout",
        away: "handball_away_shootout",
      }
    : {
        help: "bracket_penalties_help",
        home: "bracket_home_penalties",
        away: "bracket_away_penalties",
      };
}

export function getWithdrawalDescriptionKey(sport: Sport): TranslationKey {
  switch (sport) {
    case "volleyball":
      return "volleyball_withdraw_team_description";
    case "basketball":
      return "basketball_withdraw_team_description";
    case "handball":
      return "handball_withdraw_team_description";
    default:
      return "league_withdraw_team_description";
  }
}

export function isValidBestOfSets(sport: string, value: unknown) {
  if (sport === "tennis") return value === 3 || value === 5;
  if (sport === "volleyball") return value === 5;
  if (sport === "badminton") return value === 3;
  if (sport === "padel") return value === 3;
  if (sport === "table_tennis") return value === 3 || value === 5 || value === 7;
  return value === undefined;
}

export function isValidPointsPerGame(sport: string, value: unknown) {
  return sport === "badminton" ? value === 15 || value === 21 : value === undefined;
}

export function getBadmintonScoreHelpKey(pointsPerGame: 15 | 21 | undefined): TranslationKey {
  return pointsPerGame === 15 ? "badminton_score_help_15" : "badminton_score_help_21";
}
