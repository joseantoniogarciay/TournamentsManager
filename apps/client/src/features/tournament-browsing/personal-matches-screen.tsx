import { router, useNavigation } from "expo-router";
import type { NativeStackNavigationProp } from "expo-router/native-stack";
import { useCallback, useEffect, useRef, useState } from "react";
import {
  AccessibilityInfo,
  Platform,
  Pressable,
  RefreshControl,
  ScrollView,
  StyleSheet,
  View,
} from "react-native";
import { color, control, radius, space } from "@tournaments-manager/design-tokens";
import type { PersonalMatch, PersonalMatchPage } from "@/api/generated/models";
import { MatchSchedule } from "@/features/scheduling/match-schedule";
import { IncidentSummary } from "@/features/match-results/incident-summary";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { useSession } from "@/shared/session/session-provider";
import { getTranslator } from "@/shared/i18n/locale";
import {
  Button,
  Card,
  DisclosureIndicator,
  LoadingTransition,
  RequestErrorCard,
  ResponsiveColumns,
  Screen,
  Text,
  useTabContentBottomPadding,
} from "@/shared/ui";
import { getPersonalMatches } from "./api";
import { useCollection } from "./use-collection";

const matchKey = (item: PersonalMatch) => item.match.id;
export function PersonalMatchesScreen() {
  const t = getTranslator();
  const { user, isRestoring } = useSession();
  const { colors } = usePreferences();
  const bottom = useTabContentBottomPadding();
  const [view, setView] = useState<"pending" | "results">("pending");
  const navigation = useNavigation();
  const rootNavigation = useNavigation<NativeStackNavigationProp<Record<string, undefined>>>("/");
  const cardRefs = useRef(new Map<string, View>());
  const returnMatch = useRef<string | null>(null);
  useEffect(() => {
    returnMatch.current = null;
  }, [user?.id, view]);
  useEffect(() => {
    if (Platform.OS === "web") return;
    let frame: number | undefined;
    const unsubscribe = rootNavigation.addListener("transitionEnd", (event) => {
      if (event.data.closing || !navigation.isFocused() || !returnMatch.current) return;
      // Restore after the native modal has disappeared, so its focus reset
      // cannot replace the originating card with the first segment.
      frame = requestAnimationFrame(() => {
        const matchId = returnMatch.current;
        void AccessibilityInfo.isScreenReaderEnabled().then((enabled) => {
          if (!navigation.isFocused() || returnMatch.current !== matchId) return;
          returnMatch.current = null;
          const card = matchId ? cardRefs.current.get(matchId) : undefined;
          if (enabled && card) AccessibilityInfo.sendAccessibilityEvent(card, "focus");
        });
      });
    });
    return () => {
      unsubscribe();
      if (frame !== undefined) cancelAnimationFrame(frame);
    };
  }, [navigation, rootNavigation]);
  const fetchPage = useCallback(
    (cursor: string | undefined, signal: AbortSignal) => getPersonalMatches(view, cursor, signal),
    [view],
  );
  const collection = useCollection<PersonalMatch, PersonalMatchPage>(
    user ? `${user.id}:${view}` : null,
    fetchPage,
    matchKey,
  );
  const { page, error, loading, refreshing, loadingMore, moreError, load } = collection;
  return (
    <Screen bottomInset="none" webWidth="wide">
      {!isRestoring && !user ? (
        <ScrollView contentContainerStyle={[styles.content, { paddingBottom: bottom }]}>
          <Card>
            <View style={styles.copy}>
              <Text variant="title">{t("matches_guest_title")}</Text>
              <Text color="secondary">{t("matches_guest_description")}</Text>
              <Button label={t("home_sign_in")} onPress={() => router.push("/account")} />
            </View>
          </Card>
        </ScrollView>
      ) : null}
      {user ? (
        <>
          <View
            accessibilityRole="tablist"
            style={[styles.segments, { backgroundColor: colors.surface.subtle }]}
          >
            {(["pending", "results"] as const).map((option) => (
              <Pressable
                key={option}
                accessibilityRole="tab"
                aria-selected={view === option}
                onPress={() => setView(option)}
                style={[
                  styles.segment,
                  view === option ? { backgroundColor: color.brand.primary } : undefined,
                ]}
              >
                <Text color={view === option ? "onBrand" : "primary"}>
                  {t(option === "pending" ? "matches_pending" : "matches_results")}
                </Text>
              </Pressable>
            ))}
          </View>
          <ScrollView
            contentContainerStyle={[styles.content, { paddingBottom: bottom }]}
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
            {page && page.items.length === 0 ? (
              <View style={styles.empty}>
                <Text color="secondary">
                  {t(
                    !page.hasMatches
                      ? "matches_empty"
                      : view === "pending"
                        ? "matches_pending_empty"
                        : "matches_results_empty",
                  )}
                </Text>
                {!page.hasMatches ? (
                  <Button
                    label={t("nav_tournaments")}
                    variant="secondary"
                    onPress={() => router.push("/tournaments")}
                  />
                ) : null}
              </View>
            ) : null}
            {page ? (
              <ResponsiveColumns>
                {page.items.map((item) => (
                  <PersonalMatchCard
                    key={item.match.id}
                    item={item}
                    cardRef={(node) => {
                      if (node) cardRefs.current.set(item.match.id, node);
                      else cardRefs.current.delete(item.match.id);
                    }}
                    onOpen={() => {
                      returnMatch.current = item.match.id;
                      router.push({
                        pathname: "/tournament/[id]",
                        params: { id: item.tournamentId, matchId: item.match.id },
                      });
                    }}
                  />
                ))}
              </ResponsiveColumns>
            ) : null}
            {moreError ? (
              <View style={styles.pagination}>
                <Text color="secondary">{t(moreError)}</Text>
              </View>
            ) : null}
            {page?.nextCursor ? (
              <View style={styles.pagination}>
                <Button
                  label={t(moreError ? "common_retry" : "tournaments_load_more")}
                  loading={loadingMore}
                  disabled={refreshing}
                  variant="secondary"
                  onPress={() => void load("more")}
                />
              </View>
            ) : null}
          </ScrollView>
        </>
      ) : null}
      {isRestoring || (user && ((!page && !error) || loading)) ? (
        <LoadingTransition active message={t("common_loading")} />
      ) : null}
    </Screen>
  );
}
function PersonalMatchCard({
  item,
  cardRef,
  onOpen,
}: {
  item: PersonalMatch;
  cardRef: (node: View | null) => void;
  onOpen: () => void;
}) {
  const t = getTranslator();
  const match = item.match;
  const teams = new Map([
    [match.homeTeamId, item.homeTeamName],
    [match.awayTeamId, item.awayTeamName],
  ]);
  return (
    <Card>
      <Pressable
        ref={cardRef}
        accessible
        accessibilityRole="button"
        accessibilityLabel={t("matches_open")
          .replace("{home}", item.homeTeamName)
          .replace("{away}", item.awayTeamName)
          .replace("{tournament}", item.tournamentName)}
        style={styles.card}
        onPress={onOpen}
      >
        <View style={styles.copy}>
          <Text color="secondary">{item.tournamentName}</Text>
          <Text>
            {item.homeTeamName} — {item.awayTeamName}
          </Text>
          <Text color="secondary">
            {t(
              item.stageType === "qualification_tiebreak"
                ? "tournament_mixed_phase_tiebreak"
                : item.stageType === "single_elimination"
                  ? "tournament_mixed_phase_elimination"
                  : "tournament_mixed_phase_league",
            )}{" "}
            ·{" "}
            {t(
              item.stageType === "single_elimination"
                ? "bracket_round"
                : item.stageType === "qualification_tiebreak"
                  ? "tournament_mixed_tiebreak_cycle"
                  : "league_match_round",
            ).replace("{number}", String(match.round))}
          </Text>
          {match.state === "completed" ? (
            <>
              <Text variant="title">
                {match.homeScore} – {match.awayScore}
              </Text>
              {match.sets.length > 0 && !match.incident ? (
                <Text color="secondary">
                  {match.sets.map((set) => `${set.homeScore}–${set.awayScore}`).join(" · ")}
                </Text>
              ) : null}
              <IncidentSummary match={match} teams={teams} showAdministrativeScore />
            </>
          ) : null}
          {!match.scheduledDate && match.state === "pending" ? (
            <Text color="secondary">{t("matches_unscheduled")}</Text>
          ) : null}
          <MatchSchedule match={match} />
        </View>
        <DisclosureIndicator />
      </Pressable>
    </Card>
  );
}
const styles = StyleSheet.create({
  content: { flexGrow: 1, gap: space[5] },
  copy: { flex: 1, gap: space[2] },
  card: { flexDirection: "row", alignItems: "center", gap: space[3], minHeight: control.minHeight },
  empty: {
    flex: 1,
    justifyContent: "center",
    alignItems: "center",
    paddingHorizontal: space[5],
    gap: space[5],
  },
  pagination: { paddingHorizontal: space[5], gap: space[3] },
  segments: {
    flexDirection: "row",
    marginHorizontal: space[5],
    marginBottom: space[5],
    borderRadius: radius.control,
    padding: space[1],
  },
  segment: {
    flex: 1,
    minHeight: control.minHeight,
    justifyContent: "center",
    alignItems: "center",
    borderRadius: radius.control - 2,
  },
});
