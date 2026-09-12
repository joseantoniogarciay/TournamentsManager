import { SymbolView } from "expo-symbols";
import { forwardRef, useEffect, useImperativeHandle, useRef, useState } from "react";
import {
  type LayoutChangeEvent,
  type NativeScrollEvent,
  type NativeSyntheticEvent,
  Platform,
  Pressable,
  ScrollView,
  StyleSheet,
  View,
  useWindowDimensions,
} from "react-native";
import {
  color,
  control,
  gradient,
  radius,
  space,
  typography,
} from "@tournaments-manager/design-tokens";
import { LinearGradient } from "expo-linear-gradient";
import type { Match, PublicTournament } from "@/api/generated/models";
import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { Button, Card, Text } from "@/shared/ui";
import { WebIcon } from "@/shared/ui/web-icon";

export function getBracketRoundLabel(round: number, total: number) {
  const t = getTranslator();
  const remaining = total - round;
  if (remaining === 0) return t("bracket_final");
  if (remaining === 1) return t("bracket_semifinal");
  if (remaining === 2) return t("bracket_quarterfinal");
  if (remaining === 3) return t("bracket_round_of_16");
  return t("bracket_round").replace("{number}", String(round));
}

const columnWidth = control.minHeight * 8;

export type BracketHorizontalMetrics = {
  contentWidth: number;
  viewportWidth: number;
};

export type BracketViewHandle = {
  scrollHorizontallyTo: (offset: number) => void;
};

export const BracketView = forwardRef<
  BracketViewHandle,
  {
    tournament: PublicTournament;
    canManage: boolean;
    onHorizontalMetricsChange: (metrics: BracketHorizontalMetrics) => void;
    onHorizontalScroll: (offset: number) => void;
    onMatchFocus: (matchView: View) => void;
    onEdit: (id: string) => void;
    onRoundChange: (round: number) => void;
    roundSelectionRevision: number;
    selectedRound: number;
  }
