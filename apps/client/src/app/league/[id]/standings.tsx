import { router, Stack, useLocalSearchParams } from "expo-router";
import { useCallback, useEffect, useMemo, useState } from "react";
import { Platform, ScrollView, StyleSheet, View } from "react-native";

import { control, radius, space } from "@tournaments-manager/design-tokens";

import type { PublicLeague } from "@/api/generated/models";
import { LeagueUnavailableError } from "@/features/league-creation/api";
import { useLeague, useLeagueStore } from "@/features/league-creation/league-store";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import {
  Button,
  Card,
  LoadingTransition,
  ModalDialog,
  NavigationHeaderButton,
  RequestErrorCard,
  Screen,
  Text,
} from "@/shared/ui";

export default function LeagueStandingsScreen() {
  const t = getTranslator();
  const { id } = useLocalSearchParams<{ id: string }>();
  const { colors } = usePreferences();
  const league = useLeague(id);
  const { loadLeague, refreshLeague } = useLeagueStore();
  const [loadErrorMessage, setLoadErrorMessage] = useState<string>();
  const [leagueUnavailable, setLeagueUnavailable] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [informationVisible, setInformationVisible] = useState(false);
  const [statisticsContentWidth, setStatisticsContentWidth] = useState(0);
  const [statisticsViewportWidth, setStatisticsViewportWidth] = useState(0);
  const statisticsOverflow = statisticsContentWidth > statisticsViewportWidth + 1;

  const load = useCallback(
    async (force = false) => {
      if (!id) {
        setLoadErrorMessage(t("common_request_error"));
        return;
      }
      setIsLoading(true);
      setLoadErrorMessage(undefined);
      setLeagueUnavailable(false);
      try {
        await (force ? refreshLeague(id) : loadLeague(id));
      } catch (error) {
        const unavailable = error instanceof LeagueUnavailableError;
        setLeagueUnavailable(unavailable);
        setLoadErrorMessage(
          t(unavailable ? "league_unavailable" : getRequestFailure(error).messageKey),
        );
      } finally {
        setIsLoading(false);
      }
    },
    [id, loadLeague, refreshLeague, t],
  );
  useEffect(() => {
    void load();
  }, [load]);

  const teams = useMemo(() => new Map(league?.teams.map((team) => [team.id, team.name])), [league]);
  const close = () => {
    if (router.canDismiss()) {
      router.dismiss();
      return;
    }
    router.replace(id ? `/league/${id}` : "/");
  };
  const closeUnavailable = () => {
    if (router.canDismiss()) {
      router.dismiss();
      return;
    }
    router.replace("/");
  };
  const navigationButton = (
    <NavigationHeaderButton
      accessibilityLabel={t("common_back")}
      icon="close"
      nativeIcon={{ android: "close", ios: "xmark", web: "close" }}
      onPress={close}
    />
  );
  const informationButton = (
    <NavigationHeaderButton
      accessibilityLabel={t("league_standings_information")}
      icon="info"
      nativeIcon={{ android: "info", ios: "info.circle", web: "info" }}
      onPress={() => setInformationVisible(true)}
      side="right"
    />
  );

  return (
    <>
      <Stack.Screen
        options={{
          headerBackVisible: false,
          headerShadowVisible: false,
          headerStyle: { backgroundColor: colors.surface.canvas },
          headerTintColor: colors.text.primary,
          headerTitleAlign: "center",
          title: t("league_standings"),
          ...(Platform.OS !== "ios"
            ? { headerLeft: () => navigationButton, headerRight: () => informationButton }
            : {}),
        }}
      />
      {Platform.OS === "ios" ? (
        <>
          <Stack.Toolbar placement="left">
            <Stack.Toolbar.Button
              accessibilityLabel={t("common_back")}
              icon="xmark"
              onPress={close}
            />
          </Stack.Toolbar>
          <Stack.Toolbar placement="right">
            <Stack.Toolbar.Button
              accessibilityLabel={t("league_standings_information")}
              icon="info"
              onPress={() => setInformationVisible(true)}
            />
          </Stack.Toolbar>
        </>
      ) : null}
      <Screen topInset="navigation-bar">
        {!league ? (
          loadErrorMessage ? (
            <RequestErrorCard
              actionLabel={t(leagueUnavailable ? "common_close" : "common_retry")}
              loading={leagueUnavailable ? false : isLoading}
              message={loadErrorMessage}
              onRetry={leagueUnavailable ? closeUnavailable : () => void load(true)}
            />
          ) : (
            <LoadingTransition active message={t("common_loading")} />
          )
        ) : (
          <ScrollView contentContainerStyle={styles.content} showsVerticalScrollIndicator={false}>
            {league.standings.length === 0 ? (
              <Card>
                <Text color="secondary">{t("league_standings_unavailable")}</Text>
              </Card>
            ) : (
              <View style={styles.table}>
                <View style={styles.tableColumns}>
                  <View style={styles.leftColumn}>
                    <View style={[styles.headerRow, { borderColor: colors.border.default }]}>
                      <Text color="secondary" style={styles.position}>
                        {t("league_standings_position")}
                      </Text>
                      <Text color="secondary" numberOfLines={1} style={styles.team}>
                        {t("league_standings_team")}
                      </Text>
                    </View>
                    {league.standings.map((standing) => (
                      <View
                        key={standing.teamId}
                        style={[styles.row, { borderColor: colors.border.default }]}
                      >
                        <Text style={styles.position}>{standing.position}</Text>
                        <Text numberOfLines={1} style={styles.team}>
                          {teams.get(standing.teamId) ?? ""}
                        </Text>
                      </View>
                    ))}
                  </View>
                  <View
                    onLayout={(event) => setStatisticsViewportWidth(event.nativeEvent.layout.width)}
                    style={styles.statisticsViewport}
                  >
                    <ScrollView
                      horizontal
                      onContentSizeChange={(width) => setStatisticsContentWidth(width)}
                      showsHorizontalScrollIndicator={statisticsOverflow}
                    >
                      <View>
                        <View style={[styles.headerRow, { borderColor: colors.border.default }]}>
                          <StatisticsHeader />
                        </View>
                        {league.standings.map((standing) => (
                          <View
                            key={standing.teamId}
                            style={[styles.row, { borderColor: colors.border.default }]}
                          >
                            <StatisticsValues standing={standing} />
                          </View>
                        ))}
                      </View>
                    </ScrollView>
                  </View>
                  <View style={styles.pointsColumn}>
                    <View style={[styles.headerRow, { borderColor: colors.border.default }]}>
                      <Text color="secondary" style={styles.points}>
                        {t("league_standings_points")}
                      </Text>
                    </View>
                    {league.standings.map((standing) => (
                      <View
                        key={standing.teamId}
                        style={[styles.row, { borderColor: colors.border.default }]}
                      >
                        <Text style={styles.points} variant="title">
                          {standing.points}
                        </Text>
                      </View>
                    ))}
                  </View>
                </View>
                {statisticsOverflow ? (
                  <Text color="secondary" style={styles.scrollHint}>
                    {t("league_standings_scroll_hint")}
                  </Text>
                ) : null}
              </View>
            )}
          </ScrollView>
        )}
      </Screen>
      <ModalDialog
        dismissAccessibilityLabel={t("common_close")}
        onDismiss={() => setInformationVisible(false)}
        visible={informationVisible}
      >
        <StandingsRulesContent league={league} />
        <Button label={t("common_ok")} onPress={() => setInformationVisible(false)} />
      </ModalDialog>
    </>
  );
}

