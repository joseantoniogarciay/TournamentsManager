import { router, Stack, useFocusEffect, useLocalSearchParams } from "expo-router";
import { SymbolView } from "expo-symbols";
import { useCallback, useEffect, useRef, useState } from "react";
import {
  Platform,
  Pressable,
  ScrollView,
  SectionList,
  Share,
  StyleSheet,
  View,
} from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";

import { color, control, radius, space, typography } from "@tournaments-manager/design-tokens";

import { APIUnexpectedResponseError } from "@/api/fetch";
import type { PublicTournament } from "@/api/generated/models";
import {
  cancelTournamentRequest,
  completeTournamentRequest,
  getTournamentRelationship,
  TournamentUnavailableError,
  startTournamentRequest,
} from "@/features/league-creation/api";
import { useTournament, useTournamentStore } from "@/features/league-creation/league-store";
import { MatchResultConflictError, recordMatchResultRequest } from "@/features/match-results/api";
import {
  BracketIntro,
  BracketRoundNavigation,
  BracketView,
} from "@/features/match-results/bracket-view";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import { getTournamentStateLabel } from "@/shared/i18n/league-state";
import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { useSession } from "@/shared/session/session-provider";
import {
  Button,
  Card,
  LoadingTransition,
  ModalDialog,
  NavigationHeaderButton,
  RequestErrorCard,
  Screen,
  Text,
  TextField,
  useConfirmationDialog,
  usesLiquidGlassNavigation,
} from "@/shared/ui";

const localAppLinkURL = "http://localhost:8082";

