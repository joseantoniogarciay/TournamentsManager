import { router, Stack } from "expo-router";
import { randomUUID } from "expo-crypto";
import { useEffect, useRef, useState } from "react";
import { StyleSheet, View, type TextInput } from "react-native";

import { space } from "@tournaments-manager/design-tokens";

import { createTournamentRequest } from "@/features/league-creation/api";
import {
  clearLocalTournamentDraft,
  getLocalTournamentDraft,
  maximumTournamentNameLength,
  maximumTournamentTeams,
  saveLocalTournamentDraft,
  type TournamentSport,
} from "@/features/league-creation/draft";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { useSession } from "@/shared/session/session-provider";
import {
  Button,
  Card,
  ConfigurationOption,
  KeyboardAwareScrollView,
  NavigationHeaderButton,
  Screen,
  Text,
  TextField,
  usesLiquidGlassNavigation,
} from "@/shared/ui";

export default function CreateTournamentScreen() {
  const t = getTranslator();
  const { show } = useFeedback();
  const { user } = useSession();
  const { colors } = usePreferences();
  const [name, setName] = useState("");
  const [sport, setSport] = useState<TournamentSport>("football");
  const [teams, setTeams] = useState(["", ""]);
  const [draftId, setDraftId] = useState(() => randomUUID());
  const [draftLoaded, setDraftLoaded] = useState(false);
  const [submitted, setSubmitted] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const teamInputRefs = useRef<Record<number, TextInput | null>>({});
  const [teamToFocus, setTeamToFocus] = useState<number>();

  useEffect(() => {
    void getLocalTournamentDraft().then((draft) => {
      if (draft) {
        setDraftId(draft.draftId);
        setName(draft.name);
        setSport(draft.sport);
        setTeams(draft.teams.length >= 2 ? draft.teams : ["", ""]);
      }
      setDraftLoaded(true);
    });
  }, []);
  useEffect(() => {
    if (draftLoaded) void saveLocalTournamentDraft({ draftId, name, sport, teams });
  }, [draftId, draftLoaded, name, sport, teams]);
  useEffect(() => {
    if (teamToFocus === undefined) return;
    teamInputRefs.current[teamToFocus]?.focus();
    setTeamToFocus(undefined);
  }, [teamToFocus, teams.length]);

  const normalizedTeamValues = teams.map((team) => team.trim());
  const normalizedTeams = normalizedTeamValues.filter(Boolean);
  const nameError = !name.trim()
    ? t("league_name_required")
    : name.length > maximumTournamentNameLength
      ? t("league_name_too_long")
      : undefined;
  const teamsError =
    normalizedTeams.length < 2 ||
    new Set(normalizedTeams.map((team) => team.toLowerCase())).size !== normalizedTeams.length
      ? t("league_teams_required")
      : undefined;
  const teamError = (index: number) => {
    if (submitted) return teamsError;
    const value = normalizedTeamValues[index]?.toLowerCase();
    if (!value) return undefined;
    return normalizedTeamValues.filter((team) => team.toLowerCase() === value).length > 1
      ? t("league_teams_required")
      : undefined;
  };
  const publish = async () => {
    setSubmitted(true);
    if (nameError || teamsError) return;
    if (!user) {
      router.push("/account-authentication" as never);
      return;
    }
    setIsSubmitting(true);
    try {
      const league = await createTournamentRequest({
        name: name.trim(),
        sport,
        teams: normalizedTeams.map((team) => ({ name: team })),
      });
      await clearLocalTournamentDraft();
      router.replace(`/tournament/${league.id}` as never);
    } catch (error) {
      const failure = getRequestFailure(error);
      show({ kind: failure.kind, message: t(failure.messageKey) });
    } finally {
      setIsSubmitting(false);
    }
  };
  const addTeam = () => {
    if (teams.length >= maximumTournamentTeams) {
      show({ kind: "generic-error", message: t("league_team_limit_reached") });
      return;
    }
    setTeamToFocus(teams.length);
    setTeams((current) => [...current, ""]);
  };
  const close = async () => {
    await clearLocalTournamentDraft();
    if (router.canDismiss()) {
      router.dismiss();
      return;
    }
    router.replace("/");
  };

  return (
    <>
      <Stack.Screen
        options={{
          headerBackVisible: false,
          headerShadowVisible: false,
          headerStyle: { backgroundColor: colors.surface.canvas },
          headerTintColor: colors.text.primary,
          headerTitle: t("league_create_title"),
          headerTitleAlign: "center",
        }}
      >
        {usesLiquidGlassNavigation ? (
          <Stack.Toolbar placement="left">
            <Stack.Toolbar.Button
              accessibilityLabel={t("common_close")}
              icon="xmark"
              onPress={() => void close()}
            />
          </Stack.Toolbar>
        ) : null}
      </Stack.Screen>
      {!usesLiquidGlassNavigation ? (
        <Stack.Screen
          options={{
            headerLeft: () => (
              <NavigationHeaderButton
                accessibilityLabel={t("common_close")}
                icon="close"
                nativeIcon={{ android: "close", ios: "xmark", web: "close" }}
                onPress={() => void close()}
              />
            ),
          }}
        />
      ) : null}
      <Screen bottomInset="safe-area" topInset="navigation-bar">
        <KeyboardAwareScrollView
          contentContainerStyle={styles.content}
          showsVerticalScrollIndicator={false}
        >
          <Card>
            <View style={styles.form}>
              <View style={styles.sportSelector}>
                <Text variant="bodyLarge">{t("tournament_sport_label")}</Text>
                <View style={styles.sportOptions}>
                  <ConfigurationOption
                    label={t("tournament_sport_football")}
                    onPress={() => setSport("football")}
                    selected={sport === "football"}
                  />
                  <ConfigurationOption
                    label={t("tournament_sport_basketball")}
                    onPress={() => setSport("basketball")}
                    selected={sport === "basketball"}
                  />
                </View>
              </View>
              <TextField
                error={nameError}
                label={t("league_name_label")}
                maxLength={maximumTournamentNameLength}
                onChangeText={setName}
                validationSubmitted={submitted}
                validationTrigger="blur"
                value={name}
              />
              {teams.map((team, index) => (
                <TextField
                  key={index}
                  error={teamError(index)}
                  label={t("league_team_label").replace("{number}", String(index + 1))}
                  onChangeText={(value) =>
                    setTeams((current) =>
                      current.map((item, itemIndex) => (itemIndex === index ? value : item)),
                    )
                  }
                  ref={(input) => {
                    teamInputRefs.current[index] = input;
                  }}
                  validationSubmitted={submitted}
                  validationTrigger="blur"
                  value={team}
                />
              ))}
              <Button label={t("league_add_team")} onPress={addTeam} variant="secondary" />
              <Button
                label={t(user ? "league_publish" : "league_sign_in_to_publish")}
                loading={isSubmitting}
                onPress={() => void publish()}
              />
            </View>
          </Card>
        </KeyboardAwareScrollView>
      </Screen>
    </>
  );
}
const styles = StyleSheet.create({
  content: { gap: space[5] },
  form: { gap: space[4] },
  sportOptions: { flexDirection: "row", flexWrap: "wrap", gap: space[2] },
  sportSelector: { gap: space[2] },
});
