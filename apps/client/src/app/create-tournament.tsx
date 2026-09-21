import { router, Stack } from "expo-router";
import { randomUUID } from "expo-crypto";
import { useEffect, useState } from "react";
import { StyleSheet, View } from "react-native";

import { space } from "@tournaments-manager/design-tokens";

import { createTournamentRequest } from "@/features/league-creation/api";
import {
  clearLocalTournamentDraft,
  getLocalTournamentDraft,
  maximumTeamNameLength,
  maximumTournamentNameLength,
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
  const { rememberLastTeamName, syncUser, user } = useSession();
  const { colors } = usePreferences();
  const [name, setName] = useState("");
  const [sport, setSport] = useState<TournamentSport>("football");
  const [team, setTeam] = useState("");
  const [draftId, setDraftId] = useState(() => randomUUID());
  const [draftLoaded, setDraftLoaded] = useState(false);
  const [teamHasLocalValue, setTeamHasLocalValue] = useState(false);
  const [submitted, setSubmitted] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    void getLocalTournamentDraft().then((draft) => {
      if (draft) {
        setDraftId(draft.draftId);
        setName(draft.name);
        setSport(draft.sport);
        setTeam(draft.teams[0] ?? "");
        setTeamHasLocalValue(true);
      }
      setDraftLoaded(true);
    });
  }, []);
  useEffect(() => {
    if (draftLoaded && !teamHasLocalValue) setTeam(user?.lastTeamName ?? "");
  }, [draftLoaded, teamHasLocalValue, user?.id, user?.lastTeamName]);
  useEffect(() => {
    if (user) void syncUser().catch(() => undefined);
  }, [syncUser, user?.id]);
  useEffect(() => {
    if (draftLoaded) void saveLocalTournamentDraft({ draftId, name, sport, teams: [team] });
  }, [draftId, draftLoaded, name, sport, team]);

  const normalizedTeam = team.trim();
  const nameError = !name.trim()
    ? t("league_name_required")
    : name.length > maximumTournamentNameLength
      ? t("league_name_too_long")
      : undefined;
  const teamError = !normalizedTeam ? t("league_team_required") : undefined;
  const publish = async () => {
    setSubmitted(true);
    if (nameError || teamError) return;
    if (!user) {
      router.push("/account-authentication" as never);
      return;
    }
    setIsSubmitting(true);
    try {
      const league = await createTournamentRequest({
        name: name.trim(),
        sport,
        teams: [{ name: normalizedTeam }],
      });
      await rememberLastTeamName(normalizedTeam).catch(() => undefined);
      await clearLocalTournamentDraft();
      router.replace(`/tournament/${league.id}` as never);
    } catch (error) {
      const failure = getRequestFailure(error);
      show({ kind: failure.kind, message: t(failure.messageKey) });
    } finally {
      setIsSubmitting(false);
    }
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
                  <ConfigurationOption
                    label={t("tournament_sport_handball")}
                    onPress={() => setSport("handball")}
                    selected={sport === "handball"}
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
              <View style={styles.teamIntroduction}>
                <Text variant="bodyLarge">{t("league_own_team_title")}</Text>
                <Text color="secondary">{t("league_own_team_description")}</Text>
              </View>
              <TextField
                error={teamError}
                label={t("league_own_team_label")}
                maxLength={maximumTeamNameLength}
                onChangeText={(value) => {
                  setTeamHasLocalValue(true);
                  setTeam(value);
                }}
                validationSubmitted={submitted}
                validationTrigger="blur"
                value={team}
              />
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
  teamIntroduction: { gap: space[1] },
});
