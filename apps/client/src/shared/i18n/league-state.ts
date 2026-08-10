import type { AccountLeagueState, PublicLeagueState } from "@/api/generated/models";

import type { TranslationKey } from "./locale";

type LeagueState = AccountLeagueState | PublicLeagueState;
type Translator = (key: TranslationKey) => string;

/** Traduce los estados de liga del contrato antes de presentarlos a la persona. */
export function getLeagueStateLabel(t: Translator, state: LeagueState) {
  switch (state) {
    case "published":
      return t("league_state_published");
    case "in_progress":
      return t("league_state_in_progress");
    case "completed":
      return t("league_state_completed");
    case "cancelled":
      return t("league_state_cancelled");
  }
}
