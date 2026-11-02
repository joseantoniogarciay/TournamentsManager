import { Redirect, router, Stack } from "expo-router";
import { useCallback, useEffect, useState } from "react";
import {
  Platform,
  Pressable,
  RefreshControl,
  ScrollView as DialogScrollView,
  StyleSheet,
  View,
  useWindowDimensions,
} from "react-native";
import { control, space } from "@tournaments-manager/design-tokens";
import type {
  PersonalProfile,
  ProfileMatch,
  GetPersonalProfileSport,
} from "@/api/generated/models";
import { SportStatisticsSport } from "@/api/generated/models/sportStatisticsSport";
import { loadPersonalProfile } from "@/features/personal-profile/api";
import { useCollection } from "@/features/tournament-browsing/use-collection";
import { MatchSchedule } from "@/features/scheduling/match-schedule";
import { IncidentSummary } from "@/features/match-results/incident-summary";
import { getCurrentLanguage, getTranslator } from "@/shared/i18n/locale";
import { useSession } from "@/shared/session/session-provider";
import { getSportLabelKey } from "@/shared/tournaments/sport";
import { ScrollView } from "@/shared/ui/scroll-view";
import {
  Button,
  Card,
  ConfigurationOption,
  ContentSection,
  DisclosureIndicator,
  DialogCloseButton,
  ModalDialog,
  usesLiquidGlassNavigation,
  LoadingTransition,
  NavigationHeaderButton,
  RequestErrorCard,
  Screen,
  Text,
  TextField,
  useTabContentBottomPadding,
} from "@/shared/ui";

