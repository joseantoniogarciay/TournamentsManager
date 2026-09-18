import type { AccountTournamentState, PublicTournamentState } from "@/api/generated/models";

import type { TranslationKey } from "./locale";

type TournamentState = AccountTournamentState | PublicTournamentState;
type Translator = (key: TranslationKey) => string;

/** Traduce los estados de liga del contrato antes de presentarlos a la persona. */
export function getTournamentStateLabel(t: Translator, state: TournamentState) {
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
