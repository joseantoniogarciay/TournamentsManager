import type { PublicTournament } from "@/api/generated/models";
import type { TranslationKey } from "@/shared/i18n/locale";

type Sport = PublicTournament["sport"];

export function getSportLabelKey(sport: Sport): TranslationKey {
  switch (sport) {
    case "basketball":
      return "tournament_sport_basketball";
    case "handball":
      return "tournament_sport_handball";
    default:
      return "tournament_sport_football";
  }
}

export function sportAllowsTiedLeagueResult(sport: Sport) {
  return sport !== "basketball";
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
    case "basketball":
      return "basketball_withdraw_team_description";
    case "handball":
      return "handball_withdraw_team_description";
    default:
      return "league_withdraw_team_description";
  }
}
