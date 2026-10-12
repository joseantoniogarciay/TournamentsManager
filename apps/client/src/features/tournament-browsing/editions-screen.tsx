import { router, Stack, useLocalSearchParams } from "expo-router";
import { useCallback } from "react";
import { Pressable, RefreshControl, ScrollView, StyleSheet, View } from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";
import { control, space } from "@tournaments-manager/design-tokens";
import type { TournamentEdition, TournamentEditionPage } from "@/api/generated/models";
import { getTranslator } from "@/shared/i18n/locale";
import { getTournamentStateLabel } from "@/shared/i18n/league-state";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import {
  Button,
  Card,
  DisclosureIndicator,
  LoadingTransition,
  NavigationHeaderButton,
  RequestErrorCard,
  Screen,
  Text,
  usesLiquidGlassNavigation,
} from "@/shared/ui";
import { getEditions } from "./api";
import { useCollection } from "./use-collection";
const editionKey = (item: TournamentEdition) => item.id;
export function EditionsScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const t = getTranslator();
  const { colors } = usePreferences();
  const insets = useSafeAreaInsets();
  const fetchPage = useCallback(
    (cursor: string | undefined, signal: AbortSignal) => getEditions(id, cursor, signal),
    [id],
  );
  const { page, error, loading, refreshing, moreError, loadingMore, load } = useCollection<
    TournamentEdition,
    TournamentEditionPage
  >(id, fetchPage, editionKey);
  const close = () => {
    if (router.canDismiss()) router.dismiss();
    else router.replace(`/tournament/${id}`);
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
          title: t("tournament_editions"),
          ...(!usesLiquidGlassNavigation
            ? {
                headerLeft: () => (
                  <NavigationHeaderButton
                    accessibilityLabel={t("common_back")}
                    icon="back"
                    nativeIcon={{ android: "arrow_back", ios: "chevron.left", web: "arrow_back" }}
                    onPress={close}
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
            icon="chevron.left"
            onPress={close}
          />
        </Stack.Toolbar>
      ) : null}
      <Screen
        navigationTitle={t("tournament_editions")}
        topInset="navigation-bar"
        bottomInset="none"
      >
        <ScrollView
          contentContainerStyle={[styles.content, { paddingBottom: insets.bottom + space[5] }]}
          refreshControl={
            <RefreshControl
              refreshing={refreshing}
              onRefresh={() => void load(page ? "refresh" : "initial")}
              colors={[colors.indicator.default]}
              tintColor={colors.indicator.default}
            />
          }
        >
          {error && !page && !loading ? (
            <RequestErrorCard
              message={t(error)}
              actionLabel={t("common_retry")}
              onRetry={() => void load()}
            />
          ) : null}
          {page?.items.map((edition) => (
            <Card key={edition.id}>
              <Pressable
                accessibilityRole="button"
                accessibilityLabel={t("tournaments_open_league").replace("{name}", edition.name)}
                accessibilityState={{ selected: edition.id === id }}
                onPress={() => router.push(`/tournament/${edition.id}`)}
                style={styles.row}
              >
                <View style={styles.copy}>
                  {edition.id === id ? (
                    <Text color="secondary">{t("editions_current")}</Text>
                  ) : null}
                  <Text>{edition.name}</Text>
                  <Text color="secondary">{getTournamentStateLabel(t, edition.state)}</Text>
                  {edition.state === "completed" && edition.champions.length > 0 ? (
                    <Text color="secondary">
                      {t("editions_champions").replace("{names}", edition.champions.join(", "))}
                    </Text>
                  ) : null}
                </View>
                <DisclosureIndicator />
              </Pressable>
            </Card>
          ))}
          {moreError ? (
            <View style={styles.pagination}>
              <Text color="secondary">{t(moreError)}</Text>
            </View>
          ) : null}
          {page?.nextCursor ? (
            <View style={styles.pagination}>
              <Button
                variant="secondary"
                label={t(moreError ? "common_retry" : "tournaments_load_more")}
                loading={loadingMore}
                disabled={refreshing}
                onPress={() => void load("more")}
              />
            </View>
          ) : null}
        </ScrollView>
        {(!page && !error) || loading ? (
          <LoadingTransition active message={t("common_loading")} />
        ) : null}
      </Screen>
    </>
  );
}
const styles = StyleSheet.create({
  content: { flexGrow: 1, gap: space[5] },
  copy: { flex: 1, gap: space[2] },
  row: { flexDirection: "row", alignItems: "center", gap: space[3], minHeight: control.minHeight },
  pagination: { paddingHorizontal: space[5] },
});
