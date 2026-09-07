import { useRef, useState } from "react";
import { ScrollView, StyleSheet, View, useWindowDimensions } from "react-native";
import { control, space } from "@tournaments-manager/design-tokens";
import type { Match, PublicTournament } from "@/api/generated/models";
import { getTranslator } from "@/shared/i18n/locale";
import { Button, Card, Text } from "@/shared/ui";

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
  onEdit,
}: {
  tournament: PublicTournament;
  canManage: boolean;
  onEdit: (id: string) => void;
}) {
  const t = getTranslator();
  const { width } = useWindowDimensions();
  const overview = width >= columnWidth * 2 + space[5] * 2;
  const [selectedRound, setSelectedRound] = useState(1);
  const [focusedMatch, setFocusedMatch] = useState<string>();
  const scroll = useRef<ScrollView>(null);
  const teams = new Map(tournament.teams.map((team) => [team.id, team.name]));
  const byId = new Map(tournament.matches.map((match) => [match.id, match]));
  const rounds = [...new Set(tournament.matches.map((match) => match.round))].sort((a, b) => a - b);
  const total = rounds.length;
  const navigate = (match: Match) => {
    setSelectedRound(match.round);
    setFocusedMatch(match.id);
    if (overview) scroll.current?.scrollTo({ x: (match.round - 1) * columnWidth, animated: true });
  };
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
    return (
      <Card key={match.id}>
        <View style={styles.stack}>
          <Text variant="bodyLarge">
            {matchLabel(match)}
            {focusedMatch === match.id ? ` · ${t("bracket_selected")}` : ""}
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
                  <Text style={styles.team} variant="bodyLarge">
                    {name}
                    {match.winnerTeamId && match.winnerTeamId === teamId
                      ? ` · ${t("bracket_winner")}`
                      : ""}
                  </Text>
                  {score !== undefined ? (
                    <Text variant="title">
                      {score}
                      {penalty !== undefined ? ` (${penalty})` : ""}
                    </Text>
                  ) : null}
                </View>
                {source ? (
                  <Button
                    variant="ghost"
                    label={t("bracket_winner_of").replace("{match}", matchLabel(source))}
                    onPress={() => navigate(source)}
                  />
                ) : null}
              </View>
            );
          })}
          {match.homePenalties !== undefined ? (
            <Text color="secondary">{t("bracket_penalties_caption")}</Text>
          ) : null}
          {match.state === "bye" ? (
            <Text color="secondary">{t("bracket_auto_advance")}</Text>
          ) : null}
          {next ? (
            <Button
              variant="secondary"
              label={t("bracket_advances_to").replace("{match}", matchLabel(next))}
              onPress={() => navigate(next)}
            />
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
                label={t(match.state === "completed" ? "league_result_edit" : "league_result_add")}
                variant={match.state === "completed" ? "secondary" : "primary"}
                onPress={() => onEdit(match.id)}
              />
            )
          ) : null}
        </View>
      </Card>
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

const styles = StyleSheet.create({
  stack: { gap: space[5] },
  slot: { gap: space[2] },
  teamRow: { flexDirection: "row", alignItems: "center", gap: space[3] },
  team: { flex: 1 },
  intro: { paddingHorizontal: space[5], gap: space[2] },
  rounds: { paddingHorizontal: space[5], gap: space[2] },
  column: { width: columnWidth, gap: space[5], paddingBottom: space[5] },
});
