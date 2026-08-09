import { router, type Href, useFocusEffect } from "expo-router";
import { useCallback, useEffect, useRef, useState } from "react";
import { Pressable, RefreshControl, ScrollView, StyleSheet, View } from "react-native";
import { StatusBar } from "expo-status-bar";

import { radius, space } from "@tournaments-manager/design-tokens";

import { APISessionInvalidatedError } from "@/api/fetch";
import { getTranslator } from "@/shared/i18n/locale";
import { listRecentRelatedLeagues } from "@/features/league-creation/api";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getLeagueStateLabel } from "@/shared/i18n/league-state";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { useSession } from "@/shared/session/session-provider";
import { consumeDeferredInitialDeepLink } from "@/shared/navigation/deep-link-gate";
import { Button, Card, Screen, Text, useTabContentBottomPadding } from "@/shared/ui";

const safeAreaProbeHeight = 500;

export default function HomeScreen() {
  const { colors, resolvedTheme } = usePreferences();
  const { revision, user } = useSession();
  const { show } = useFeedback();
  const tabContentBottomPadding = useTabContentBottomPadding();
  const t = getTranslator();
  const [recentLeagues, setRecentLeagues] = useState<
    Awaited<ReturnType<typeof listRecentRelatedLeagues>>
  >([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const loadedAccountID = useRef<string | null>(null);

  useEffect(() => {
    const frame = requestAnimationFrame(() => {
      const deferredDeepLink = consumeDeferredInitialDeepLink();
      if (deferredDeepLink) router.replace(deferredDeepLink as Href);
    });
    return () => cancelAnimationFrame(frame);
  }, []);

  const loadRecentLeagues = useCallback(
    async (isManualRefresh = false) => {
      if (!user) return;
      if (isManualRefresh) setIsRefreshing(true);
      else setIsLoading(true);
      try {
        setRecentLeagues(await listRecentRelatedLeagues());
      } catch (error) {
        if (error instanceof APISessionInvalidatedError) return;
        const failure = getRequestFailure(error);
        show({ kind: failure.kind, message: t(failure.messageKey) });
      } finally {
        if (isManualRefresh) setIsRefreshing(false);
        else setIsLoading(false);
      }
    },
    [show, t, user],
  );

  useFocusEffect(
    useCallback(() => {
      if (!user) {
        loadedAccountID.current = null;
        setRecentLeagues([]);
        setIsLoading(false);
        setIsRefreshing(false);
        return;
      }
      if (loadedAccountID.current === user.id) return;
      loadedAccountID.current = user.id;
      void loadRecentLeagues();
    }, [loadRecentLeagues, user]),
  );

  return (
    <Screen bottomInset="none">
      <StatusBar style={resolvedTheme === "dark" ? "light" : "dark"} />
      <ScrollView
        key={revision}
        contentContainerStyle={[styles.content, { paddingBottom: tabContentBottomPadding }]}
        refreshControl={
          user ? (
            <RefreshControl
              onRefresh={() => void loadRecentLeagues(true)}
              refreshing={isRefreshing}
              tintColor={colors.border.focus}
            />
          ) : undefined
        }
        showsVerticalScrollIndicator={false}
        style={styles.scroll}
      >
        <Card>
          <View style={styles.hero}>
            <Text variant="display">{t("home_hero")}</Text>
            <Text color="secondary" variant="bodyLarge">
              {t("home_introduction")}
            </Text>
            <Button
              label={t("home_create_tournament")}
              onPress={() => router.push("/create-tournament" as Href)}
            />
          </View>
        </Card>

        {user && isLoading ? (
          <Card>
            <Text>{t("common_loading")}</Text>
          </Card>
        ) : null}

        {user && !isLoading ? (
          <View style={styles.recentSection}>
            <Text style={styles.recentTitle} variant="title">
              {t("home_recent_leagues_title")}
            </Text>
            {recentLeagues.length === 0 ? (
              <View style={styles.recentEmpty}>
                <Text color="secondary">{t("home_recent_leagues_empty")}</Text>
              </View>
            ) : (
              recentLeagues.map((league) => (
                <Card key={league.id}>
                  <Pressable
                    accessibilityLabel={t("tournaments_open_league").replace("{name}", league.name)}
                    accessibilityRole="button"
                    onPress={() => router.push(`/league/${league.id}` as Href)}
                    style={styles.leagueRow}
                  >
                    <View style={styles.section}>
                      <Text>{league.name}</Text>
                      <Text color="secondary">{getLeagueStateLabel(t, league.state)}</Text>
                    </View>
                    <Text color="secondary">›</Text>
                  </Pressable>
                </Card>
              ))
            )}
          </View>
        ) : null}

        <Card>
          <View style={styles.section}>
            <Text variant="title">{t("home_section_title")}</Text>
            <Text color="secondary">{t("home_section_description")}</Text>
          </View>
        </Card>

        <Card>
          <View style={styles.steps}>
            <Step
              number="1"
              title={t("home_step_1_title")}
              description={t("home_step_1_description")}
            />
            <Step
              number="2"
              title={t("home_step_2_title")}
              description={t("home_step_2_description")}
            />
            <Step
              number="3"
              title={t("home_step_3_title")}
              description={t("home_step_3_description")}
            />
          </View>
        </Card>

        <Card style={{ height: safeAreaProbeHeight }} />
      </ScrollView>
    </Screen>
  );
}

function Step({
  number,
  title,
  description,
}: {
  number: string;
  title: string;
  description: string;
}) {
  const { colors } = usePreferences();

  return (
    <View style={styles.step}>
      <View style={[styles.stepNumber, { borderColor: colors.text.primary }]}>
        <Text>{number}</Text>
      </View>
      <View style={styles.stepContent}>
        <Text variant="bodyLarge">{title}</Text>
        <Text color="secondary">{description}</Text>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  scroll: { flex: 1 },
  content: { gap: space[5] },
  hero: { gap: space[4] },
  leagueRow: {
    alignItems: "center",
    flexDirection: "row",
    justifyContent: "space-between",
    minHeight: 44,
  },
  recentEmpty: { alignItems: "center", paddingHorizontal: space[5] },
  recentSection: { gap: space[5] },
  recentTitle: { marginHorizontal: space[5] },
  section: { gap: space[2] },
  steps: { gap: space[5] },
  step: { flexDirection: "row", gap: space[3] },
  stepNumber: {
    alignItems: "center",
    borderWidth: 1,
    borderRadius: radius.pill,
    height: 28,
    justifyContent: "center",
    width: 28,
  },
  stepContent: { flex: 1, gap: space[1] },
});