export default function TournamentScreen() {
  const t = getTranslator();
  const { id } = useLocalSearchParams<{ id: string }>();
  const { user } = useSession();
  const { colors } = usePreferences();
  const insets = useSafeAreaInsets();
  const { show } = useFeedback();
  const { confirm } = useConfirmationDialog();
  const league = useTournament(id);
  const { loadTournament, putTournament, refreshTournament } = useTournamentStore();
  const [relationship, setRelationship] = useState<string>();
  const [loadErrorMessage, setLoadErrorMessage] = useState<string>();
  const [leagueUnavailable, setTournamentUnavailable] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [isStarting, setIsStarting] = useState(false);
  const [roundRobinLegs, setRoundRobinLegs] = useState<1 | 2>(1);
  const [format, setFormat] = useState<"league" | "single_elimination">("league");
  const [menuOpen, setMenuOpen] = useState(false);
  const [scores, setScores] = useState<
    Record<string, { home: string; away: string; homePenalties: string; awayPenalties: string }>
  >({});
  const [savingMatchID, setSavingMatchID] = useState<string>();
  const [editingMatchID, setEditingMatchID] = useState<string>();
  const [isCompleting, setIsCompleting] = useState(false);
  const [isCancelling, setIsCancelling] = useState(false);
  const [completionConfirmationOpen, setCompletionConfirmationOpen] = useState(false);
  const [completionOpen, setCompletionOpen] = useState(false);
  const [selectedBracketRound, setSelectedBracketRound] = useState(1);
  const [bracketRoundSelectionRevision, setBracketRoundSelectionRevision] = useState(0);
  const matchList = useRef<SectionList<PublicTournament["matches"][number]>>(null);
  const bracketList = useRef<ScrollView>(null);
  const matchListOffset = useRef(0);
  const matchListViewport = useRef<View>(null);
  const ensureBracketMatchVisible = useCallback((matchView: View) => {
    const viewport = matchListViewport.current;
    if (!viewport) return;
    viewport.measureInWindow((_viewportX, viewportY, _viewportWidth, viewportHeight) => {
      matchView.measureInWindow((_matchX, matchY, _matchWidth, matchHeight) => {
        const visibleTop = viewportY + control.minHeight + space[6];
        const visibleBottom = viewportY + viewportHeight - control.minHeight - space[5];
        const visibleHeight = visibleBottom - visibleTop;
        const offsetDelta =
          matchHeight > visibleHeight || matchY < visibleTop
            ? matchY - visibleTop
            : matchY + matchHeight > visibleBottom
              ? matchY + matchHeight - visibleBottom
              : 0;
        if (Math.abs(offsetDelta) < 1) return;
        bracketList.current?.scrollTo({
          animated: true,
          y: Math.max(0, matchListOffset.current + offsetDelta),
        });
      });
    });
  }, []);
  useEffect(() => {
    setSelectedBracketRound(1);
    setBracketRoundSelectionRevision(0);
  }, [id]);
  const load = useCallback(
    async (force = false) => {
      if (!id) {
        setLoadErrorMessage(t("common_request_error"));
        return;
      }
      setIsLoading(true);
      setLoadErrorMessage(undefined);
      setTournamentUnavailable(false);
      try {
        await (force ? refreshTournament(id) : loadTournament(id));
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
    [id, loadTournament, refreshTournament, t],
  );
  useEffect(() => {
    void load();
  }, [load]);
  useFocusEffect(
    useCallback(() => {
      if (!user) return;
      void getTournamentRelationship(id)
        .then(setRelationship)
        .catch(() => setRelationship(undefined));
    }, [id, user]),
  );
  const isOrganizer = relationship === "organizer";
  const canManageResults = relationship === "organizer" || relationship === "delegated";
  const start = async () => {
    if (!id || isStarting) return;
    setIsStarting(true);
    try {
      putTournament(
        await startTournamentRequest(
          id,
          format === "league" ? { format, roundRobinLegs } : { format },
        ),
      );
    } catch (error) {
      const failure = getRequestFailure(error);
      show({ kind: failure.kind, message: t(failure.messageKey) });
    } finally {
      setIsStarting(false);
    }
  };
  const share = async () => {
    if (!id || !league) return;
    const base = (
      process.env.EXPO_PUBLIC_APP_LINK_URL ??
      (process.env.APP_ENV === "production" ? undefined : localAppLinkURL)
    )?.replace(/\/$/, "");
    if (!base) {
      show({ kind: "generic-error", message: t("common_request_error") });
      return;
    }
    await Share.share({
      message: `${league.name}: ${base}/tournament/${id}`,
    });
  };
  const cancel = () => {
    if (isCancelling) return;
    confirm({
      title: t("league_cancel_title"),
      description: t("league_cancel_description"),
      acceptLabel: t("league_cancel"),
      acceptVariant: "destructive",
      cancelLabel: t("common_cancel"),
      onAccept: () => {
        if (!id) return;
        setIsCancelling(true);
        void cancelTournamentRequest(id)
          .then(putTournament)
          .catch((error) => {
            const failure = getRequestFailure(error);
            show({ kind: failure.kind, message: t(failure.messageKey) });
          })
          .finally(() => setIsCancelling(false));
      },
      onCancel: () => undefined,
    });
  };
  const complete = () => setCompletionConfirmationOpen(true);
  const confirmCompletion = () => {
    if (!id) return;
    setIsCompleting(true);
    void completeTournamentRequest(id)
      .then((completed) => {
        putTournament(completed);
        setCompletionConfirmationOpen(false);
        setCompletionOpen(true);
      })
      .catch((error) => {
        if (error instanceof APIUnexpectedResponseError && error.status === 409) {
          setCompletionConfirmationOpen(false);
          show({ kind: "success", message: t("league_completion_already_completed") });
          void refreshTournament(id).catch((refreshError) => {
            const failure = getRequestFailure(refreshError);
            show({ kind: failure.kind, message: t(failure.messageKey) });
          });
          return;
        }
        const failure = getRequestFailure(error);
        show({ kind: failure.kind, message: t(failure.messageKey) });
      })
      .finally(() => setIsCompleting(false));
  };
  const saveResult = async (matchID: string) => {
    if (!id || savingMatchID) return;
    const score = scores[matchID];
    const homeScore = Number(score?.home);
    const awayScore = Number(score?.away);
    if (
      !Number.isInteger(homeScore) ||
      !Number.isInteger(awayScore) ||
      homeScore < 0 ||
      awayScore < 0
    )
      return;
    setSavingMatchID(matchID);
    try {
      const shootout = league?.format === "single_elimination" && homeScore === awayScore;
      putTournament(
        await recordMatchResultRequest(id, matchID, {
          homeScore,
          awayScore,
          ...(shootout
            ? {
                homePenalties: Number(score?.homePenalties),
                awayPenalties: Number(score?.awayPenalties),
              }
            : {}),
        }),
      );
      setEditingMatchID(undefined);
    } catch (error) {
      if (error instanceof MatchResultConflictError) {
        setEditingMatchID(undefined);
        show({ kind: "generic-error", message: t("bracket_result_conflict") });
        void refreshTournament(id).catch(() => undefined);
        return;
      }
      const failure = getRequestFailure(error);
      show({ kind: failure.kind, message: t(failure.messageKey) });
    } finally {
      setSavingMatchID(undefined);
    }
  };
  const returnToPreviousScreen = () => {
    if (router.canDismiss()) {
      router.dismiss();
      return;
    }
    router.replace("/");
  };
  if (!league) {
    const loadingHeaderOptions = {
      headerBackVisible: false,
      headerShadowVisible: false,
      headerStyle: { backgroundColor: colors.surface.canvas },
      headerTintColor: colors.text.primary,
      headerTitle: t("league_title"),
      headerTitleAlign: "center" as const,
    };
    return (
      <>
        <Stack.Screen
          options={{
            ...loadingHeaderOptions,
            ...(!usesLiquidGlassNavigation
              ? {
                  headerLeft: () => (
                    <NavigationHeaderButton
                      accessibilityLabel={t("common_close")}
                      icon="close"
                      nativeIcon={{ android: "close", ios: "xmark", web: "close" }}
                      onPress={returnToPreviousScreen}
                    />
                  ),
                }
              : {}),
          }}
        >
          {usesLiquidGlassNavigation ? (
            <Stack.Toolbar placement="left">
              <Stack.Toolbar.Button
                accessibilityLabel={t("common_close")}
                icon="xmark"
                onPress={returnToPreviousScreen}
              />
            </Stack.Toolbar>
          ) : null}
        </Stack.Screen>
        <Screen topInset="navigation-bar">
          {loadErrorMessage ? (
            <RequestErrorCard
              actionLabel={t(leagueUnavailable ? "common_close" : "common_retry")}
              loading={leagueUnavailable ? false : isLoading}
              message={loadErrorMessage}
              onRetry={leagueUnavailable ? returnToPreviousScreen : () => void load(true)}
            />
          ) : (
            <LoadingTransition active message={t("common_loading")} />
          )}
        </Screen>
      </>
    );
  }
  const canCancel = league.state === "published" || league.state === "in_progress";
  const hasStarted =
    league.state === "in_progress" ||
    league.state === "completed" ||
    (league.state === "cancelled" && league.matches.length > 0);
  const canComplete =
    isOrganizer &&
    league.state === "in_progress" &&
    league.matches.length > 0 &&
    league.matches.every((match) => match.state === "completed" || match.state === "bye");
  const primaryTournamentAction = (() => {
    switch (league.state) {
      case "published":
        return isOrganizer
          ? { label: t("league_start"), loading: isStarting, onPress: () => void start() }
          : undefined;
      case "in_progress":
        return canComplete
          ? { label: t("league_complete"), loading: isCompleting, onPress: complete }
          : undefined;
      default:
        return undefined;
    }
  })();
  const teamsByID = new Map(league.teams.map((team) => [team.id, team.name]));
  const editingMatch = league.matches.find((match) => match.id === editingMatchID);
  const editingScore = editingMatch
    ? (scores[editingMatch.id] ?? {
        home: editingMatch.homeScore?.toString() ?? "",
        away: editingMatch.awayScore?.toString() ?? "",
        homePenalties: editingMatch.homePenalties?.toString() ?? "",
        awayPenalties: editingMatch.awayPenalties?.toString() ?? "",
      })
    : undefined;
  const needsShootout =
    league.format === "single_elimination" &&
    editingScore !== undefined &&
    /^\d+$/.test(editingScore.home) &&
    /^\d+$/.test(editingScore.away) &&
    Number(editingScore.home) === Number(editingScore.away);
  const canSaveResult =
    editingScore !== undefined &&
    /^\d+$/.test(editingScore.home) &&
    /^\d+$/.test(editingScore.away) &&
    (!needsShootout ||
      (/^\d+$/.test(editingScore.homePenalties) &&
        /^\d+$/.test(editingScore.awayPenalties) &&
        Number(editingScore.homePenalties) !== Number(editingScore.awayPenalties)));
  const openResultEditor = (matchID: string) => {
    const match = league.matches.find((item) => item.id === matchID);
    if (!match) return;
    setScores((value) => ({
      ...value,
      [matchID]: {
        home: match.homeScore?.toString() ?? "",
        away: match.awayScore?.toString() ?? "",
        homePenalties: match.homePenalties?.toString() ?? "",
        awayPenalties: match.awayPenalties?.toString() ?? "",
      },
    }));
    setEditingMatchID(matchID);
  };
  const matchesByRound = new Map<number, PublicTournament["matches"]>();
  for (const match of league.matches) {
    const matches = matchesByRound.get(match.round) ?? [];
    matches.push(match);
    matchesByRound.set(match.round, matches);
  }
  const matchSections = [...matchesByRound.entries()]
    .sort(([firstRound], [secondRound]) => firstRound - secondRound)
    .map(([round, data]) => ({ data, round }));
  const closeWebMenu = () => setMenuOpen(false);
  const openAdministrators = () => router.push(`/tournament/${league.id}/administrators`);
  const openTransfer = () => router.push(`/tournament/${league.id}/transfer`);
  const headerOptions = {
    headerBackVisible: false,
    headerShadowVisible: false,
    headerStyle: { backgroundColor: colors.surface.canvas },
    headerTintColor: colors.text.primary,
    headerTitleAlign: "center" as const,
    headerTitle: () => (
      <Text numberOfLines={2} style={styles.navigationTitle} variant="bodyLarge">
        {league.name}
      </Text>
    ),
  };
  const bracketRounds = [...new Set(league.matches.map((match) => match.round))].sort(
    (first, second) => first - second,
  );
  const showsBracket = league.format === "single_elimination" && league.matches.length > 0;
  const selectBracketRound = (round: number) => {
    setSelectedBracketRound(round);
    setBracketRoundSelectionRevision((revision) => revision + 1);
  };
  const tournamentSummary = (
    <>
      <Card>
        <View style={styles.stack}>
          <View style={styles.summaryList}>
            <View style={styles.summaryItem}>
              <View
                accessible={false}
                style={[styles.bullet, { backgroundColor: colors.text.secondary }]}
              />
              <Text color="secondary" style={styles.summaryText}>
                <Text style={styles.summaryLabel}>{t("league_creator_label")}</Text>
                {t("league_creator_permissions")}
              </Text>
            </View>
            <View style={styles.summaryItem}>
              <View
                accessible={false}
                style={[styles.bullet, { backgroundColor: colors.text.secondary }]}
              />
              <Text color="secondary" style={styles.summaryText}>
                <Text style={styles.summaryLabel}>{t("league_status_label")}</Text>
                {getTournamentStateLabel(t, league.state)}
              </Text>
            </View>
          </View>
        </View>
      </Card>
      <View style={styles.summaryActions}>
        <View style={styles.summaryAction}>
          <Button
            label={t("league_teams")}
            onPress={() => router.push(`/tournament/${league.id}/teams`)}
            variant="secondary"
          />
        </View>
        {league.format === "league" && hasStarted ? (
          <View style={styles.summaryAction}>
            <Button
              label={t("league_standings")}
              onPress={() => router.push(`/tournament/${league.id}/standings`)}
              variant="secondary"
            />
          </View>
        ) : null}
      </View>
    </>
  );
  return (
    <>
      <Stack.Screen options={headerOptions} />
      {!usesLiquidGlassNavigation ? (
        <Stack.Screen
          options={{
            headerLeft: () => (
              <NavigationHeaderButton
                accessibilityLabel={t("common_back")}
                icon="close"
                nativeIcon={{ android: "close", ios: "xmark", web: "close" }}
                onPress={returnToPreviousScreen}
              />
            ),
            headerRight: () => (
              <NavigationHeaderButton
                accessibilityLabel={t("league_actions")}
                icon="more"
                nativeIcon={{ android: "more_vert", ios: "ellipsis", web: "more_vert" }}
                onPress={() => setMenuOpen((open) => !open)}
                side="right"
              />
            ),
          }}
        />
      ) : (
        <>
          <Stack.Toolbar placement="left">
            <Stack.Toolbar.Button
              accessibilityLabel={t("common_back")}
              icon="xmark"
              onPress={returnToPreviousScreen}
            />
          </Stack.Toolbar>
          <Stack.Toolbar placement="right">
            <Stack.Toolbar.Button
              accessibilityLabel={t("league_actions")}
              icon="ellipsis"
              onPress={() => setMenuOpen((open) => !open)}
            />
          </Stack.Toolbar>
        </>
      )}
      <Screen bottomInset="none" topInset="navigation-bar">
        <View ref={matchListViewport} style={styles.listViewport}>
          {showsBracket ? (
            <ScrollView
              ref={bracketList}
              contentContainerStyle={[
                styles.content,
                {
                  paddingBottom:
                    insets.bottom +
                    (primaryTournamentAction ? control.minHeight + space[3] + space[5] : space[4]),
                },
              ]}
              onScroll={(event) => {
                matchListOffset.current = event.nativeEvent.contentOffset.y;
              }}
              scrollEventThrottle={16}
              showsVerticalScrollIndicator={false}
              stickyHeaderIndices={Platform.OS === "web" ? undefined : [1]}
            >
              <View style={styles.listHeader}>
                {tournamentSummary}
                <BracketIntro />
              </View>
              <BracketRoundNavigation
                onSelect={selectBracketRound}
                rounds={bracketRounds}
                selectedRound={selectedBracketRound}
              />
              <BracketView
                tournament={league}
                canManage={canManageResults}
                horizontalControlBottomOffset={
                  primaryTournamentAction
                    ? insets.bottom + space[3] + control.minHeight + space[2]
                    : insets.bottom
                }
                onMatchFocus={ensureBracketMatchVisible}
                onEdit={openResultEditor}
                onRoundChange={setSelectedBracketRound}
                roundSelectionRevision={bracketRoundSelectionRevision}
                selectedRound={selectedBracketRound}
              />
            </ScrollView>
          ) : (
            <SectionList
              ref={matchList}
              contentContainerStyle={[
                styles.content,
                {
                  paddingBottom:
                    insets.bottom +
                    (primaryTournamentAction ? control.minHeight + space[3] + space[5] : space[4]),
                },
              ]}
              ItemSeparatorComponent={() => <View style={styles.matchSeparator} />}
              onScroll={(event) => {
                matchListOffset.current = event.nativeEvent.contentOffset.y;
              }}
              scrollEventThrottle={16}
              sections={league.format === "single_elimination" ? [] : matchSections}
              showsVerticalScrollIndicator={false}
              stickySectionHeadersEnabled
              ListHeaderComponent={
                <View style={styles.listHeader}>
                  {tournamentSummary}
                  {league.state === "published" && isOrganizer ? (
                    <Card>
                      <View style={styles.stack}>
                        <Text style={styles.configurationTitle} variant="bodyLarge">
                          {t("league_start_title")}
                        </Text>
                        <View style={styles.configurationOptions}>
                          <ConfigurationOption
                            label={t("tournament_format_league")}
                            selected={format === "league"}
                            disabled={isStarting}
                            onPress={() => setFormat("league")}
                          />
                          <ConfigurationOption
                            label={t("tournament_format_bracket")}
                            selected={format === "single_elimination"}
                            disabled={isStarting}
                            onPress={() => setFormat("single_elimination")}
                          />
                        </View>
                        {format === "single_elimination" ? (
                          <Text color="secondary">{t("bracket_configuration_help")}</Text>
                        ) : (
                          <View
                            style={[
                              styles.configurationOptions,
                              { borderColor: colors.border.default },
                            ]}
                          >
                            <ConfigurationOption
                              label={t("league_start_one_leg")}
                              selected={roundRobinLegs === 1}
                              disabled={isStarting}
                              onPress={() => setRoundRobinLegs(1)}
                            />
                            <ConfigurationOption
                              label={t("league_start_two_legs")}
                              selected={roundRobinLegs === 2}
                              disabled={isStarting}
                              onPress={() => setRoundRobinLegs(2)}
                            />
                          </View>
                        )}
                      </View>
                    </Card>
                  ) : null}
                </View>
              }
              renderItem={({ item: match }) => {
                return (
                  <Card>
                    <View style={styles.match}>
                      <View style={styles.matchSummary}>
                        <Text style={styles.teamName} variant="bodyLarge">
                          {teamsByID.get(match.homeTeamId)}
                        </Text>
                        <Text style={styles.matchScore} variant="title">
                          {match.state === "completed"
                            ? `${match.homeScore} – ${match.awayScore}`
                            : "–"}
                        </Text>
                        <Text style={styles.teamName} variant="bodyLarge">
                          {teamsByID.get(match.awayTeamId)}
                        </Text>
                      </View>
                      {canManageResults && league.state === "in_progress" ? (
                        <Button
                          label={
                            match.state === "completed"
                              ? t("league_result_edit")
                              : t("league_result_add")
                          }
                          onPress={() => openResultEditor(match.id)}
                          variant={match.state === "completed" ? "secondary" : "primary"}
                        />
                      ) : null}
                    </View>
                  </Card>
                );
              }}
              renderSectionHeader={({ section }) => (
                <View style={[styles.roundHeader, { backgroundColor: colors.surface.canvas }]}>
                  <Text variant="title">
                    {t("league_match_round").replace("{number}", String(section.round))}
                  </Text>
                </View>
              )}
            />
          )}
        </View>
        {primaryTournamentAction ? (
          <View style={[styles.floatingAction, { bottom: insets.bottom + space[3] }]}>
            <Button
              label={primaryTournamentAction.label}
              loading={primaryTournamentAction.loading}
              onPress={primaryTournamentAction.onPress}
            />
          </View>
        ) : null}
        <ModalDialog
          dismissAccessibilityLabel={t("common_close")}
          onDismiss={closeWebMenu}
          visible={menuOpen}
        >
          <View style={styles.menuActions}>
            <Button
              label={t("league_share")}
              onPress={() => {
                closeWebMenu();
                void share();
              }}
              variant="secondary"
            />
            {isOrganizer ? (
              <>
                <Button
                  label={t("tournament_administrators")}
                  onPress={() => {
                    closeWebMenu();
                    openAdministrators();
                  }}
                  variant="secondary"
                />
                {league.state !== "cancelled" ? (
                  <Button
                    label={t("league_transfer")}
                    onPress={() => {
                      closeWebMenu();
                      openTransfer();
                    }}
                    variant="destructive"
                  />
                ) : null}
                {canCancel && !isCancelling ? (
                  <Button
                    label={t("league_cancel")}
                    onPress={() => {
                      closeWebMenu();
                      cancel();
                    }}
                    variant="destructive"
                  />
                ) : null}
              </>
            ) : null}
          </View>
        </ModalDialog>
        <ModalDialog
          dismissAccessibilityLabel={t("common_cancel")}
          onDismiss={() => {
            if (!isCompleting) setCompletionConfirmationOpen(false);
          }}
          visible={completionConfirmationOpen}
        >
          <View style={styles.completionConfirmationCopy}>
            <Text style={styles.completionTitle} variant="title">
              {t("league_complete_title")}
            </Text>
            <Text color="secondary" style={styles.completionChampion}>
              {t("league_complete_description")}
            </Text>
          </View>
          <View style={styles.completionConfirmationActions}>
            <Button
              label={t("league_complete")}
              loading={isCompleting}
              onPress={confirmCompletion}
            />
            <Button
              disabled={isCompleting}
              label={t("common_cancel")}
              onPress={() => setCompletionConfirmationOpen(false)}
              variant="secondary"
            />
          </View>
        </ModalDialog>
        <ModalDialog
          dismissAccessibilityLabel={t("common_close")}
          onDismiss={() => {
            if (!savingMatchID) setEditingMatchID(undefined);
          }}
          visible={editingMatch !== undefined}
        >
          {editingMatch && editingScore ? (
            <View style={styles.stack}>
              <View style={styles.stack}>
                <Text variant="title">
                  {editingMatch.state === "completed"
                    ? t("league_result_edit")
                    : t("league_result_add")}
                </Text>
                <Text color="secondary">{`${teamsByID.get(editingMatch.homeTeamId)} — ${teamsByID.get(editingMatch.awayTeamId)}`}</Text>
              </View>
              <View style={styles.scoreFields}>
                <View style={styles.scoreField}>
                  <TextField
                    label={t("league_home_score")}
                    keyboardType="number-pad"
                    onChangeText={(home) =>
                      setScores((value) => ({
                        ...value,
                        [editingMatch.id]: { ...editingScore, home },
                      }))
                    }
                    value={editingScore.home}
                  />
                </View>
                <View style={styles.scoreField}>
                  <TextField
                    label={t("league_away_score")}
                    keyboardType="number-pad"
                    onChangeText={(away) =>
                      setScores((value) => ({
                        ...value,
                        [editingMatch.id]: { ...editingScore, away },
                      }))
                    }
                    value={editingScore.away}
                  />
                </View>
              </View>
              {needsShootout ? (
                <>
                  <Text color="secondary">{t("bracket_penalties_help")}</Text>
                  <View style={styles.scoreFields}>
                    <View style={styles.scoreField}>
                      <TextField
                        label={t("bracket_home_penalties")}
                        keyboardType="number-pad"
                        value={editingScore.homePenalties}
                        onChangeText={(homePenalties) =>
                          setScores((value) => ({
                            ...value,
                            [editingMatch.id]: { ...editingScore, homePenalties },
                          }))
                        }
                      />
                    </View>
                    <View style={styles.scoreField}>
                      <TextField
                        label={t("bracket_away_penalties")}
                        keyboardType="number-pad"
                        value={editingScore.awayPenalties}
                        onChangeText={(awayPenalties) =>
                          setScores((value) => ({
                            ...value,
                            [editingMatch.id]: { ...editingScore, awayPenalties },
                          }))
                        }
                      />
                    </View>
                  </View>
                </>
              ) : null}
              <Button
                disabled={!canSaveResult}
                label={t("league_result_save")}
                loading={savingMatchID === editingMatch.id}
                onPress={() => void saveResult(editingMatch.id)}
              />
              <LoadingTransition
                active={savingMatchID === editingMatch.id}
                message={t("league_result_saving")}
              />
            </View>
          ) : null}
        </ModalDialog>
        <ModalDialog
          dismissAccessibilityLabel={t("common_close")}
          onDismiss={() => setCompletionOpen(false)}
          visible={completionOpen}
        >
          <View style={styles.completionSuccessContent}>
            <SymbolView name="trophy.fill" size={56} tintColor={colors.feedback.success} />
            <Text style={styles.completionTitle} variant="title">
              {t("league_completion_title")}
            </Text>
            <Text style={styles.completionChampion} variant="display">
              {(league.championTeamIds ?? [])
                .map((teamID) => teamsByID.get(teamID))
                .filter((name): name is string => Boolean(name))
                .join(" · ")}
            </Text>
            <Text color="secondary">
              {league.championTeamIds.length > 1
                ? t("league_completion_co_champions")
                : t("league_completion_champion")}
            </Text>
          </View>
          <View style={styles.completionSuccessActions}>
            {league.format === "league" ? (
              <Button
                label={t("league_completion_view_standings")}
                onPress={() => {
                  setCompletionOpen(false);
                  router.push(`/tournament/${league.id}/standings`);
                }}
              />
            ) : null}
            <Button
              label={t("common_close")}
              onPress={() => setCompletionOpen(false)}
              variant="secondary"
            />
          </View>
        </ModalDialog>
      </Screen>
    </>
  );
}

function ConfigurationOption({
  disabled,
  label,
  onPress,
  selected,
}: {
  disabled: boolean;
  label: string;
  onPress: () => void;
  selected: boolean;
}) {
  const { colors } = usePreferences();

  return (
    <Pressable
      accessibilityRole="button"
      accessibilityState={{ disabled, selected }}
      disabled={disabled}
      onPress={onPress}
      style={[
        styles.configurationChip,
        selected
          ? { backgroundColor: color.brand.primary, borderColor: color.brand.primary }
          : { backgroundColor: colors.surface.default, borderColor: colors.border.default },
        disabled ? styles.configurationChipDisabled : undefined,
      ]}
    >
      <Text color={selected ? "onBrand" : "primary"}>{label}</Text>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  content: { paddingBottom: space[4] },
  listViewport: { flex: 1 },
  configurationChip: {
    alignItems: "center",
    borderRadius: radius.pill,
    borderWidth: 1,
    justifyContent: "center",
    minHeight: control.minHeight,
    paddingHorizontal: control.horizontalPadding,
  },
  configurationChipDisabled: { opacity: 0.55 },
  configurationOptions: {
    borderTopWidth: 1,
    flexDirection: "row",
    flexWrap: "wrap",
    gap: space[2],
    paddingTop: space[5],
  },
  configurationTitle: { fontFamily: typography.family.semibold },
  floatingAction: {
    left: space[5],
    position: "absolute",
    right: space[5],
  },
  listHeader: { gap: space[5], paddingBottom: space[5] },
  stack: { flex: 1, gap: space[3] },
  bullet: {
    borderRadius: radius.pill,
    height: space[1],
    marginTop: space[2],
    width: space[1],
  },
  summaryAction: { flex: 1 },
  summaryActions: { flexDirection: "row", gap: space[3], marginHorizontal: space[5] },
  summaryItem: { alignItems: "flex-start", flexDirection: "row", gap: space[2] },
  summaryLabel: { fontFamily: typography.family.bold },
  summaryList: { gap: space[2] },
  summaryText: { flex: 1 },
  navigationButton: {
    alignItems: "center",
    borderRadius: radius.pill,
    borderWidth: 1,
    height: control.minHeight,
    justifyContent: "center",
    width: control.minHeight,
  },
  navigationTitle: {
    flexShrink: 1,
    marginHorizontal: space[5],
    textAlign: "center",
  },
  menuActions: { gap: space[5] },
  matchSeparator: { height: space[5] },
  match: { gap: space[3] },
  matchSummary: { alignItems: "center", flexDirection: "row", gap: space[2] },
  matchScore: {
    fontFamily: typography.family.bold,
    minWidth: 48,
    textAlign: "center",
  },
  scoreField: { flex: 1 },
  scoreFields: { flexDirection: "row", gap: space[3] },
  teamName: { flex: 1, textAlign: "center" },
  roundHeader: { paddingBottom: space[3], paddingHorizontal: space[5], paddingTop: space[5] },
  resultModalBackdrop: {
    ...StyleSheet.absoluteFill,
    alignItems: "center",
    justifyContent: "center",
    padding: space[5],
  },
  resultModal: {
    borderRadius: radius.card,
    borderWidth: 1,
    gap: space[5],
    maxWidth: 440,
    padding: space[5],
    width: "100%",
  },
  completionConfirmationActions: { gap: space[3] },
  completionConfirmationCopy: { gap: space[2] },
  completionChampion: { textAlign: "center" },
  completionSuccessActions: { gap: space[3] },
  completionSuccessContent: {
    alignItems: "center",
    gap: space[3],
  },
  completionTitle: { textAlign: "center" },
});
