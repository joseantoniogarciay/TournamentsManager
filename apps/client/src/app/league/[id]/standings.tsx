import { router, Stack, useLocalSearchParams } from "expo-router";
import { SymbolView } from "expo-symbols";
import { useEffect, useMemo, useState } from "react";
import {
  ActivityIndicator,
  Modal,
  Platform,
  Pressable,
  ScrollView,
  StyleSheet,
  View,
} from "react-native";

import { control, radius, space } from "@tournaments-manager/design-tokens";

import type { PublicLeague } from "@/api/generated/models";
import { getLeague } from "@/features/league-creation/api";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { Card, Screen, Text } from "@/shared/ui";
import { WebIcon } from "@/shared/ui/web-icon";

export default function LeagueStandingsScreen() {
  const t = getTranslator();
  const { id } = useLocalSearchParams<{ id: string }>();
  const { colors } = usePreferences();
  const { show } = useFeedback();
  const [league, setLeague] = useState<PublicLeague | null>(null);
  const [informationVisible, setInformationVisible] = useState(false);

  useEffect(() => {
    if (!id) return;
    void getLeague(id)
      .then(setLeague)
      .catch((error) => {
        const failure = getRequestFailure(error);
        show({ kind: failure.kind, message: t(failure.messageKey) });
      });
  }, [id, show, t]);

  const teams = useMemo(() => new Map(league?.teams.map((team) => [team.id, team.name])), [league]);
  const close = () => {
    if (router.canDismiss()) {
      router.dismiss();
      return;
    }
    router.replace(id ? `/league/${id}` : "/");
  };
  const navigationButton = (
    <Pressable
      accessibilityLabel={t("common_back")}
      accessibilityRole="button"
      onPress={close}
      style={[
        styles.navigationButton,
        { backgroundColor: colors.surface.default, borderColor: colors.border.default },
      ]}
    >
      {Platform.OS === "web" ? (
        <WebIcon color={colors.text.primary} name="close" size={control.iconSize} />
      ) : (
        <SymbolView
          name={{ android: "close", ios: "xmark", web: "close" }}
          size={control.iconSize}
          tintColor={colors.text.primary}
        />
      )}
    </Pressable>
  );
  const informationButton = (
    <Pressable
      accessibilityLabel={t("league_standings_information")}
      accessibilityRole="button"
      onPress={() => setInformationVisible(true)}
      style={[
        styles.navigationButton,
        { backgroundColor: colors.surface.default, borderColor: colors.border.default },
      ]}
    >
      {Platform.OS === "web" ? (
        <WebIcon color={colors.text.primary} name="info" size={control.iconSize} />
      ) : (
        <SymbolView
          name={{ android: "info", ios: "info.circle", web: "info" }}
          size={control.iconSize}
          tintColor={colors.text.primary}
        />
      )}
    </Pressable>
  );

  return (
    <>
      <Stack.Screen
        options={{
          headerBackVisible: false,
          headerShadowVisible: false,
          headerStyle: { backgroundColor: colors.surface.canvas },
          headerTintColor: colors.text.primary,
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
              icon="info.circle"
              onPress={() => setInformationVisible(true)}
            />
          </Stack.Toolbar>
        </>
      ) : null}
      <Screen topInset="navigation-bar">
        {!league ? (
          <View style={styles.loader}>
            <ActivityIndicator color={colors.text.primary} />
          </View>
        ) : (
          <ScrollView contentContainerStyle={styles.content} showsVerticalScrollIndicator={false}>
            <Card>
              <View style={styles.stack}>
                <Text variant="title">{t("league_standings_rules_title")}</Text>
                <Text color="secondary">{t("league_standings_rule_points")}</Text>
                <Text color="secondary">
                  {league.roundRobinLegs === 2
                    ? t("league_standings_rule_two_legs")
                    : t("league_standings_rule_one_leg")}
                </Text>
                <Text color="secondary">{t("league_standings_rule_general")}</Text>
                <Text color="secondary">{t("league_standings_rule_shared")}</Text>
              </View>
            </Card>
            {league.standings.length === 0 ? (
              <Card>
                <Text color="secondary">{t("league_standings_unavailable")}</Text>
              </Card>
            ) : (
              <Card style={styles.tableCard}>
                <ScrollView horizontal showsHorizontalScrollIndicator={false}>
                  <View>
                    <View
                      style={[styles.row, styles.headerRow, { borderColor: colors.border.default }]}
                    >
                      <Text color="secondary" style={styles.position}>
                        {t("league_standings_position")}
                      </Text>
                      <Text color="secondary" style={styles.team}>
                        {t("league_standings_team")}
                      </Text>
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
                        {t("league_standings_goal_difference")}
                      </Text>
                      <Text color="secondary" style={styles.points}>
                        {t("league_standings_points")}
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
                        <Text style={styles.stat}>{standing.played}</Text>
                        <Text style={styles.stat}>{standing.won}</Text>
                        <Text style={styles.stat}>{standing.drawn}</Text>
                        <Text style={styles.stat}>{standing.lost}</Text>
                        <Text style={styles.stat}>
                          {standing.goalDifference > 0
                            ? t("league_standings_positive_goal_difference").replace(
                                "{value}",
                                standing.goalDifference.toString(),
                              )
                            : standing.goalDifference}
                        </Text>
                        <Text style={styles.points} variant="title">
                          {standing.points}
                        </Text>
                      </View>
                    ))}
                  </View>
                </ScrollView>
              </Card>
            )}
          </ScrollView>
        )}
      </Screen>
      <Modal
        transparent
        visible={informationVisible}
        onRequestClose={() => setInformationVisible(false)}
      >
        <View style={styles.modalBackdrop}>
          <Pressable
            accessibilityLabel={t("common_close")}
            accessibilityRole="button"
            onPress={() => setInformationVisible(false)}
            style={StyleSheet.absoluteFill}
          />
          <Card>
            <View style={styles.stack} accessibilityViewIsModal>
              <Text variant="title">{t("league_standings_information")}</Text>
              <Text color="secondary">{t("league_standings_rule_points")}</Text>
              <Text color="secondary">
                {league?.roundRobinLegs === 2
                  ? t("league_standings_rule_two_legs")
                  : t("league_standings_rule_one_leg")}
              </Text>
              <Text color="secondary">{t("league_standings_rule_general")}</Text>
              <Text color="secondary">{t("league_standings_rule_shared")}</Text>
              <Pressable
                accessibilityLabel={t("common_close")}
                accessibilityRole="button"
                onPress={() => setInformationVisible(false)}
                style={styles.closeButton}
              >
                <WebIcon color={colors.text.primary} name="close" size={control.iconSize} />
              </Pressable>
            </View>
          </Card>
        </View>
      </Modal>
    </>
  );
}

const styles = StyleSheet.create({
  closeButton: {
    alignSelf: "flex-end",
    height: control.minHeight,
    justifyContent: "center",
    width: control.minHeight,
  },
  content: { gap: space[5], paddingBottom: space[5] },
  headerRow: { borderBottomWidth: 1 },
  loader: { alignItems: "center", flex: 1, justifyContent: "center" },
  modalBackdrop: {
    ...StyleSheet.absoluteFill,
    alignItems: "center",
    justifyContent: "center",
    padding: space[5],
  },
  navigationButton: {
    alignItems: "center",
    borderRadius: radius.pill,
    borderWidth: 1,
    height: control.minHeight,
    justifyContent: "center",
    width: control.minHeight,
  },
  points: { textAlign: "right", width: 44 },
  position: { textAlign: "center", width: 34 },
  row: {
    alignItems: "center",
    borderBottomWidth: 1,
    flexDirection: "row",
    minHeight: control.minHeight,
  },
  stack: { gap: space[3] },
  stat: { textAlign: "center", width: 36 },
  tableCard: { paddingHorizontal: space[3] },
  team: { flex: 1, minWidth: 120, paddingHorizontal: space[2] },
});
