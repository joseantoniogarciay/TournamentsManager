import { router } from "expo-router";
import { useEffect, useState } from "react";
import { StyleSheet, View } from "react-native";

import { space } from "@tournaments-manager/design-tokens";

import { createLeagueRequest } from "@/features/league-creation/api";
import {
  clearLocalLeagueDraft,
  getLocalLeagueDraft,
  saveLocalLeagueDraft,
} from "@/features/league-creation/draft";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import { getTranslator } from "@/shared/i18n/locale";
import { useSession } from "@/shared/session/session-provider";
import { Button, Card, KeyboardAwareScrollView, Screen, Text, TextField } from "@/shared/ui";

export default function CreateTournamentScreen() {
  const t = getTranslator();
  const { show } = useFeedback();
  const { user } = useSession();
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

  const normalizedTeams = teams.map((team) => team.trim()).filter(Boolean);
  const nameError = !name.trim() ? t("league_name_required") : undefined;
  const teamsError =
    normalizedTeams.length < 2 ||
    new Set(normalizedTeams.map((team) => team.toLowerCase())).size !== normalizedTeams.length
      ? t("league_teams_required")
      : undefined;
  const publish = async () => {
    setSubmitted(true);
    if (nameError || teamsError) return;
    if (!user) {
      router.push("/account" as never);
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

  return (
    <Screen bottomInset="safe-area">
      <KeyboardAwareScrollView
        contentContainerStyle={styles.content}
        showsVerticalScrollIndicator={false}
      >
        <Card>
          <View style={styles.form}>
            <Text variant="title">{t("league_create_title")}</Text>
            <Text color="secondary">{t("league_draft_description")}</Text>
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
                error={teamsError}
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
            <Button
              label={t("league_add_team")}
              onPress={() => setTeams((current) => [...current, ""])}
              variant="secondary"
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
  );
}
const styles = StyleSheet.create({ content: { gap: space[5] }, form: { gap: space[4] } });