>(function BracketView(
  {
    tournament,
    canManage,
    onHorizontalMetricsChange,
    onHorizontalScroll,
    onMatchFocus,
    onEdit,
    onRoundChange,
    roundSelectionRevision,
    selectedRound,
  },
  ref,
) {
  const t = getTranslator();
  const { colors } = usePreferences();
  const { width } = useWindowDimensions();
  const overview = width >= columnWidth * 2 + space[5] * 2;
  const [focusedMatch, setFocusedMatch] = useState<string>();
  const [focusRevision, setFocusRevision] = useState(0);
  const [boardViewportWidth, setBoardViewportWidth] = useState<number>();
  const boardScroll = useRef<ScrollView>(null);
  const columnHeaders = useRef<ScrollView>(null);
  const horizontalOffsets = useRef({ board: 0, headers: 0 });
  const matchViews = useRef(new Map<string, View>());
  const teams = new Map(tournament.teams.map((team) => [team.id, team.name]));
  const byId = new Map(tournament.matches.map((match) => [match.id, match]));
  const rounds = [...new Set(tournament.matches.map((match) => match.round))].sort((a, b) => a - b);
  const total = rounds.length;
  const bracketWidth = total * columnWidth;
  const measureBoardViewport = (event: LayoutChangeEvent) => {
    setBoardViewportWidth(event.nativeEvent.layout.width);
  };
  const scrollHorizontallyTo = (nextOffset: number) => {
    horizontalOffsets.current.board = nextOffset;
    horizontalOffsets.current.headers = nextOffset;
    boardScroll.current?.scrollTo({ x: nextOffset, animated: false });
    columnHeaders.current?.scrollTo({ x: nextOffset, animated: false });
  };
  useImperativeHandle(ref, () => ({ scrollHorizontallyTo }));
  useEffect(() => {
    onHorizontalMetricsChange({
      contentWidth: overview ? bracketWidth : 0,
      viewportWidth: overview ? (boardViewportWidth ?? bracketWidth) : 0,
    });
  }, [boardViewportWidth, bracketWidth, onHorizontalMetricsChange, overview]);
  const syncHorizontalScroll = (
    source: "board" | "headers",
    event: NativeSyntheticEvent<NativeScrollEvent>,
  ) => {
    const nextOffset = event.nativeEvent.contentOffset.x;
    horizontalOffsets.current[source] = nextOffset;
    if (source !== "board" && Math.abs(horizontalOffsets.current.board - nextOffset) > 1) {
      horizontalOffsets.current.board = nextOffset;
      boardScroll.current?.scrollTo({ x: nextOffset, animated: false });
    }
    if (source !== "headers" && Math.abs(horizontalOffsets.current.headers - nextOffset) > 1) {
      horizontalOffsets.current.headers = nextOffset;
      columnHeaders.current?.scrollTo({ x: nextOffset, animated: false });
    }
    onHorizontalScroll(nextOffset);
  };
  const navigate = (match: Match) => {
    onRoundChange(match.round);
    setFocusedMatch(match.id);
    setFocusRevision((revision) => revision + 1);
    if (overview) {
      boardScroll.current?.scrollTo({ x: (match.round - 1) * columnWidth, animated: true });
    }
  };
  useEffect(() => {
    if (overview) {
      boardScroll.current?.scrollTo({
        x: (selectedRound - 1) * columnWidth,
        animated: true,
      });
    }
  }, [overview, selectedRound]);
  useEffect(() => setFocusedMatch(undefined), [roundSelectionRevision]);
  useEffect(() => {
    if (!focusedMatch) return;
    const frame = requestAnimationFrame(() => {
      const matchView = matchViews.current.get(focusedMatch);
      if (matchView) onMatchFocus(matchView);
    });
    return () => cancelAnimationFrame(frame);
  }, [focusRevision, focusedMatch, onMatchFocus, selectedRound]);
  const matchTitle = (match: Match) => {
    const roundLabel = getBracketRoundLabel(match.round, total);
    return match.round === total
      ? roundLabel
      : t("bracket_match").replace("{number}", String(match.sequence));
  };
  const matchLabel = (match: Match) =>
    match.round === total
      ? matchTitle(match)
      : `${getBracketRoundLabel(match.round, total)} · ${matchTitle(match)}`;
  const successor = (match: Match) =>
    tournament.matches.find(
      (next) => next.homeSourceMatchId === match.id || next.awaySourceMatchId === match.id,
    );
  const hasPlayedDescendant = (match: Match): boolean => {
    let next = successor(match);
    while (next) {
      if (next.state === "completed") return true;
      next = successor(next);
    }
    return false;
  };
  const renderMatch = (match: Match) => {
    const next = successor(match);
    const locked = hasPlayedDescendant(match);
    const selected = focusedMatch === match.id;
    return (
      <View
        collapsable={false}
        key={match.id}
        ref={(view) => {
          if (view) matchViews.current.set(match.id, view);
          else matchViews.current.delete(match.id);
        }}
        style={[styles.matchFrame, selected ? styles.selectedMatchFrame : undefined]}
      >
        {selected ? (
          <LinearGradient
            {...gradient.brand}
            pointerEvents="none"
            style={StyleSheet.absoluteFill}
          />
        ) : null}
        <Card style={[styles.matchCard, selected ? styles.selectedMatchCard : undefined]}>
          <View style={styles.stack}>
            <Text style={styles.matchTitle} variant="bodyLarge">
              {matchTitle(match)}
            </Text>
            {(["home", "away"] as const).map((side) => {
              const teamId = side === "home" ? match.homeTeamId : match.awayTeamId;
              const kind = side === "home" ? match.homeSourceKind : match.awaySourceKind;
              const sourceId = side === "home" ? match.homeSourceMatchId : match.awaySourceMatchId;
              const source = sourceId ? byId.get(sourceId) : undefined;
              const score = side === "home" ? match.homeScore : match.awayScore;
              const penalty = side === "home" ? match.homePenalties : match.awayPenalties;
              const name =
                teams.get(teamId) ??
                (kind === "bye" ? t("bracket_bye") : t("bracket_awaiting_winner"));
              return (
                <View key={side} style={styles.slot}>
                  <View style={styles.teamRow}>
                    {source ? (
                      <Pressable
                        accessibilityLabel={t("bracket_go_to_source").replace(
                          "{match}",
                          matchLabel(source),
                        )}
                        accessibilityRole="button"
                        onPress={() => navigate(source)}
                        style={styles.sourceLink}
                      >
                        <BracketNavigationIcon color={colors.text.secondary} direction="back" />
                      </Pressable>
                    ) : null}
                    <View style={styles.teamIdentity}>
                      <Text style={styles.teamName} variant="bodyLarge">
                        {name}
                      </Text>
                      {match.winnerTeamId && match.winnerTeamId === teamId ? (
                        <WinnerCrown accessibilityLabel={t("bracket_winner")} />
                      ) : null}
                    </View>
                    {score !== undefined ? (
                      <Text variant="title">
                        {score}
                        {penalty !== undefined ? ` (${penalty})` : ""}
                      </Text>
                    ) : null}
                  </View>
                </View>
              );
            })}
            {match.homePenalties !== undefined ? (
              <Text color="secondary">{t("bracket_penalties_caption")}</Text>
            ) : null}
            {match.state === "bye" ? (
              <Text color="secondary">{t("bracket_auto_advance")}</Text>
            ) : null}
            {canManage &&
            tournament.state === "in_progress" &&
            match.homeTeamId &&
            match.awayTeamId &&
            match.state !== "bye" ? (
              locked ? (
                <Text color="secondary">{t("bracket_result_locked")}</Text>
              ) : (
                <Button
                  label={t(
                    match.state === "completed" ? "league_result_edit" : "league_result_add",
                  )}
                  variant={match.state === "completed" ? "secondary" : "primary"}
                  onPress={() => onEdit(match.id)}
                />
              )
            ) : null}
            {next ? (
              <Pressable
                accessibilityLabel={t("bracket_go_to_destination").replace(
                  "{match}",
                  matchLabel(next),
                )}
                accessibilityRole="button"
                onPress={() => navigate(next)}
                style={styles.nextLink}
              >
                <Text color="secondary" style={styles.nextLinkLabel}>
                  {matchLabel(next)}
                </Text>
                <BracketNavigationIcon color={colors.text.secondary} direction="forward" />
              </Pressable>
            ) : null}
          </View>
        </Card>
      </View>
    );
  };
  return (
    <View style={styles.stack}>
      {overview ? (
        <>
          {Platform.OS === "web" ? (
            <View
              style={[
                styles.stickyColumnHeaders,
                {
                  backgroundColor: colors.surface.canvas,
                  borderColor: colors.border.default,
                },
              ]}
            >
              <ScrollView
                ref={columnHeaders}
                horizontal
                onScroll={(event) => syncHorizontalScroll("headers", event)}
                scrollEventThrottle={16}
                showsHorizontalScrollIndicator={false}
              >
                {rounds.map((round) => (
                  <View key={round} style={styles.columnHeader}>
                    <Text variant="title">{getBracketRoundLabel(round, total)}</Text>
                  </View>
                ))}
              </ScrollView>
            </View>
          ) : null}
          <ScrollView
            ref={boardScroll}
            horizontal
            onLayout={measureBoardViewport}
            onScroll={(event) => syncHorizontalScroll("board", event)}
            scrollEventThrottle={16}
            showsHorizontalScrollIndicator={Platform.OS !== "web"}
          >
            {rounds.map((round) => (
              <View key={round} style={styles.column}>
                {tournament.matches.filter((match) => match.round === round).map(renderMatch)}
              </View>
            ))}
          </ScrollView>
        </>
      ) : (
        <View style={styles.stack}>
          {tournament.matches.filter((match) => match.round === selectedRound).map(renderMatch)}
        </View>
      )}
    </View>
  );
});

