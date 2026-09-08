import { SymbolView } from "expo-symbols";
import { useEffect, useRef, useState } from "react";
import {
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

export function BracketView({
  tournament,
  canManage,
  onMatchFocus,
  onEdit,
}: {
  tournament: PublicTournament;
  canManage: boolean;
  onMatchFocus: (matchView: View) => void;
  onEdit: (id: string) => void;
}) {
  const t = getTranslator();
  const { colors } = usePreferences();
  const { width } = useWindowDimensions();
  const overview = width >= columnWidth * 2 + space[5] * 2;
  const [selectedRound, setSelectedRound] = useState(1);
  const [focusedMatch, setFocusedMatch] = useState<string>();
  const [focusRevision, setFocusRevision] = useState(0);
  const scroll = useRef<ScrollView>(null);
  const matchViews = useRef(new Map<string, View>());
  const teams = new Map(tournament.teams.map((team) => [team.id, team.name]));
  const byId = new Map(tournament.matches.map((match) => [match.id, match]));
  const rounds = [...new Set(tournament.matches.map((match) => match.round))].sort((a, b) => a - b);
  const total = rounds.length;
  const navigate = (match: Match) => {
    setSelectedRound(match.round);
    setFocusedMatch(match.id);
    setFocusRevision((revision) => revision + 1);
    if (overview) scroll.current?.scrollTo({ x: (match.round - 1) * columnWidth, animated: true });
  };
  useEffect(() => {
    if (!focusedMatch) return;
    const frame = requestAnimationFrame(() => {
      const matchView = matchViews.current.get(focusedMatch);
      if (matchView) onMatchFocus(matchView);
    });
    return () => cancelAnimationFrame(frame);
  }, [focusRevision, focusedMatch, onMatchFocus, selectedRound]);
  const matchLabel = (match: Match) =>
    `${getBracketRoundLabel(match.round, total)} · ${t("bracket_match").replace("{number}", String(match.sequence))}`;
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
              {matchLabel(match)}
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
      <View style={styles.intro}>
        <Text variant="title">{t("bracket_title")}</Text>
        <Text color="secondary">{t("bracket_navigation_hint")}</Text>
      </View>
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
            onPress={() => {
              setSelectedRound(round);
              setFocusedMatch(undefined);
              scroll.current?.scrollTo({ x: (round - 1) * columnWidth, animated: true });
            }}
          />
        ))}
      </ScrollView>
      {overview ? (
        <ScrollView ref={scroll} horizontal>
          {rounds.map((round) => (
            <View key={round} style={styles.column}>
              <Text variant="title" style={styles.intro}>
                {getBracketRoundLabel(round, total)}
              </Text>
              {tournament.matches.filter((match) => match.round === round).map(renderMatch)}
            </View>
          ))}
        </ScrollView>
      ) : (
        <View style={styles.stack}>
          <Text variant="title" style={styles.intro}>
            {getBracketRoundLabel(selectedRound, total)}
          </Text>
          {tournament.matches.filter((match) => match.round === selectedRound).map(renderMatch)}
        </View>
      )}
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
  column: { width: columnWidth, gap: space[5], paddingBottom: space[5] },
});