const key = (item: ProfileMatch) => item.match.id;
export default function ProfileScreen() {
  const t = getTranslator();
  const { user, isRestoring } = useSession();
  const bottom = useTabContentBottomPadding();
  const { height } = useWindowDimensions();
  const [filtersOpen, setFiltersOpen] = useState(false);
  const [sport, setSport] = useState<GetPersonalProfileSport>();
  const [tournament, setTournament] = useState("");
  const [filters, setFilters] = useState<{ sport?: GetPersonalProfileSport; tournament: string }>({
    tournament: "",
  });
  useEffect(() => {
    setSport(undefined);
    setTournament("");
    setFilters({ tournament: "" });
    setFiltersOpen(false);
  }, [user?.id]);
  const fetchPage = useCallback(
    (cursor: string | undefined, signal: AbortSignal) =>
      loadPersonalProfile({ ...filters, cursor }, signal),
    [filters],
  );
  const { page, error, moreError, loading, refreshing, loadingMore, load } = useCollection<
    ProfileMatch,
    PersonalProfile
  >(user ? `${user.id}:${JSON.stringify(filters)}` : null, fetchPage, key);
  const openFilters = () => {
    setSport(filters.sport);
    setTournament(filters.tournament);
    setFiltersOpen(true);
  };
  if (!user && !isRestoring) return <Redirect href="/account" />;
  return (
    <>
      <Stack.Screen
        options={{
          ...(!usesLiquidGlassNavigation
            ? {
                headerRight: () => (
                  <View style={styles.headerActions}>
                    <NavigationHeaderButton
                      accessibilityLabel={t("profile_filters")}
                      icon="filter"
                      nativeIcon={{
                        android: "filter_alt",
                        ios: "line.3.horizontal.decrease",
                        web: "filter_alt",
                      }}
                      side="none"
                      onPress={openFilters}
                    />
                    {Platform.OS === "web" ? (
                      <NavigationHeaderButton
                        accessibilityLabel={t("profile_refresh")}
                        icon="refresh"
                        nativeIcon={{ android: "refresh", ios: "arrow.clockwise", web: "refresh" }}
                        side="none"
                        loading={loading || refreshing || loadingMore}
                        onPress={() => void load(page ? "refresh" : "initial")}
                      />
                    ) : null}
                  </View>
                ),
              }
            : {}),
        }}
      />
      {usesLiquidGlassNavigation ? (
        <Stack.Toolbar placement="right">
          <Stack.Toolbar.Button
            accessibilityLabel={t("profile_filters")}
            icon="line.3.horizontal.decrease"
            onPress={openFilters}
          />
        </Stack.Toolbar>
      ) : null}
      <Screen navigationTitle={t("profile_title")} topInset="navigation-bar" bottomInset="none">
        {error && !page ? (
          <RequestErrorCard
            message={t(error)}
            actionLabel={t("common_retry")}
            onRetry={() => void load()}
          />
        ) : !page || isRestoring || loading ? (
          <LoadingTransition active message={t("common_loading")} />
        ) : (
          <ScrollView
            contentContainerStyle={[styles.content, { paddingBottom: bottom }]}
            refreshControl={
              Platform.OS !== "web" ? (
                <RefreshControl refreshing={refreshing} onRefresh={() => void load("refresh")} />
              ) : undefined
            }
          >
            <ContentSection>
              <Text variant="title">{user?.username}</Text>
              <Text color="secondary">{t("profile_private")}</Text>
              {filters.sport || filters.tournament ? (
                <Text color="secondary">
                  {t("profile_active_filters")
                    .replace(
                      "{sport}",
                      filters.sport ? t(getSportLabelKey(filters.sport)) : t("profile_all_sports"),
                    )
                    .replace("{tournament}", filters.tournament || t("profile_all_tournaments"))}
                </Text>
              ) : null}
            </ContentSection>
            <ContentSection>
              <Text variant="title">{t("profile_statistics")}</Text>
              <Text color="secondary">{t("profile_statistics_help")}</Text>
              {page.statistics.length === 0 ? (
                <Text>{t("profile_no_statistics")}</Text>
              ) : (
                page.statistics.map((stat) => (
                  <View key={stat.sport} style={styles.stack}>
                    <Text variant="bodyLarge">{t(getSportLabelKey(stat.sport))}</Text>
                    <View style={styles.options}>
                      {(
                        [
                          ["profile_played", stat.played],
                          ["profile_wins", stat.wins],
                          ...(stat.drawsAllowed ? [["profile_draws", stat.draws] as const] : []),
                          ["profile_losses", stat.losses],
                        ] as const
                      ).map(([label, value]) => (
                        <Text key={label}>
                          {t(label)}: {value}
                        </Text>
                      ))}
                    </View>
                    <Text>
                      {t("profile_win_percentage")}:{" "}
                      {stat.winPercentage === null
                        ? t("profile_no_data")
                        : new Intl.NumberFormat(getCurrentLanguage(), {
                            style: "percent",
                            maximumFractionDigits: 1,
                          }).format(stat.winPercentage / 100)}
                    </Text>
                  </View>
                ))
              )}
            </ContentSection>
            <ContentSection>
              <Text variant="title">{t("profile_history")}</Text>
              {!page.items.length ? (
                <Text>
                  {t(
                    filters.sport || filters.tournament
                      ? "profile_empty_filtered"
                      : "profile_empty",
                  )}
                </Text>
              ) : null}
            </ContentSection>
            {page.items.map((item) => (
              <HistoryCard key={item.match.id} item={item} />
            ))}
            {moreError ? (
              <ContentSection>
                <Text color="error">{t(moreError)}</Text>
                <Button
                  label={t("common_retry")}
                  onPress={() => void load("more")}
                  loading={loadingMore}
                />
              </ContentSection>
            ) : page.nextCursor ? (
              <ContentSection>
                <Button
                  variant="secondary"
                  label={t("tournaments_load_more")}
                  loading={loadingMore}
                  onPress={() => void load("more")}
                />
              </ContentSection>
            ) : null}
          </ScrollView>
        )}
      </Screen>
      <ModalDialog
        visible={filtersOpen}
        avoidKeyboard
        dismissAccessibilityLabel={t("common_close")}
        onDismiss={() => setFiltersOpen(false)}
        dialogStyle={{ maxHeight: Math.max(control.minHeight, height - space[12] - space[12]) }}
      >
        <View style={styles.dialogHeader}>
          <Text variant="title" style={styles.dialogTitle}>
            {t("profile_filters")}
          </Text>
          <DialogCloseButton
            accessibilityLabel={t("common_close")}
            onPress={() => setFiltersOpen(false)}
          />
        </View>
        <DialogScrollView
          style={styles.dialogScroll}
          contentContainerStyle={styles.stack}
          keyboardShouldPersistTaps="handled"
        >
          <View style={styles.options}>
            <ConfigurationOption
              label={t("profile_all_sports")}
              selected={!sport}
              onPress={() => setSport(undefined)}
            />
            {Object.values(SportStatisticsSport).map((value) => (
              <ConfigurationOption
                key={value}
                label={t(getSportLabelKey(value))}
                selected={sport === value}
                onPress={() => setSport(value)}
              />
            ))}
          </View>
          <TextField
            label={t("profile_tournament_filter")}
            value={tournament}
            onChangeText={setTournament}
            maxLength={120}
            autoCorrect={false}
          />
          <Button
            label={t("profile_apply_filters")}
            onPress={() => {
              setFilters({ sport, tournament: tournament.trim() });
              setFiltersOpen(false);
            }}
          />
        </DialogScrollView>
      </ModalDialog>
    </>
  );
}
function HistoryCard({ item }: { item: ProfileMatch }) {
  const t = getTranslator();
  const match = item.match;
  const teams = new Map([
    [match.homeTeamId, item.homeTeamName],
    [match.awayTeamId, item.awayTeamName],
  ]);
  return (
    <Card>
      <Pressable
        accessibilityRole="button"
        accessibilityLabel={t("matches_open")
          .replace("{home}", item.homeTeamName || t("bracket_bye"))
          .replace("{away}", item.awayTeamName || t("bracket_bye"))
          .replace("{tournament}", item.tournamentName)}
        style={styles.card}
        onPress={() =>
          router.push({
            pathname: "/tournament/[id]",
            params: { id: item.tournamentId, matchId: match.id },
          })
        }
      >
        <View style={styles.copy}>
          <Text color="secondary">
            {item.tournamentName} · {t(getSportLabelKey(item.sport))}
          </Text>
          <Text>
            {item.homeTeamName || t("bracket_bye")} — {item.awayTeamName || t("bracket_bye")}
          </Text>
          <Text color="secondary">
            {t("profile_my_team")}:{" "}
            {item.ownSide === "home" ? item.homeTeamName : item.awayTeamName}
          </Text>
          <Text color="secondary">
            {t(
              item.stageType === "single_elimination"
                ? "bracket_round"
                : item.stageType === "qualification_tiebreak"
                  ? "tournament_mixed_tiebreak_cycle"
                  : "league_match_round",
            ).replace("{number}", String(match.round))}
          </Text>
          <MatchSchedule match={match} />
          {match.state === "bye" ? (
            <Text>{t("bracket_bye")}</Text>
          ) : (
            <>
              <Text variant="title">
                {match.homeScore} – {match.awayScore}
                {match.homePenalties !== undefined
                  ? ` (${match.homePenalties}–${match.awayPenalties})`
                  : ""}
              </Text>
              {match.sets.length > 0 && !match.incident ? (
                <Text color="secondary">
                  {match.sets.map((set) => `${set.homeScore}–${set.awayScore}`).join(" · ")}
                </Text>
              ) : null}
              {match.resultType === "administrative" ? (
                <Text>{t("profile_administrative")}</Text>
              ) : null}
              <IncidentSummary match={match} teams={teams} showAdministrativeScore />
            </>
          )}
          {match.state === "bye" || match.resultType !== "played" ? (
            <Text color="secondary">{t("profile_excluded")}</Text>
          ) : null}
        </View>
        <DisclosureIndicator />
      </Pressable>
    </Card>
  );
}
const styles = StyleSheet.create({
  headerActions: {
    flexDirection: "row",
    alignItems: "center",
    gap: space[3],
    marginRight:
      Platform.OS === "ios" ? 0 : Platform.OS === "android" ? space[5] - space[4] : space[5],
  },
  dialogHeader: { flexDirection: "row", alignItems: "center", gap: space[3] },
  dialogTitle: { flex: 1, flexShrink: 1 },
  dialogScroll: { flexShrink: 1 },
  content: { gap: space[5] },
  stack: { gap: space[3] },
  options: { flexDirection: "row", flexWrap: "wrap", gap: space[3] },
  copy: { flex: 1, gap: space[2] },
  card: { flexDirection: "row", alignItems: "center", gap: space[3], minHeight: control.minHeight },
});
