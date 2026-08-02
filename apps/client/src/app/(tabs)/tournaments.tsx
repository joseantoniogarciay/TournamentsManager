import { router, useFocusEffect } from "expo-router";
import { useCallback, useState } from "react";
import { Pressable, ScrollView, StyleSheet, View } from "react-native";

import { space } from "@tournaments-manager/design-tokens";

import { getTranslator } from "@/shared/i18n/locale";
import { listRelatedLeagues } from "@/features/league-creation/api";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { useSession } from "@/shared/session/session-provider";
import { Card, Screen, Text, useTabContentBottomPadding } from "@/shared/ui";

export default function TournamentsScreen() {
  const t = getTranslator();
  const { revision, user } = useSession();
  const { show } = useFeedback();
  const tabContentBottomPadding = useTabContentBottomPadding();
  const [administered, setAdministered] = useState<Awaited<ReturnType<typeof listRelatedLeagues>>>(
    [],
  );
  const [followed, setFollowed] = useState<Awaited<ReturnType<typeof listRelatedLeagues>>>([]);
  const [isLoading, setIsLoading] = useState(false);

  useFocusEffect(
    useCallback(() => {
      if (!user) {
        setAdministered([]);
        setFollowed([]);
        setIsLoading(false);
        return;
      }
      setIsLoading(true);
      void Promise.all([listRelatedLeagues("administered"), listRelatedLeagues("followed")])
        .then(([nextAdministered, nextFollowed]) => {
          setAdministered(nextAdministered);
          setFollowed(nextFollowed);
        })
        .catch((error) => {
          const failure = getRequestFailure(error);
          show({ kind: failure.kind, message: t(failure.messageKey) });
        })
        .finally(() => setIsLoading(false));
    }, [show, t, user]),
  );

  return (
    <Screen bottomInset="none">
      <ScrollView
        contentContainerStyle={[styles.content, { paddingBottom: tabContentBottomPadding }]}
        key={revision}
        showsVerticalScrollIndicator={false}
      >
        {!user ? (
          <Card>
            <View style={styles.copy}>
              <Text variant="title">{t("tournaments_title")}</Text>
              <Text color="secondary">{t("tournaments_description")}</Text>
            </View>
          </Card>
        ) : null}
        {user && isLoading ? (
          <Card>
            <Text>{t("common_loading")}</Text>
          </Card>
        ) : null}
        {user && !isLoading ? (
          <LeagueSection
            title={t("tournaments_administered")}
            empty={t("tournaments_administered_empty")}
            leagues={administered}
          />
        ) : null}
        {user && !isLoading ? (
          <LeagueSection
            title={t("tournaments_followed")}
            empty={t("tournaments_followed_empty")}
            leagues={followed}
          />
        ) : null}
      </ScrollView>
    </Screen>
  );
}

function LeagueSection({
  title,
  empty,
  leagues,
}: {
  title: string;
  empty: string;
  leagues: Awaited<ReturnType<typeof listRelatedLeagues>>;
}) {
  const t = getTranslator();
  return (
    <Card>
      <View style={styles.copy}>
        <Text variant="title">{title}</Text>
        {leagues.length === 0 ? (
          <Text color="secondary">{empty}</Text>
        ) : (
          leagues.map((league) => (
            <Pressable
              accessibilityLabel={t("tournaments_open_league").replace("{name}", league.name)}
              accessibilityRole="button"
              key={league.id}
              onPress={() => router.push(`/league/${league.id}` as never)}
              style={styles.row}
            >
              <View style={styles.copy}>
                <Text>{league.name}</Text>
                <Text color="secondary">
                  {league.state === "published" ? t("league_not_started") : league.state}
                </Text>
              </View>
              <Text color="secondary">›</Text>
            </Pressable>
          ))
        )}
      </View>
    </Card>
  );
}

const styles = StyleSheet.create({
  content: { flexGrow: 1, justifyContent: "center" },
  copy: { gap: space[2] },
  row: {
    alignItems: "center",
    flexDirection: "row",
    justifyContent: "space-between",
    minHeight: 44,
  },
});
