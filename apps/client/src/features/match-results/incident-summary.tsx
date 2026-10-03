import { StyleSheet, View } from "react-native";
import { space } from "@tournaments-manager/design-tokens";
import type { Match } from "@/api/generated/models";
import { getTranslator } from "@/shared/i18n/locale";
import { Text } from "@/shared/ui";

export function IncidentSummary({
  match,
  teams,
  showAdministrativeScore = false,
}: {
  match: Match;
  teams: Map<string, string>;
  showAdministrativeScore?: boolean;
}) {
  const i = match.incident;
  if (!i) return null;
  const t = getTranslator();
  const affected = teams.get(i.side === "home" ? match.homeTeamId : match.awayTeamId) ?? "";
  const winner = teams.get(i.side === "home" ? match.awayTeamId : match.homeTeamId) ?? "";
  const partial =
    i.partialSets?.map((set) => `${set.homeScore}–${set.awayScore}`).join(" · ") ??
    (i.partialHomeScore !== undefined && i.partialAwayScore !== undefined
      ? `${i.partialHomeScore}–${i.partialAwayScore}`
      : undefined);
  return (
    <View style={styles.stack}>
      <Text variant="bodyLarge">
        {t(i.type === "no_show" ? "result_incident_no_show" : "result_incident_retirement")}
      </Text>
      <Text color="secondary">
        {t(
          i.type === "no_show"
            ? "result_incident_no_show_description"
            : "result_incident_retirement_description",
        ).replace("{participant}", affected)}
      </Text>
      <Text>{t("result_incident_winner").replace("{participant}", winner)}</Text>
      {partial ? (
        <Text color="secondary">{t("result_incident_partial").replace("{score}", partial)}</Text>
      ) : null}
      {showAdministrativeScore ? (
        <Text color="secondary">
          {t("result_incident_administrative").replace(
            "{score}",
            `${match.homeScore}–${match.awayScore}${match.sets.length ? ` (${match.sets.map((set) => `${set.homeScore}–${set.awayScore}`).join(" · ")})` : ""}`,
          )}
        </Text>
      ) : null}
    </View>
  );
}
const styles = StyleSheet.create({ stack: { gap: space[2] } });
