import { router, Stack, useLocalSearchParams } from "expo-router";
import { randomUUID } from "expo-crypto";
import { useCallback, useEffect, useRef, useState } from "react";
import { Platform, StyleSheet, View } from "react-native";
import { space } from "@tournaments-manager/design-tokens";
import type { PublicTournament } from "@/api/generated/models";
import {
  getTournament,
  getTournamentRelationship,
  TournamentUnavailableError,
} from "@/features/league-creation/api";
import {
  maximumTeamNameLength,
  maximumTournamentNameLength,
} from "@/features/league-creation/draft";
import { useTournamentStore } from "@/features/league-creation/league-store";
import { invalidateTournamentLibrary } from "@/features/league-creation/tournament-library-revision";
import { useSession } from "@/shared/session/session-provider";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import {
  Button,
  Card,
  ConfigurationOption,
  KeyboardAwareScrollView,
  LoadingTransition,
  NavigationHeaderButton,
  RequestErrorCard,
  Screen,
  Text,
  TextField,
  usesLiquidGlassNavigation,
} from "@/shared/ui";
import { repeatTournamentRequest, RepetitionRejectedError } from "./api";
import { canRepeatTournament, repetitionConfiguration, reviewedRepetitionInput } from "./values";

type ReviewedTeam = { key: string; name: string };
export function RepetitionScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const t = getTranslator();
  const { user, rememberLastTeamName } = useSession();
  const { colors } = usePreferences();
  const { show } = useFeedback();
  const { putTournament } = useTournamentStore();
  const [source, setSource] = useState<PublicTournament>();
  const [name, setName] = useState("");
  const [teams, setTeams] = useState<ReviewedTeam[]>([]);
  const [ownKey, setOwnKey] = useState("");
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<string>();
  const [submitted, setSubmitted] = useState(false);
  const [saving, setSaving] = useState(false);
  const [rejected, setRejected] = useState(false);
  const requestID = useRef(randomUUID());
  const active = useRef(true);
  const pending = useRef<AbortController | null>(null);
  const generation = useRef(0);
  const load = useCallback(async () => {
    const revision = ++generation.current;
    setLoading(true);
    setLoadError(undefined);
    try {
      const [value, relationship] = await Promise.all([
        getTournament(id),
        user ? getTournamentRelationship(id) : Promise.resolve(null),
      ]);
      if (!active.current || revision !== generation.current) return;
      if (!canRepeatTournament(relationship, value.state))
        throw new RepetitionRejectedError(
          relationship === "organizer" ? "not_completed" : "forbidden",
        );
      setSource(value);
      setName(value.name);
      const copied = value.teams.map((team) => ({ key: team.id, name: team.name }));
      setTeams(copied);
      setOwnKey(
        copied.find((team) => team.name === user?.lastTeamName)?.key ?? copied[0]?.key ?? "",
      );
      setRejected(false);
      setSubmitted(false);
    } catch (error) {
      if (!active.current || revision !== generation.current) return;
      setLoadError(
        t(
          error instanceof RepetitionRejectedError
            ? `tournament_repeat_${error.reason}`
            : error instanceof TournamentUnavailableError
              ? "tournament_repeat_unavailable"
              : getRequestFailure(error).messageKey,
        ),
      );
    } finally {
      if (active.current && revision === generation.current) setLoading(false);
    }
  }, [id, t, user?.id]);
  useEffect(() => {
    active.current = true;
    void load();
    return () => {
      active.current = false;
      generation.current++;
      pending.current?.abort();
    };
  }, [load]);
  const close = () => {
    if (Platform.OS !== "web" && router.canDismiss()) router.dismiss();
    else router.replace(id ? `/tournament/${id}` : "/");
  };
  const teamError = (team: ReviewedTeam) => {
    const value = team.name.trim();
    if (!value) return t("league_team_required");
    if (Array.from(value).length > maximumTeamNameLength)
      return t("tournament_repeat_team_invalid");
    if (
      teams.some(
        (other) =>
          other.key !== team.key && other.name.trim().toLowerCase() === value.toLowerCase(),
      )
    )
      return t("tournament_repeat_team_duplicate");
    return undefined;
  };
  const nameError = !name.trim()
    ? t("league_name_required")
    : Array.from(name.trim()).length > maximumTournamentNameLength
      ? t("league_name_too_long")
      : undefined;
  const save = async () => {
    setSubmitted(true);
    const input = reviewedRepetitionInput(requestID.current, name, teams, ownKey);
    if (!source || !input || pending.current || rejected) return;
    const request = new AbortController();
    pending.current = request;
    setSaving(true);
    try {
      const value = await repeatTournamentRequest(id, input, request.signal);
      if (!active.current || request.signal.aborted) return;
      putTournament(value);
      invalidateTournamentLibrary();
      await rememberLastTeamName(input.teams[0].name).catch(() => undefined);
      if (!active.current || request.signal.aborted) return;
      router.replace(`/tournament/${value.id}`);
    } catch (error) {
      if (!active.current || request.signal.aborted) return;
      if (error instanceof RepetitionRejectedError) setRejected(true);
      const failure = getRequestFailure(error);
      show({
        kind: failure.kind,
        message: t(
          error instanceof RepetitionRejectedError
            ? `tournament_repeat_${error.reason}`
            : failure.messageKey,
        ),
      });
    } finally {
      if (pending.current === request) pending.current = null;
      if (active.current && !request.signal.aborted) setSaving(false);
    }
  };
  const title = t("tournament_repeat");
  const config = source ? repetitionConfiguration(source) : undefined;
  return (
    <>
      <Stack.Screen
        options={{
          headerBackVisible: false,
          headerShadowVisible: false,
          headerStyle: { backgroundColor: colors.surface.canvas },
          headerTintColor: colors.text.primary,
          headerTitleAlign: "center",
          title,
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
      <Screen bottomInset="none" topInset="navigation-bar" navigationTitle={title}>
        {loading ? (
          <LoadingTransition active message={t("common_loading")} />
        ) : loadError ? (
          <RequestErrorCard
            actionLabel={t("common_retry")}
            message={loadError}
            onRetry={() => void load()}
          />
        ) : source ? (
          <KeyboardAwareScrollView bottomInset="safe-area" contentContainerStyle={styles.content}>
            <Card>
              <View style={styles.form}>
                <Text variant="title">{t("tournament_repeat_new_edition")}</Text>
                <Text color="secondary">{t("tournament_repeat_description")}</Text>
                <TextField
                  label={t("league_name_label")}
                  value={name}
                  onChangeText={setName}
                  maxLength={maximumTournamentNameLength}
                  error={nameError}
                  validationSubmitted={submitted}
                  validationTrigger="blur"
                  editable={!saving && !rejected}
                />
                <Text variant="bodyLarge">{t("tournament_repeat_configuration")}</Text>
                <Text>{t(`tournament_sport_${source.sport}`)}</Text>
                <Text>
                  {t(
                    config?.format === "single_elimination"
                      ? "tournament_format_bracket"
                      : config?.format === "league_then_single_elimination"
                        ? "tournament_format_mixed"
                        : "tournament_format_league",
                  )}
                </Text>
                {config?.roundRobinLegs ? (
                  <Text>
                    {t(
                      config.roundRobinLegs === 2
                        ? "league_start_two_legs"
                        : "league_start_one_leg",
                    )}
                  </Text>
                ) : null}
                {source.bestOfSets ? (
                  <Text>
                    {t("tournament_repeat_best_of").replace("{count}", String(source.bestOfSets))}
                  </Text>
                ) : null}
                {source.pointsPerGame ? (
                  <Text>
                    {t("tournament_repeat_points").replace("{count}", String(source.pointsPerGame))}
                  </Text>
                ) : null}
                {config?.leagueStructure ? (
                  <Text>
                    {t(
                      config.leagueStructure === "groups"
                        ? "tournament_repeat_groups"
                        : "tournament_repeat_qualifiers",
                    )
                      .replace("{groups}", String(config.groupCount))
                      .replace(
                        "{count}",
                        String(config.qualifiersPerGroup ?? config.qualifierCount),
                      )}
                  </Text>
                ) : null}
                <Text color="secondary">{t("tournament_repeat_configuration_help")}</Text>
              </View>
            </Card>
            <Card>
              <View style={styles.form}>
                <Text variant="title">{t("tournament_repeat_teams")}</Text>
                <Text color="secondary">{t("tournament_repeat_teams_help")}</Text>
                {teams.map((team, index) => (
                  <View key={team.key} style={styles.form}>
                    <TextField
                      label={t("tournament_repeat_team_number").replace(
                        "{count}",
                        String(index + 1),
                      )}
                      value={team.name}
                      maxLength={maximumTeamNameLength}
                      onChangeText={(value) =>
                        setTeams((current) =>
                          current.map((item) =>
                            item.key === team.key ? { ...item, name: value } : item,
                          ),
                        )
                      }
                      error={teamError(team)}
                      validationSubmitted={submitted}
                      validationTrigger="blur"
                      editable={!saving && !rejected}
                    />
                    <View style={styles.options}>
                      <ConfigurationOption
                        disabled={saving || rejected}
                        label={t("tournament_repeat_own_team")}
                        selected={ownKey === team.key}
                        onPress={() => {
                          if (!saving && !rejected) setOwnKey(team.key);
                        }}
                      />
                      {teams.length > 1 ? (
                        <Button
                          label={t("league_remove_team")}
                          variant="secondary"
                          disabled={saving || rejected}
                          onPress={() => {
                            const remaining = teams.filter((item) => item.key !== team.key);
                            setTeams(remaining);
                            if (ownKey === team.key) setOwnKey(remaining[0].key);
                          }}
                        />
                      ) : null}
                    </View>
                  </View>
                ))}
                {teams.length < 64 ? (
                  <Button
                    label={t("league_add_team")}
                    variant="secondary"
                    disabled={saving || rejected}
                    onPress={() =>
                      setTeams((current) => [...current, { key: randomUUID(), name: "" }])
                    }
                  />
                ) : null}
                <Button
                  label={t("tournament_repeat_create")}
                  loading={saving}
                  disabled={rejected}
                  onPress={() => void save()}
                />
              </View>
            </Card>
          </KeyboardAwareScrollView>
        ) : null}
      </Screen>
    </>
  );
}
const styles = StyleSheet.create({
  content: { gap: space[5] },
  form: { gap: space[4] },
  options: { flexDirection: "row", flexWrap: "wrap", gap: space[2] },
});
