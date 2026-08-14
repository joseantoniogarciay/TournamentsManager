import { router, Stack } from "expo-router";
import { useEffect, useState } from "react";
import { Platform, StyleSheet, View } from "react-native";

import { control, radius, space } from "@tournaments-manager/design-tokens";

import { createLeagueRequest } from "@/features/league-creation/api";
import {
  clearLocalLeagueDraft,
  getLocalLeagueDraft,
  maximumLeagueNameLength,
  maximumLeagueTeams,
  saveLocalLeagueDraft,
} from "@/features/league-creation/draft";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { useSession } from "@/shared/session/session-provider";
import {
  Button,
  Card,
  KeyboardAwareScrollView,
  NavigationHeaderButton,
  Screen,
  TextField,
} from "@/shared/ui";

export default function CreateTournamentScreen() {
  const t = getTranslator();
  const { show } = useFeedback();
  const { user } = useSession();
  const { colors } = usePreferences();
  const [name, setName] = useState("");
  const [teams, setTeams] = useState(["", ""]);
  const [submitted, setSubmitted] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    void getLocalLeagueDraft().then((draft) => {
      if (draft) {
        setName(draft.name);
        setTeams(draft.teams.length >= 2 ? draft.teams : ["", ""]);
      }
    });
  }, []);
  useEffect(() => {
    void saveLocalLeagueDraft({ name, teams });
  }, [name, teams]);

  const normalizedTeamValues = teams.map((team) => team.trim());
  const normalizedTeams = normalizedTeamValues.filter(Boolean);
  const nameError = !name.trim()
    ? t("league_name_required")
    : name.length > maximumLeagueNameLength
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
      const league = await createLeagueRequest({
        name: name.trim(),
        teams: normalizedTeams.map((team) => ({ name: team })),
      });
      await clearLocalLeagueDraft();
      router.replace(`/league/${league.id}` as never);
    } catch (error) {
      const failure = getRequestFailure(error);
      show({ kind: failure.kind, message: t(failure.messageKey) });
    } finally {
      setIsSubmitting(false);
    }
  };
  const addTeam = () => {
    if (teams.length >= maximumLeagueTeams) {
      show({ kind: "generic-error", message: t("league_team_limit_reached") });
      return;
    }
    setTeams((current) => [...current, ""]);
  };
  const close = async () => {
    await clearLocalLeagueDraft();
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
        {Platform.OS === "ios" ? (
          <Stack.Toolbar placement="left">
            <Stack.Toolbar.Button
              accessibilityLabel={t("common_close")}
              icon="xmark"
              onPress={() => void close()}
            />
          </Stack.Toolbar>
        ) : null}
      </Stack.Screen>
      {Platform.OS !== "ios" ? (
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
              <TextField
                error={nameError}
                label={t("league_name_label")}
                maxLength={maximumLeagueNameLength}
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
  navigationButton: {
    alignItems: "center",
    borderRadius: radius.pill,
    borderWidth: 1,
    height: control.minHeight,
    justifyContent: "center",
    width: control.minHeight,
  },
});
