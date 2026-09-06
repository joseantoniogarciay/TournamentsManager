import { router, type Href, useFocusEffect } from "expo-router";
import Head from "expo-router/head";
import { useCallback, useEffect, useRef, useState } from "react";
import { RefreshControl, ScrollView, StyleSheet, View } from "react-native";
import { StatusBar } from "expo-status-bar";

import { radius, space } from "@tournaments-manager/design-tokens";

import { APISessionInvalidatedError } from "@/api/fetch";
import { getTranslator } from "@/shared/i18n/locale";
import { isStaticWebRender } from "@/shared/i18n/is-static-web-render";
import { listRecentRelatedLeagues } from "@/features/league-creation/api";
import { LeagueCard } from "@/features/league-creation/components/league-card";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { ProductAnalyticsPreferenceCard } from "@/shared/preferences/product-analytics-preference-card";
import { useSession } from "@/shared/session/session-provider";
import { consumeDeferredInitialDeepLink } from "@/shared/navigation/deep-link-gate";
import { Button, Card, Screen, Text, useTabContentBottomPadding } from "@/shared/ui";

export default function HomeScreen() {
  const { colors, resolvedTheme } = usePreferences();
  const { isRestoring, revision, user } = useSession();
  const { show } = useFeedback();
  const tabContentBottomPadding = useTabContentBottomPadding();
  const t = getTranslator();
  const [recentLeagues, setRecentLeagues] = useState<
    Awaited<ReturnType<typeof listRecentRelatedLeagues>>
  >([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const loadedAccountID = useRef<string | null>(null);
  const showGuestHome = !user && (!isRestoring || isStaticWebRender());

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
    <>
      <HomeMetadata />
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
                colors={[colors.indicator.default]}
                tintColor={colors.indicator.default}
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

          {user ? (
            <View style={styles.recentSection}>
              <Text style={styles.recentTitle} variant="title">
                {t("home_recent_leagues_title")}
              </Text>
              {isLoading ? (
                <Text color="secondary" style={styles.recentEmpty}>
                  {t("common_loading")}
                </Text>
              ) : recentLeagues.length === 0 ? (
                <View style={styles.recentEmpty}>
                  <Text color="secondary">{t("home_recent_leagues_empty")}</Text>
                </View>
              ) : (
                recentLeagues.map((league) => <LeagueCard key={league.id} league={league} />)
              )}
            </View>
          ) : null}

          {showGuestHome ? <GuestOnboarding t={t} /> : null}

          <ProductAnalyticsCard />
        </ScrollView>
      </Screen>
    </>
  );
}

function HomeMetadata() {
  const t = getTranslator();
  const title = t("home_web_title");
  const description = t("home_web_description");
  const imageAlt = t("home_web_image_alt");
  const publicBaseURL = (process.env.EXPO_PUBLIC_APP_LINK_URL ?? "https://fasttourney.com").replace(
    /\/$/,
    "",
  );
  const homeURL = `${publicBaseURL}/`;
  const previewImageURL = `${publicBaseURL}/fasttourney-league-preview.png`;

  return (
    <Head>
      <title>{title}</title>
      <meta name="description" content={description} />
      <link rel="canonical" href={homeURL} />
      <meta property="og:type" content="website" />
      <meta property="og:title" content={title} />
      <meta property="og:description" content={description} />
      <meta property="og:url" content={homeURL} />
      <meta property="og:image" content={previewImageURL} />
      <meta property="og:image:width" content="1200" />
      <meta property="og:image:height" content="630" />
      <meta property="og:image:alt" content={imageAlt} />
      <meta name="twitter:card" content="summary_large_image" />
      <meta name="twitter:title" content={title} />
      <meta name="twitter:description" content={description} />
      <meta name="twitter:image" content={previewImageURL} />
    </Head>
  );
}

function ProductAnalyticsCard() {
  const t = getTranslator();
  const { productAnalyticsEnabled } = usePreferences();
  const { show } = useFeedback();

  if (productAnalyticsEnabled) return null;

  return (
    <ProductAnalyticsPreferenceCard
      onEnabled={() => show({ kind: "success", message: t("product_analytics_enabled_feedback") })}
    />
  );
}

function GuestOnboarding({ t }: { t: ReturnType<typeof getTranslator> }) {
  return (
    <>
      <Card>
        <View style={styles.section}>
          <Text variant="title">{t("home_section_title")}</Text>
          <Text color="secondary">{t("home_section_description")}</Text>
        </View>
      </Card>
      <Card>
        <View style={styles.steps}>
          <Step
            description={t("home_step_1_description")}
            number="1"
            title={t("home_step_1_title")}
          />
          <Step
            description={t("home_step_2_description")}
            number="2"
            title={t("home_step_2_title")}
          />
          <Step
            description={t("home_step_3_description")}
            number="3"
            title={t("home_step_3_title")}
          />
        </View>
      </Card>
    </>
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
  recentEmpty: { alignItems: "center", paddingHorizontal: space[5], textAlign: "center" },
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
