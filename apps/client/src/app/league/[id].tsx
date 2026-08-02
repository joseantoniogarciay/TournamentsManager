import { useLocalSearchParams } from "expo-router";
import { useEffect, useState } from "react";
import { StyleSheet, View } from "react-native";

import { space } from "@tournaments-manager/design-tokens";

import { canAdministerLeague, getLeague, startLeagueRequest } from "@/features/league-creation/api";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import { getTranslator } from "@/shared/i18n/locale";
import { useSession } from "@/shared/session/session-provider";
import { Button, Card, Screen, Text } from "@/shared/ui";
import type { PublicLeague } from "@/api/generated/models";

export default function LeagueScreen() {
  const t = getTranslator();
  const { id } = useLocalSearchParams<{ id: string }>();
  const { user } = useSession();
  const { show } = useFeedback();
  const [league, setLeague] = useState<PublicLeague | null>(null);
  const [isOrganizer, setIsOrganizer] = useState(false);
  const [isStarting, setIsStarting] = useState(false);
  useEffect(() => {
    if (!id) return;
    void getLeague(id)
      .then(setLeague)
      .catch((error) => {
        const failure = getRequestFailure(error);
        show({ kind: failure.kind, message: t(failure.messageKey) });
      });
    if (user) void canAdministerLeague(id).then(setIsOrganizer);
  }, [id, show, t, user]);
  const start = async (roundRobinLegs: 1 | 2) => {
    if (!id) return;
    setIsStarting(true);
    try {
      setLeague(await startLeagueRequest(id, { roundRobinLegs }));
    } catch (error) {
      const failure = getRequestFailure(error);
      show({ kind: failure.kind, message: t(failure.messageKey) });
    } finally {
      setIsStarting(false);
    }
  };
  if (!league)
    return (
      <Screen>
        <Card>
          <Text>{t("common_loading")}</Text>
        </Card>
      </Screen>
    );
  return (
    <Screen>
      <View style={styles.content}>
        <Card>
          <View style={styles.stack}>
            <Text variant="title">{league.name}</Text>
            <Text color="secondary">
              {league.state === "published" ? t("league_not_started") : league.state}
            </Text>
            <Text color="secondary">{league.teams.map((team) => team.name).join(" · ")}</Text>
          </View>
        </Card>
        {league.state === "published" && isOrganizer ? (
          <Card>
            <View style={styles.stack}>
              <Text variant="title">{t("league_start_title")}</Text>
              <Button
                label={t("league_start_one_leg")}
                loading={isStarting}
                onPress={() => void start(1)}
              />
              <Button
                label={t("league_start_two_legs")}
                variant="secondary"
                onPress={() => void start(2)}
              />
            </View>
          </Card>
        ) : null}
      </View>
    </Screen>
  );
}
const styles = StyleSheet.create({ content: { gap: space[5] }, stack: { gap: space[3] } });
