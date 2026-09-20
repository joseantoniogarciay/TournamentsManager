import { router, Stack, useLocalSearchParams } from "expo-router";
import { useCallback, useEffect, useState } from "react";
import { ScrollView, StyleSheet } from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";

import { space } from "@tournaments-manager/design-tokens";

import {
  getTournamentRelationship,
  TournamentUnavailableError,
} from "@/features/league-creation/api";
import { useTournament, useTournamentStore } from "@/features/league-creation/league-store";
import { TournamentTeamManagement } from "@/features/league-creation/team-management";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { useSession } from "@/shared/session/session-provider";
import {
  LoadingTransition,
  NavigationHeaderButton,
  RequestErrorCard,
  Screen,
  usesLiquidGlassNavigation,
} from "@/shared/ui";

export default function TournamentTeamsScreen() {
  const t = getTranslator();
  const { id } = useLocalSearchParams<{ id: string }>();
  const { user } = useSession();
  const { colors } = usePreferences();
  const insets = useSafeAreaInsets();
  const league = useTournament(id);
  const { loadTournament, refreshTournament } = useTournamentStore();
  const [relationship, setRelationship] = useState<string | null>();
  const [loadErrorMessage, setLoadErrorMessage] = useState<string>();
  const [leagueUnavailable, setTournamentUnavailable] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

  const load = useCallback(
    async (force = false) => {
      if (!id) {
        setLoadErrorMessage(t("common_request_error"));
        return;
      }
      setIsLoading(true);
      setLoadErrorMessage(undefined);
      setTournamentUnavailable(false);
      setRelationship(undefined);
      try {
        await (force ? refreshTournament(id) : loadTournament(id));
        if (user) setRelationship(await getTournamentRelationship(id));
        else setRelationship(null);
      } catch (error) {
        const unavailable = error instanceof TournamentUnavailableError;
        setTournamentUnavailable(unavailable);
        setLoadErrorMessage(
          t(unavailable ? "league_unavailable" : getRequestFailure(error).messageKey),
        );
      } finally {
        setIsLoading(false);
      }
    },
    [id, loadTournament, refreshTournament, t, user],
  );
  useEffect(() => {
    void load();
  }, [load]);

  const close = () => {
    if (router.canDismiss()) {
      router.dismiss();
      return;
    }
    router.replace(id ? `/tournament/${id}` : "/");
  };
  const closeUnavailable = () => {
    if (router.canDismiss()) {
      router.dismiss();
      return;
    }
    router.replace("/");
  };

  return (
    <>
      <Stack.Screen
        options={{
          headerBackVisible: false,
          headerShadowVisible: false,
          headerStyle: { backgroundColor: colors.surface.canvas },
          headerTintColor: colors.text.primary,
          headerTitleAlign: "center",
          title: t("league_teams"),
          ...(!usesLiquidGlassNavigation
            ? {
                headerLeft: () => (
                  <NavigationHeaderButton
                    accessibilityLabel={t("common_back")}
                    icon="close"
                    nativeIcon={{ android: "close", ios: "xmark", web: "close" }}
                    onPress={close}
                    side="left"
                  />
                ),
              }
            : {}),
        }}
      />
      {usesLiquidGlassNavigation ? (
        <Stack.Toolbar placement="left">
          <Stack.Toolbar.Button
            accessibilityLabel={t("common_back")}
            icon="xmark"
            onPress={close}
          />
        </Stack.Toolbar>
      ) : null}
      <Screen bottomInset="none" topInset="navigation-bar">
        {loadErrorMessage ? (
          <RequestErrorCard
            actionLabel={t(leagueUnavailable ? "common_close" : "common_retry")}
            loading={leagueUnavailable ? false : isLoading}
            message={loadErrorMessage}
            onRetry={leagueUnavailable ? closeUnavailable : () => void load(true)}
          />
        ) : !league || (user && relationship === undefined) ? (
          <LoadingTransition active message={t("common_loading")} />
        ) : (
          <ScrollView
            contentContainerStyle={[styles.content, { paddingBottom: insets.bottom + space[5] }]}
            showsVerticalScrollIndicator={false}
          >
            <TournamentTeamManagement relationship={relationship ?? null} tournament={league} />
          </ScrollView>
        )}
      </Screen>
    </>
  );
}

const styles = StyleSheet.create({
  content: { gap: space[5], paddingBottom: space[5] },
});