export function BracketIntro() {
  const t = getTranslator();
  const { width } = useWindowDimensions();
  const showsOverview = Platform.OS === "web" && width >= columnWidth * 2 + space[5] * 2;
  return (
    <View style={styles.intro}>
      <Text variant="title">{t("bracket_title")}</Text>
      <Text color="secondary">
        {t(showsOverview ? "bracket_overview_navigation_hint" : "bracket_navigation_hint")}
      </Text>
    </View>
  );
}

export function BracketRoundNavigation({
  onSelect,
  rounds,
  selectedRound,
}: {
  onSelect: (round: number) => void;
  rounds: number[];
  selectedRound: number;
}) {
  const { colors } = usePreferences();
  const { width } = useWindowDimensions();
  const total = rounds.length;
  if (Platform.OS === "web" && width >= columnWidth * 2 + space[5] * 2) return null;
  return (
    <View style={[styles.stickyRounds, { backgroundColor: colors.surface.canvas }]}>
      <ScrollView
        horizontal
        showsHorizontalScrollIndicator={false}
        contentContainerStyle={styles.rounds}
      >
        {rounds.map((round) => (
          <Button
            key={round}
            label={getBracketRoundLabel(round, total)}
            variant={selectedRound === round ? "primary" : "secondary"}
            onPress={() => onSelect(round)}
          />
        ))}
      </ScrollView>
    </View>
  );
}

