import { useLocalSearchParams } from "expo-router";
import { useEffect, useState } from "react";
import { Pressable, Share, StyleSheet, View } from "react-native";

import { control, space } from "@tournaments-manager/design-tokens";

import type { PublicLeague } from "@/api/generated/models";
import {
  assignLeagueAdministratorRequest,
  cancelLeagueRequest,
  getLeague,
  getLeagueRelationship,
  startLeagueRequest,
} from "@/features/league-creation/api";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import { getTranslator } from "@/shared/i18n/locale";
import { useSession } from "@/shared/session/session-provider";
import { Button, Card, Screen, Text, TextField, useConfirmationDialog } from "@/shared/ui";

export default function LeagueScreen() {
  const t = getTranslator();
  const { id } = useLocalSearchParams<{ id: string }>();
  const { user } = useSession();
  const { show } = useFeedback();
  const { confirm } = useConfirmationDialog();
  const [league, setLeague] = useState<PublicLeague | null>(null);
  const [relationship, setRelationship] = useState<string>();
  const [isStarting, setIsStarting] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);
  const [addingAdministrator, setAddingAdministrator] = useState(false);
  const [username, setUsername] = useState("");
  useEffect(() => {
    if (!id) return;
    void getLeague(id)
      .then(setLeague)
      .catch((error) => {
        const failure = getRequestFailure(error);
        show({ kind: failure.kind, message: t(failure.messageKey) });
      });
    if (user) void getLeagueRelationship(id).then(setRelationship);
  }, [id, show, t, user]);
  const isOrganizer = relationship === "organizer";
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
  const share = async () => {
    if (!id || !league) return;
    const base = process.env.EXPO_PUBLIC_APP_LINK_URL?.replace(/\/$/, "");
    await Share.share({
      message: `${league.name}: ${base ? `${base}/league/${id}` : `/league/${id}`}`,
    });
  };
  const cancel = () =>
    confirm({
      title: t("league_cancel_title"),
      description: t("league_cancel_description"),
      acceptLabel: t("league_cancel"),
      cancelLabel: t("common_cancel"),
      onAccept: () => {
        if (!id) return;
        void cancelLeagueRequest(id)
          .then(setLeague)
          .catch((error) => {
            const failure = getRequestFailure(error);
            show({ kind: failure.kind, message: t(failure.messageKey) });
          });
      },
      onCancel: () => undefined,
    });
  const assign = async () => {
    if (!id) return;
    try {
      await assignLeagueAdministratorRequest(id, username);
      setUsername("");
      setAddingAdministrator(false);
      show({ kind: "success", message: t("league_administrator_added") });
    } catch (error) {
      const failure = getRequestFailure(error);
      show({ kind: failure.kind, message: t(failure.messageKey) });
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
  const canCancel = league.state === "published" || league.state === "in_progress";
  return (
    <Screen>
      <View style={styles.content}>
        <Card>
          <View style={styles.header}>
            <View style={styles.stack}>
              <Text variant="title">{league.name}</Text>
              <Text color="secondary">
                {league.state === "published" ? t("league_not_started") : league.state}
              </Text>
              <Text color="secondary">{league.teams.map((team) => team.name).join(" · ")}</Text>
            </View>
            <Pressable
              accessibilityLabel={t("league_actions")}
              accessibilityRole="button"
              onPress={() => setMenuOpen((value) => !value)}
              style={styles.menuButton}
            >
              <Text variant="title">•••</Text>
            </Pressable>
          </View>
          {menuOpen ? (
            <View style={styles.menu}>
              <Button
                label={t("league_share")}
                onPress={() => {
                  setMenuOpen(false);
                  void share();
                }}
                variant="ghost"
              />
              {isOrganizer ? (
                <>
                  <Button
                    label={t("league_add_administrator")}
                    onPress={() => {
                      setMenuOpen(false);
                      setAddingAdministrator(true);
                    }}
                    variant="ghost"
                  />
                  {canCancel ? (
                    <Button
                      label={t("league_cancel")}
                      onPress={() => {
                        setMenuOpen(false);
                        cancel();
                      }}
                      variant="destructive"
                    />
                  ) : null}
                </>
              ) : null}
            </View>
          ) : null}
        </Card>
        {addingAdministrator ? (
          <Card>
            <View style={styles.stack}>
              <Text variant="title">{t("league_add_administrator")}</Text>
              <TextField
                label={t("league_administrator_username")}
                onChangeText={setUsername}
                value={username}
                autoCapitalize="none"
              />
              <Button
                label={t("common_confirm")}
                onPress={() => void assign()}
                disabled={!username}
              />
            </View>
          </Card>
        ) : null}
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
const styles = StyleSheet.create({
  content: { gap: space[5] },
  stack: { flex: 1, gap: space[3] },
  header: { flexDirection: "row", gap: space[3], justifyContent: "space-between" },
  menuButton: {
    alignItems: "center",
    height: control.minHeight,
    justifyContent: "center",
    width: control.minHeight,
  },
  menu: { gap: space[1], marginTop: space[3] },
});
