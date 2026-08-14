import { router, type Href } from "expo-router";
import { Pressable, StyleSheet, View } from "react-native";

import { space } from "@tournaments-manager/design-tokens";

import type { AccountLeague } from "@/api/generated/models";
import { getTranslator } from "@/shared/i18n/locale";
import { getLeagueStateLabel } from "@/shared/i18n/league-state";
import { Card, DisclosureIndicator, Text } from "@/shared/ui";

import { useLeagueState } from "../league-store";
import { LeagueCreatorChip } from "./league-creator-chip";

export function LeagueCard({ league }: { league: AccountLeague }) {
  const t = getTranslator();
  const state = useLeagueState(league.id, league.state);

  return (
    <Card>
      <Pressable
        accessibilityLabel={t("tournaments_open_league").replace("{name}", league.name)}
        accessibilityRole="button"
        onPress={() => router.push(`/league/${league.id}` as Href)}
        style={styles.row}
      >
        <View style={styles.copy}>
          <Text numberOfLines={2}>{league.name}</Text>
          <View style={styles.leagueState}>
            {league.relationship === "organizer" ? <LeagueCreatorChip /> : null}
            <Text color="secondary">{getLeagueStateLabel(t, state)}</Text>
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
