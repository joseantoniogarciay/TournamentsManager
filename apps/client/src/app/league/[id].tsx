import { router, Stack, useLocalSearchParams } from "expo-router";
import { SymbolView } from "expo-symbols";
import { useEffect, useState } from "react";
import { Modal, Platform, Pressable, SectionList, Share, StyleSheet, View } from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";

import { color, control, radius, space, typography } from "@tournaments-manager/design-tokens";

import { APIUnexpectedResponseError } from "@/api/fetch";
import type { PublicLeague } from "@/api/generated/models";
import {
  cancelLeagueRequest,
  completeLeagueRequest,
  getLeagueRelationship,
  startLeagueRequest,
} from "@/features/league-creation/api";
import { useLeague, useLeagueStore } from "@/features/league-creation/league-store";
import { recordMatchResultRequest } from "@/features/match-results/api";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import { getLeagueStateLabel } from "@/shared/i18n/league-state";
import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { useSession } from "@/shared/session/session-provider";
import {
  Button,
  Card,
  LoadingTransition,
  Screen,
  Text,
  TextField,
  useConfirmationDialog,
} from "@/shared/ui";
import { WebIcon } from "@/shared/ui/web-icon";

const localAppLinkURL = "http://localhost:8081";

export default function LeagueScreen() {
  const t = getTranslator();
  const { id } = useLocalSearchParams<{ id: string }>();
  const { user } = useSession();
  const { colors } = usePreferences();
  const insets = useSafeAreaInsets();
  const { show } = useFeedback();
  const { confirm } = useConfirmationDialog();
  const league = useLeague(id);
  const { loadLeague, putLeague, refreshLeague } = useLeagueStore();
  const [relationship, setRelationship] = useState<string>();
  const [isStarting, setIsStarting] = useState(false);
  const [roundRobinLegs, setRoundRobinLegs] = useState<1 | 2>(1);
  const [menuOpen, setMenuOpen] = useState(false);
  const [scores, setScores] = useState<Record<string, { home: string; away: string }>>({});
  const [savingMatchID, setSavingMatchID] = useState<string>();
  const [editingMatchID, setEditingMatchID] = useState<string>();
  const [isCompleting, setIsCompleting] = useState(false);
  const [isCancelling, setIsCancelling] = useState(false);
  const [completionConfirmationOpen, setCompletionConfirmationOpen] = useState(false);
  const [completionOpen, setCompletionOpen] = useState(false);
  useEffect(() => {
    if (!id) return;
    void loadLeague(id).catch((error) => {
      const failure = getRequestFailure(error);
      show({ kind: failure.kind, message: t(failure.messageKey) });
    });
    if (user) void getLeagueRelationship(id).then(setRelationship);
  }, [id, loadLeague, show, t, user]);
  const isOrganizer = relationship === "organizer";
  const canManageResults = relationship === "organizer" || relationship === "delegated";
  const start = async () => {
    if (!id) return;
    setIsStarting(true);
    try {
      putLeague(await startLeagueRequest(id, { roundRobinLegs }));
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
      message: `${league.name}: ${base}/league/${id}`,
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
        void cancelLeagueRequest(id)
          .then(putLeague)
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
    void completeLeagueRequest(id)
      .then((completed) => {
        putLeague(completed);
        setCompletionConfirmationOpen(false);
        setCompletionOpen(true);
      })
      .catch((error) => {
        if (error instanceof APIUnexpectedResponseError && error.status === 409) {
          setCompletionConfirmationOpen(false);
          show({ kind: "success", message: t("league_completion_already_completed") });
          void refreshLeague(id).catch((refreshError) => {
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
    if (!id) return;
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
      putLeague(await recordMatchResultRequest(id, matchID, { homeScore, awayScore }));
      setEditingMatchID(undefined);
    } catch (error) {
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
  if (!league)
    return (
      <Screen topInset="navigation-bar">
        <Card>
          <Text>{t("common_loading")}</Text>
        </Card>
      </Screen>
    );
  const canCancel = league.state === "published" || league.state === "in_progress";
  const canComplete =
    isOrganizer &&
    league.state === "in_progress" &&
    league.matches.length > 0 &&
    league.matches.every((match) => match.state === "completed");
  const primaryLeagueAction = (() => {
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
      })
    : undefined;
  const canSaveResult =
    editingScore !== undefined &&
    /^\d+$/.test(editingScore.home) &&
    /^\d+$/.test(editingScore.away);
  const openResultEditor = (matchID: string) => setEditingMatchID(matchID);
  const matchesByRound = new Map<number, PublicLeague["matches"]>();
  for (const match of league.matches) {
    const matches = matchesByRound.get(match.round) ?? [];
    matches.push(match);
    matchesByRound.set(match.round, matches);
  }
  const matchSections = [...matchesByRound.entries()]
    .sort(([firstRound], [secondRound]) => firstRound - secondRound)
    .map(([round, data]) => ({ data, round }));
  const closeWebMenu = () => setMenuOpen(false);
  const openAdministrators = () => router.push(`/league/${league.id}/administrators`);
  const headerOptions = {
    headerBackVisible: false,
    headerShadowVisible: false,
    headerStyle: { backgroundColor: colors.surface.canvas },
    headerTintColor: colors.text.primary,
    headerTitle: () => (
      <Text numberOfLines={2} style={styles.navigationTitle} variant="title">
        {league.name}
      </Text>
    ),
  };
  return (
    <>
      <Stack.Screen options={headerOptions} />
      {Platform.OS !== "ios" ? (
        <Stack.Screen
          options={{
            headerLeft: () => (
              <Pressable
                accessibilityLabel={t("common_back")}
                accessibilityRole="button"
                onPress={returnToPreviousScreen}
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
            ),
            headerRight: () => (
              <Pressable
                accessibilityLabel={t("league_actions")}
                accessibilityRole="button"
                onPress={() => setMenuOpen((open) => !open)}
                style={[
                  styles.navigationButton,
                  { backgroundColor: colors.surface.default, borderColor: colors.border.default },
                ]}
              >
                {Platform.OS === "web" ? (
                  <WebIcon color={colors.text.primary} name="more" size={control.iconSize} />
                ) : (
                  <SymbolView
                    name={{ android: "more_vert", ios: "ellipsis", web: "more_vert" }}
                    size={control.iconSize}
                    tintColor={colors.text.primary}
                  />
                )}
              </Pressable>
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
            <Stack.Toolbar.Menu accessibilityLabel={t("league_actions")} icon="ellipsis">
              <Stack.Toolbar.MenuAction onPress={() => void share()}>
                {t("league_share")}
              </Stack.Toolbar.MenuAction>
              {isOrganizer ? (
                <Stack.Toolbar.MenuAction onPress={openAdministrators}>
                  {t("league_administrators")}
                </Stack.Toolbar.MenuAction>
              ) : null}
              {isOrganizer && canCancel && !isCancelling ? (
                <Stack.Toolbar.MenuAction destructive onPress={cancel}>
                  {t("league_cancel")}
                </Stack.Toolbar.MenuAction>
              ) : null}
            </Stack.Toolbar.Menu>
          </Stack.Toolbar>
        </>
      )}
      <Screen bottomInset="none" topInset="navigation-bar">
        <SectionList
          contentContainerStyle={[
            styles.content,
            {
              paddingBottom:
                insets.bottom + (primaryLeagueAction ? space[12] + space[5] : space[4]),
            },
          ]}
          ItemSeparatorComponent={() => <View style={styles.matchSeparator} />}
          sections={matchSections}
          showsVerticalScrollIndicator={false}
          stickySectionHeadersEnabled
          ListHeaderComponent={
            <View style={styles.listHeader}>
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
                        {getLeagueStateLabel(t, league.state)}
                      </Text>
                    </View>
                  </View>
                </View>
              </Card>
              <View style={styles.summaryActions}>
                <View style={styles.summaryAction}>
                  <Button
                    label={t("league_teams")}
                    onPress={() => router.push(`/league/${league.id}/teams`)}
                    variant="secondary"
                  />
                </View>
                <View style={styles.summaryAction}>
                  <Button
                    label={t("league_standings")}
                    onPress={() => router.push(`/league/${league.id}/standings`)}
                    variant="secondary"
                  />
                </View>
              </View>
              {Platform.OS !== "ios" && menuOpen ? (
                <View
                  accessibilityViewIsModal
                  style={[
                    styles.menu,
                    { backgroundColor: colors.surface.default, borderColor: colors.border.default },
                  ]}
                >
                  <Button
                    label={t("league_share")}
                    onPress={() => {
                      closeWebMenu();
                      void share();
                    }}
                    variant="ghost"
                  />
                  {isOrganizer ? (
                    <>
                      <Button
                        label={t("league_administrators")}
                        onPress={() => {
                          closeWebMenu();
                          openAdministrators();
                        }}
                        variant="ghost"
                      />
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
              ) : null}
              {league.state === "published" && isOrganizer ? (
                <Card>
                  <View style={styles.stack}>
                    <Text style={styles.configurationTitle} variant="bodyLarge">
                      {t("league_start_title")}
                    </Text>
                    <View
                      style={[styles.configurationOptions, { borderColor: colors.border.default }]}
                    >
                      <Pressable
                        accessibilityRole="button"
                        accessibilityState={{ selected: roundRobinLegs === 1 }}
                        disabled={isStarting}
                        onPress={() => setRoundRobinLegs(1)}
                        style={[
                          styles.configurationChip,
                          roundRobinLegs === 1
                            ? {
                                backgroundColor: color.brand.primary,
                                borderColor: color.brand.primary,
                              }
                            : {
                                backgroundColor: colors.surface.default,
                                borderColor: colors.border.default,
                              },
                        ]}
                      >
                        <Text color={roundRobinLegs === 1 ? "onBrand" : "primary"}>
                          {t("league_start_one_leg")}
                        </Text>
                      </Pressable>
                      <Pressable
                        accessibilityRole="button"
                        accessibilityState={{ selected: roundRobinLegs === 2 }}
                        disabled={isStarting}
                        onPress={() => setRoundRobinLegs(2)}
                        style={[
                          styles.configurationChip,
                          roundRobinLegs === 2
                            ? {
                                backgroundColor: color.brand.primary,
                                borderColor: color.brand.primary,
                              }
                            : {
                                backgroundColor: colors.surface.default,
                                borderColor: colors.border.default,
                              },
                        ]}
                      >
                        <Text color={roundRobinLegs === 2 ? "onBrand" : "primary"}>
                          {t("league_start_two_legs")}
                        </Text>
                      </Pressable>
                    </View>
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
                    <Text style={styles.matchScore} variant="display">
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
        {primaryLeagueAction ? (
          <View style={[styles.floatingAction, { bottom: insets.bottom + space[3] }]}>
            <Button
              label={primaryLeagueAction.label}
              loading={primaryLeagueAction.loading}
              onPress={primaryLeagueAction.onPress}
            />
          </View>
        ) : null}
        <Modal
          animationType="fade"
          onRequestClose={() => {
            if (!isCompleting) setCompletionConfirmationOpen(false);
          }}
          transparent
          visible={completionConfirmationOpen}
        >
          <View style={styles.completionBackdrop}>
            <Pressable
              accessibilityLabel={t("common_cancel")}
              accessibilityRole="button"
              disabled={isCompleting}
              onPress={() => setCompletionConfirmationOpen(false)}
              style={StyleSheet.absoluteFill}
            />
            <View
              accessibilityViewIsModal
              style={[
                styles.completionModal,
                { backgroundColor: colors.surface.default, borderColor: colors.border.default },
              ]}
            >
              <Text style={styles.completionTitle} variant="title">
                {t("league_complete_title")}
              </Text>
              <Text color="secondary" style={styles.completionChampion}>
                {t("league_complete_description")}
              </Text>
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
          </View>
        </Modal>
        <Modal
          animationType="fade"
          onRequestClose={() => {
            if (!savingMatchID) setEditingMatchID(undefined);
          }}
          transparent
          visible={editingMatch !== undefined}
        >
          <View style={[styles.resultModalBackdrop, { backgroundColor: colors.surface.canvas }]}>
            <Pressable
              accessibilityLabel={t("common_close")}
              accessibilityRole="button"
              disabled={savingMatchID !== undefined}
              onPress={() => setEditingMatchID(undefined)}
              style={StyleSheet.absoluteFill}
            />
            {editingMatch && editingScore ? (
              <View
                accessibilityViewIsModal
                style={[
                  styles.resultModal,
                  { backgroundColor: colors.surface.default, borderColor: colors.border.default },
                ]}
              >
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
          </View>
        </Modal>
        <Modal
          animationType="fade"
          onRequestClose={() => setCompletionOpen(false)}
          transparent
          visible={completionOpen}
        >
          <View style={styles.completionBackdrop}>
            <Pressable
              accessibilityLabel={t("common_close")}
              accessibilityRole="button"
              onPress={() => setCompletionOpen(false)}
              style={StyleSheet.absoluteFill}
            />
            <View
              accessibilityViewIsModal
              style={[
                styles.completionModal,
                { backgroundColor: colors.surface.default, borderColor: colors.border.default },
              ]}
            >
              <SymbolView name="trophy.fill" size={56} tintColor={colors.feedback.success} />
              <Text style={styles.completionTitle} variant="display">
                {t("league_completion_title")}
              </Text>
              <Text color="secondary" style={styles.completionChampion}>
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
              <Button
                label={t("league_completion_view_standings")}
                onPress={() => {
                  setCompletionOpen(false);
                  router.push(`/league/${league.id}/standings`);
                }}
              />
              <Button
                label={t("common_close")}
                onPress={() => setCompletionOpen(false)}
                variant="secondary"
              />
            </View>
          </View>
        </Modal>
      </Screen>
    </>
  );
}
const styles = StyleSheet.create({
  content: { paddingBottom: space[4] },
  configurationChip: {
    alignItems: "center",
    borderRadius: radius.pill,
    borderWidth: 1,
    justifyContent: "center",
    minHeight: control.minHeight,
    paddingHorizontal: control.horizontalPadding,
  },
  configurationOptions: {
    borderTopWidth: 1,
    flexDirection: "row",
    flexWrap: "wrap",
    gap: space[2],
    paddingTop: space[5],
  },
  configurationTitle: { fontWeight: typography.weight.semibold },
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
  summaryLabel: { fontWeight: typography.weight.bold },
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
  navigationTitle: { maxWidth: 220, textAlign: "center" },
  menu: {
    borderRadius: radius.card,
    borderWidth: 1,
    gap: space[1],
    marginHorizontal: space[5],
    padding: space[2],
  },
  matchSeparator: { height: space[5] },
  match: { gap: space[3] },
  matchSummary: { alignItems: "center", flexDirection: "row", gap: space[2] },
  matchScore: { minWidth: 48, textAlign: "center" },
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
  completionBackdrop: {
    ...StyleSheet.absoluteFill,
    alignItems: "center",
    backgroundColor: "rgba(0, 0, 0, 0.48)",
    justifyContent: "center",
    padding: space[5],
  },
  completionChampion: { textAlign: "center" },
  completionModal: {
    alignItems: "center",
    borderRadius: radius.card,
    borderWidth: 1,
    gap: space[3],
    maxWidth: 440,
    padding: space[6],
    width: "100%",
  },
  completionTitle: { textAlign: "center" },
});