function StatisticsHeader() {
  const t = getTranslator();

  return (
    <View style={styles.statisticsRow}>
      <Text color="secondary" style={styles.stat}>
        {t("league_standings_played")}
      </Text>
      <Text color="secondary" style={styles.stat}>
        {t("league_standings_won")}
      </Text>
      <Text color="secondary" style={styles.stat}>
        {t("league_standings_drawn")}
      </Text>
      <Text color="secondary" style={styles.stat}>
        {t("league_standings_lost")}
      </Text>
      <Text color="secondary" style={styles.stat}>
        {t("league_standings_goals_for")}
      </Text>
      <Text color="secondary" style={styles.stat}>
        {t("league_standings_goals_against")}
      </Text>
      <Text color="secondary" style={styles.stat}>
        {t("league_standings_goal_difference")}
      </Text>
    </View>
  );
}

function StatisticsValues({
  standing,
}: {
  standing: NonNullable<PublicLeague["standings"]>[number];
}) {
  const t = getTranslator();

  return (
    <View style={styles.statisticsRow}>
      <Text style={styles.stat}>{standing.played}</Text>
      <Text style={styles.stat}>{standing.won}</Text>
      <Text style={styles.stat}>{standing.drawn}</Text>
      <Text style={styles.stat}>{standing.lost}</Text>
      <Text style={styles.stat}>{standing.goalsFor}</Text>
      <Text style={styles.stat}>{standing.goalsAgainst}</Text>
      <Text style={styles.stat}>
        {standing.goalDifference > 0
          ? t("league_standings_positive_goal_difference").replace(
              "{value}",
              standing.goalDifference.toString(),
            )
          : standing.goalDifference}
      </Text>
    </View>
  );
}

function StandingsRulesContent({ league }: { league: PublicLeague | null | undefined }) {
  const t = getTranslator();

  return (
    <View style={styles.stack}>
      <Text variant="title">{t("league_standings_rules_title")}</Text>
      <Text color="secondary">{t("league_standings_rule_points")}</Text>
      <Text color="secondary">
        {league?.roundRobinLegs === 2
          ? t("league_standings_rule_two_legs")
          : t("league_standings_rule_one_leg")}
      </Text>
      <Text color="secondary">{t("league_standings_rule_general")}</Text>
      <Text color="secondary">{t("league_standings_rule_shared")}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  content: { gap: space[5], paddingBottom: space[5], paddingHorizontal: space[5] },
  headerRow: {
    alignItems: "center",
    borderBottomWidth: 1,
    flexDirection: "row",
    minHeight: control.minHeight,
  },
  leftColumn: { flexShrink: 0, width: 164 },
  navigationButton: {
    alignItems: "center",
    borderRadius: radius.pill,
    borderWidth: 1,
    height: control.minHeight,
    justifyContent: "center",
    width: control.minHeight,
  },
  points: { textAlign: "right", width: 44 },
  pointsColumn: { flexShrink: 0, width: 44 },
  position: { textAlign: "center", width: 34 },
  row: {
    alignItems: "center",
    borderBottomWidth: 1,
    flexDirection: "row",
    minHeight: control.minHeight,
  },
  stack: { gap: space[3] },
  scrollHint: { paddingTop: space[2], textAlign: "right" },
  stat: { textAlign: "center", width: 36 },
  statisticsRow: { flexDirection: "row" },
  statisticsViewport: { flex: 1, minWidth: 0 },
  table: { width: "100%" },
  tableColumns: { flexDirection: "row" },
  team: { flex: 1, minWidth: 0, paddingHorizontal: space[2] },
});
