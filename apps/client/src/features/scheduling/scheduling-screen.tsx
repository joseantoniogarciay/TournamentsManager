import { router, Stack, useLocalSearchParams } from "expo-router";
import { useCallback, useEffect, useRef, useState } from "react";
import { Platform, StyleSheet, View } from "react-native";
import { space } from "@tournaments-manager/design-tokens";
import {
  getTournamentRelationship,
  TournamentUnavailableError,
} from "@/features/league-creation/api";
import { useTournament, useTournamentStore } from "@/features/league-creation/league-store";
import { useSession } from "@/shared/session/session-provider";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import {
  Button,
  Card,
  KeyboardAwareScrollView,
  LoadingTransition,
  NavigationHeaderButton,
  RequestErrorCard,
  Screen,
  Text,
  TextField,
  usesLiquidGlassNavigation,
} from "@/shared/ui";
import {
  saveMatchSchedule,
  saveTournamentVenue,
  SchedulingConflictError,
  SchedulingForbiddenError,
  SchedulingUnavailableError,
} from "./api";
import {
  canEditScheduling,
  dateFromInput,
  dateToInput,
  validCivilTime,
  validTimeZone,
  validVenue,
} from "./values";

export function SchedulingScreen({ mode }: { mode: "match" | "venue" }) {
  const { id, matchId } = useLocalSearchParams<{ id: string; matchId?: string }>();
  const t = getTranslator();
  const { user } = useSession();
  const { colors } = usePreferences();
  const { show } = useFeedback();
  const tournament = useTournament(id);
  const { refreshTournament, putTournament } = useTournamentStore();
  const [relationship, setRelationship] = useState<string | null>();
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<string>();
  const [venue, setVenue] = useState("");
  const [date, setDate] = useState("");
  const [time, setTime] = useState("");
  const [zone, setZone] = useState("UTC");
  const [submitted, setSubmitted] = useState(false);
  const [saving, setSaving] = useState(false);
  const [rejected, setRejected] = useState(false);
  const active = useRef(true);
  const pending = useRef<AbortController | null>(null);
  const generation = useRef(0);
  const load = useCallback(async () => {
    const revision = ++generation.current;
    setLoading(true);
    setSaving(false);
    setLoadError(undefined);
    setSubmitted(false);
    setRejected(false);
    try {
      const [value, relation] = await Promise.all([
        refreshTournament(id),
        user ? getTournamentRelationship(id) : Promise.resolve(null),
      ]);
      if (!active.current || revision !== generation.current) return;
      const match =
        mode === "match" ? value.matches.find((item) => item.id === matchId) : undefined;
      if (mode === "match" && (!match || match.state === "bye"))
        throw new SchedulingUnavailableError();
      setRelationship(relation);
      setVenue(mode === "match" ? (match?.venue ?? "") : (value.venue ?? ""));
      setDate(dateToInput(match?.scheduledDate ?? ""));
      setTime(match?.scheduledTime ?? "");
      setZone(match?.timeZone ?? Intl.DateTimeFormat().resolvedOptions().timeZone);
    } catch (error) {
      if (!active.current || revision !== generation.current) return;
      setLoadError(
        t(
          error instanceof TournamentUnavailableError || error instanceof SchedulingUnavailableError
            ? "scheduling_unavailable"
            : getRequestFailure(error).messageKey,
        ),
      );
    } finally {
      if (active.current && revision === generation.current) setLoading(false);
    }
  }, [id, matchId, mode, refreshTournament, t, user]);
  useEffect(() => {
    active.current = true;
    void load();
    return () => {
      active.current = false;
      generation.current++;
      pending.current?.abort();
      pending.current = null;
    };
  }, [load]);
  const close = () => {
    if (Platform.OS !== "web" && router.canDismiss()) router.dismiss();
    else router.replace(id ? `/tournament/${id}` : "/");
  };
  const parsedDate = dateFromInput(date);
  const dateError = date && parsedDate === null ? t("scheduling_date_error") : undefined;
  const timeError = time && !validCivilTime(time) ? t("scheduling_time_error") : undefined;
  const venueError = !validVenue(venue.trim()) ? t("scheduling_venue_error") : undefined;
  const allowed = tournament && canEditScheduling(relationship, tournament.state);
  const save = async () => {
    setSubmitted(true);
    if (
      !allowed ||
      rejected ||
      (mode === "match" && !matchId) ||
      pending.current ||
      venueError ||
      (mode === "match" && (dateError || timeError || (time && !validTimeZone(zone))))
    )
      return;
    const request = new AbortController();
    pending.current = request;
    setSaving(true);
    try {
      const value =
        mode === "venue"
          ? await saveTournamentVenue(id, venue.trim(), request.signal)
          : await saveMatchSchedule(
              id,
              matchId ?? "",
              {
                scheduledDate: parsedDate ?? "",
                scheduledTime: time,
                timeZone: time ? zone : "",
                venue: venue.trim(),
              },
              request.signal,
            );
      if (!active.current || request.signal.aborted) return;
      putTournament(value);
      close();
    } catch (error) {
      if (!active.current || request.signal.aborted) return;
      if (
        error instanceof SchedulingForbiddenError ||
        error instanceof SchedulingConflictError ||
        error instanceof SchedulingUnavailableError
      )
        setRejected(true);
      show({
        message: t(
          error instanceof SchedulingForbiddenError
            ? "scheduling_forbidden"
            : error instanceof SchedulingConflictError
              ? "scheduling_conflict"
              : error instanceof SchedulingUnavailableError
                ? "scheduling_unavailable"
                : getRequestFailure(error).messageKey,
        ),
        kind: getRequestFailure(error).kind,
      });
    } finally {
      if (pending.current === request) pending.current = null;
      if (active.current && !request.signal.aborted) setSaving(false);
    }
  };
  const title = t(mode === "match" ? "scheduling_title" : "scheduling_tournament_venue");
  const match = tournament?.matches.find((item) => item.id === matchId);
  const teams = new Map(tournament?.teams.map((team) => [team.id, team.name]));
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
        ) : !allowed ? (
          <RequestErrorCard
            actionLabel={t("common_retry")}
            message={t("scheduling_forbidden")}
            onRetry={() => void load()}
          />
        ) : (
          <KeyboardAwareScrollView bottomInset="safe-area" contentContainerStyle={styles.content}>
            <Card>
              <Text variant="title">{tournament?.name}</Text>
              {match ? (
                <Text color="secondary">
                  {teams.get(match.homeTeamId) ?? t("bracket_awaiting_winner")} —{" "}
                  {teams.get(match.awayTeamId) ?? t("bracket_awaiting_winner")}
                </Text>
              ) : null}
            </Card>
            <Card>
              <View style={styles.fields}>
                {rejected ? (
                  <Button
                    label={t("common_retry")}
                    variant="secondary"
                    onPress={() => void load()}
                  />
                ) : null}
                {mode === "match" ? (
                  <>
                    <Text color="secondary">{t("scheduling_optional")}</Text>
                    <TextField
                      label={t("scheduling_date")}
                      placeholder={t("scheduling_date_placeholder")}
                      value={date}
                      onChangeText={setDate}
                      error={dateError}
                      validationTrigger="blur"
                      validationSubmitted={submitted}
                      editable={!saving && !rejected}
                      maxLength={10}
                      keyboardType="numbers-and-punctuation"
                    />
                    <TextField
                      label={t("scheduling_time")}
                      placeholder={t("scheduling_time_placeholder")}
                      value={time}
                      onChangeText={setTime}
                      error={timeError}
                      validationTrigger="blur"
                      validationSubmitted={submitted}
                      editable={!saving && !rejected}
                      maxLength={5}
                      keyboardType="numbers-and-punctuation"
                    />
                    <Text color="secondary">{t("scheduling_zone").replace("{zone}", zone)}</Text>
                  </>
                ) : null}
                <TextField
                  label={t("scheduling_venue")}
                  placeholder={t("scheduling_venue_placeholder")}
                  value={venue}
                  onChangeText={setVenue}
                  error={venueError}
                  validationTrigger="blur"
                  validationSubmitted={submitted}
                  editable={!saving && !rejected}
                />
                {mode === "venue" ? (
                  <>
                    <Text color="secondary">{t("scheduling_common_venue_help")}</Text>
                    <Text color="secondary">{t("scheduling_common_venue_optional")}</Text>
                  </>
                ) : tournament?.venue ? (
                  <Text color="secondary">
                    {t("scheduling_inherit").replace("{venue}", tournament.venue)}
                  </Text>
                ) : null}
                <Button
                  label={t("scheduling_save")}
                  loading={saving}
                  disabled={rejected}
                  onPress={() => void save()}
                />
                {mode === "match" && (date || time) ? (
                  <Button
                    label={t("scheduling_clear_when")}
                    variant="secondary"
                    disabled={saving || rejected}
                    onPress={() => {
                      setDate("");
                      setTime("");
                    }}
                  />
                ) : null}
              </View>
            </Card>
          </KeyboardAwareScrollView>
        )}
      </Screen>
    </>
  );
}
const styles = StyleSheet.create({ content: { gap: space[5] }, fields: { gap: space[4] } });
