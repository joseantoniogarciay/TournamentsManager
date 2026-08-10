import { router, Stack } from "expo-router";
import { SymbolView } from "expo-symbols";
import { useEffect, useState } from "react";
import { Platform, Pressable, StyleSheet, View } from "react-native";

import { control, radius, space } from "@tournaments-manager/design-tokens";

import { createLeagueRequest } from "@/features/league-creation/api";
import {
  clearLocalLeagueDraft,
  getLocalLeagueDraft,
  maximumLeagueTeams,
  saveLocalLeagueDraft,
} from "@/features/league-creation/draft";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { useSession } from "@/shared/session/session-provider";
import { Button, Card, KeyboardAwareScrollView, Screen, TextField } from "@/shared/ui";
import { WebIcon } from "@/shared/ui/web-icon";

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
  const nameError = !name.trim() ? t("league_name_required") : undefined;
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
              <Pressable
                accessibilityLabel={t("common_close")}
                accessibilityRole="button"
                onPress={() => void close()}
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