function WinnerCrown({ accessibilityLabel }: { accessibilityLabel: string }) {
  return (
    <LinearGradient
      {...gradient.brand}
      accessibilityLabel={accessibilityLabel}
      accessibilityRole="image"
      accessible
      style={styles.winnerCrown}
    >
      <SymbolView
        name={{ android: "crown", ios: "crown.fill", web: "crown" }}
        size={14}
        tintColor={color.text.inverse}
      />
    </LinearGradient>
  );
}

function BracketNavigationIcon({
  color,
  direction,
}: {
  color: string;
  direction: "back" | "forward";
}) {
  if (Platform.OS === "web") {
    return (
      <WebIcon
        color={color}
        name={direction === "back" ? "back" : "chevronRight"}
        size={control.iconSize}
      />
    );
  }

  return (
    <SymbolView
      name={
        direction === "back"
          ? { android: "arrow_back", ios: "chevron.left", web: "arrow_back" }
          : { android: "arrow_forward", ios: "chevron.right", web: "arrow_forward" }
      }
      size={control.iconSize}
      tintColor={color}
    />
  );
}

const styles = StyleSheet.create({
  stack: { gap: space[5] },
  slot: { gap: space[2] },
  matchFrame: { borderRadius: radius.card, marginHorizontal: space[5] },
  selectedMatchFrame: {
    backgroundColor: color.brand.primary,
    overflow: "hidden",
    padding: 1,
  },
  matchCard: { marginHorizontal: 0 },
  selectedMatchCard: { borderRadius: radius.card - 1, borderWidth: 0 },
  matchTitle: { fontFamily: typography.family.semibold },
  teamRow: { flexDirection: "row", alignItems: "center", gap: space[3] },
  teamIdentity: { alignItems: "center", flex: 1, flexDirection: "row", gap: space[1] },
  teamName: { flexShrink: 1 },
  winnerCrown: {
    alignItems: "center",
    backgroundColor: color.brand.primary,
    borderRadius: radius.pill,
    height: 22,
    justifyContent: "center",
    overflow: "hidden",
    width: 22,
  },
  sourceLink: {
    alignItems: "center",
    height: control.minHeight,
    justifyContent: "center",
    width: control.minHeight,
  },
  nextLink: {
    alignItems: "center",
    alignSelf: "flex-end",
    flexDirection: "row",
    gap: space[1],
    justifyContent: "center",
    minHeight: control.minHeight,
  },
  nextLinkLabel: { fontFamily: typography.family.semibold },
  intro: { paddingHorizontal: space[5], gap: space[2] },
  rounds: { paddingHorizontal: space[5], gap: space[2] },
  stickyRounds: Platform.select({
    default: { paddingVertical: space[3], zIndex: 1 },
    web: { paddingVertical: space[3], position: "sticky", top: 0, zIndex: 1 },
  }),
  column: { width: columnWidth, gap: space[5], paddingBottom: space[5] },
  columnHeader: {
    justifyContent: "center",
    minHeight: control.minHeight,
    paddingHorizontal: space[5],
    width: columnWidth,
  },
  stickyColumnHeaders: {
    borderBottomWidth: StyleSheet.hairlineWidth,
    position: "sticky",
    top: 0,
    zIndex: 1,
  },
});
