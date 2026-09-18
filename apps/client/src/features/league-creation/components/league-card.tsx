import { router, type Href } from "expo-router";
import { Pressable, StyleSheet, View } from "react-native";

import { space } from "@tournaments-manager/design-tokens";

import type { AccountTournament } from "@/api/generated/models";
import { getTranslator } from "@/shared/i18n/locale";
import { getTournamentStateLabel } from "@/shared/i18n/league-state";
import { Card, DisclosureIndicator, Text } from "@/shared/ui";

import { useTournamentState } from "../league-store";
import { TournamentCreatorChip } from "./league-creator-chip";

export function TournamentCard({ league }: { league: AccountTournament }) {
  const t = getTranslator();
  const state = useTournamentState(league.id, league.state);

  return (
    <Card>
      <Pressable
        accessibilityLabel={t("tournaments_open_league").replace("{name}", league.name)}
        accessibilityRole="button"
        onPress={() => router.push(`/tournament/${league.id}` as Href)}
        style={styles.row}
      >
        <View style={styles.copy}>
          <Text numberOfLines={2}>{league.name}</Text>
          <View style={styles.leagueState}>
            {league.relationship === "organizer" ? <TournamentCreatorChip /> : null}
            <Text color="secondary">{getTournamentStateLabel(t, state)}</Text>
          </View>
        </View>
        <DisclosureIndicator />
      </Pressable>
    </Card>
  );
}

const styles = StyleSheet.create({
  copy: { flex: 1, flexShrink: 1, gap: space[2] },
  leagueState: { alignItems: "center", flexDirection: "row", flexWrap: "wrap", gap: space[2] },
  row: {
    alignItems: "center",
    flexDirection: "row",
    justifyContent: "space-between",
    minHeight: 44,
  },
});
