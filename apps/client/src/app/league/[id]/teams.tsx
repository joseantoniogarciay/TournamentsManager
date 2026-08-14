import { router, Stack, useLocalSearchParams } from "expo-router";
import { SymbolView } from "expo-symbols";
import { useCallback, useEffect, useState } from "react";
import {
  ActivityIndicator,
  Modal,
  Platform,
  Pressable,
  ScrollView,
  StyleSheet,
  View,
} from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";

import { control, radius, space } from "@tournaments-manager/design-tokens";

import {
  addLeagueTeamRequest,
  getLeagueRelationship,
  LeagueUnavailableError,
  removeLeagueTeamRequest,
  withdrawLeagueTeamRequest,
} from "@/features/league-creation/api";
import { useLeague, useLeagueStore } from "@/features/league-creation/league-store";
import { maximumLeagueTeams } from "@/features/league-creation/draft";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { useSession } from "@/shared/session/session-provider";
import {
  Button,
  Card,
  LoadingTransition,
  NavigationHeaderButton,
  RequestErrorCard,
  Screen,
  Text,
  TextField,
  useConfirmationDialog,
} from "@/shared/ui";
import { WebIcon } from "@/shared/ui/web-icon";

export default function LeagueTeamsScreen() {
  const t = getTranslator();
  const { id } = useLocalSearchParams<{ id: string }>();
  const { user } = useSession();
  const { colors } = usePreferences();
  const insets = useSafeAreaInsets();
  const { show } = useFeedback();
  const { confirm } = useConfirmationDialog();
  const league = useLeague(id);
  const { loadLeague, refreshLeague, updateLeague } = useLeagueStore();
  const [relationship, setRelationship] = useState<string>();
  const [loadErrorMessage, setLoadErrorMessage] = useState<string>();
  const [leagueUnavailable, setLeagueUnavailable] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [adding, setAdding] = useState(false);
  const [name, setName] = useState("");
  const [saving, setSaving] = useState(false);
  const [removingTeamID, setRemovingTeamID] = useState<string>();

  const load = useCallback(
    async (force = false) => {
      if (!id) {
        setLoadErrorMessage(t("common_request_error"));
        return;
      }
      setIsLoading(true);
      setLoadErrorMessage(undefined);
      setLeagueUnavailable(false);
      try {
        await (force ? refreshLeague(id) : loadLeague(id));
      } catch (error) {
        const unavailable = error instanceof LeagueUnavailableError;
        setLeagueUnavailable(unavailable);
        setLoadErrorMessage(
          t(unavailable ? "league_unavailable" : getRequestFailure(error).messageKey),
        );
      } finally {
        setIsLoading(false);
      }
    },
    [id, loadLeague, refreshLeague, t],
  );
  useEffect(() => {
    void load();
    if (user)
      void getLeagueRelationship(id)
        .then(setRelationship)
        .catch(() => setRelationship(undefined));
  }, [id, load, user]);

  const canAddTeam = relationship === "organizer" && league?.state === "published";
  const canRemoveTeam = canAddTeam && (league?.teams.length ?? 0) > 2;
  const canWithdrawTeam = relationship === "organizer" && league?.state === "in_progress";
  const openAddTeam = () => {
    if ((league?.teams.length ?? 0) >= maximumLeagueTeams) {
      show({ kind: "generic-error", message: t("league_team_limit_reached") });
      return;
    }
    setAdding(true);
  };
  const close = () => {
    if (router.canDismiss()) {
      router.dismiss();
      return;
    }
    router.replace(id ? `/league/${id}` : "/");
  };
  const closeUnavailable = () => {
    if (router.canDismiss()) {
      router.dismiss();
      return;
    }
    router.replace("/");
  };
  const dismissDialog = () => {
    if (saving) return;
    setAdding(false);
    setName("");
  };
  const save = async () => {
    if (!id || !name.trim()) return;
    if ((league?.teams.length ?? 0) >= maximumLeagueTeams) {
      show({ kind: "generic-error", message: t("league_team_limit_reached") });
      return;
    }
    setSaving(true);
    try {
      const team = await addLeagueTeamRequest(id, { name: name.trim() });
      updateLeague(id, (current) => ({ ...current, teams: [...current.teams, team] }));
      setAdding(false);
      setName("");
    } catch (error) {
      const failure = getRequestFailure(error);
      show({ kind: failure.kind, message: t(failure.messageKey) });
    } finally {
      setSaving(false);
    }
  };
  const remove = async (teamID: string, withdrawn: boolean) => {
    if (!id || (!canRemoveTeam && !canWithdrawTeam)) return;
    setRemovingTeamID(teamID);
    try {
      if (withdrawn) {
        const updatedLeague = await withdrawLeagueTeamRequest(id, teamID);
        updateLeague(id, () => updatedLeague);
      } else {
        await removeLeagueTeamRequest(id, teamID);
        updateLeague(id, (current) => ({
          ...current,
          teams: current.teams.filter((team) => team.id !== teamID),
        }));
      }
    } catch (error) {
      const failure = getRequestFailure(error);
      show({ kind: failure.kind, message: t(failure.messageKey) });
    } finally {
      setRemovingTeamID(undefined);
    }
  };
  const confirmRemove = (teamID: string, teamName: string, withdrawn: boolean) =>
    confirm({
      title: t(withdrawn ? "league_withdraw_team_title" : "league_remove_team_title"),
      description: t(
        withdrawn ? "league_withdraw_team_description" : "league_remove_team_description",
      ).replace("{name}", teamName),
      acceptLabel: t(withdrawn ? "league_withdraw_team" : "league_remove_team"),
      cancelLabel: t("common_cancel"),
      onAccept: () => void remove(teamID, withdrawn),
      onCancel: () => undefined,
    });
  const navigationButton = (
    onPress: () => void,
    label: string,
    icon: "close" | "add",
    side: "left" | "right",
  ) => (
    <NavigationHeaderButton
      accessibilityLabel={label}
      icon={icon}
      nativeIcon={
        icon === "add"
          ? { android: "add", ios: "plus", web: "add" }
          : { android: "close", ios: "xmark", web: "close" }
      }
      onPress={onPress}
      side={side}
    />
  );

  return (
    <>
      <Stack.Screen
        options={{
          headerBackVisible: false,
          headerShadowVisible: false,
          headerStyle: { backgroundColor: colors.surface.canvas },
          headerTintColor: colors.text.primary,
          headerTitleAlign: "center",
          title: t("league_teams"),
          ...(Platform.OS !== "ios"
            ? {
                headerLeft: () => navigationButton(close, t("common_back"), "close", "left"),
                headerRight: canAddTeam
                  ? () => navigationButton(openAddTeam, t("league_add_team"), "add", "right")
                  : undefined,
              }
            : {}),
        }}
      />
      {Platform.OS === "ios" ? (
        <>
          <Stack.Toolbar placement="left">
            <Stack.Toolbar.Button
              accessibilityLabel={t("common_back")}
              icon="xmark"
              onPress={close}
            />
          </Stack.Toolbar>
          {canAddTeam ? (
            <Stack.Toolbar placement="right">
              <Stack.Toolbar.Button
                accessibilityLabel={t("league_add_team")}
                icon="plus"
                onPress={openAddTeam}
              />
            </Stack.Toolbar>
          ) : null}
        </>
      ) : null}
      <Screen bottomInset="none" topInset="navigation-bar">
        {!league ? (
          loadErrorMessage ? (
            <RequestErrorCard
              actionLabel={t(leagueUnavailable ? "common_close" : "common_retry")}
              loading={leagueUnavailable ? false : isLoading}
              message={loadErrorMessage}
              onRetry={leagueUnavailable ? closeUnavailable : () => void load(true)}
            />
          ) : (
            <LoadingTransition active message={t("common_loading")} />
          )
        ) : (
          <ScrollView
            contentContainerStyle={[styles.content, { paddingBottom: insets.bottom + space[5] }]}
            showsVerticalScrollIndicator={false}
          >
            {league.teams.map((team) => (
              <Card density="compact" key={team.id}>
                <View style={styles.teamRow}>
                  <Text style={styles.teamName} variant="bodyLarge">
                    {team.name}
                  </Text>
                  {team.withdrawn ? (
                    <Text style={styles.withdrawn} variant="caption">
                      {t("league_team_withdrawn")}
                    </Text>
                  ) : canRemoveTeam || canWithdrawTeam ? (
                    <Pressable
                      accessibilityLabel={t(
                        canWithdrawTeam ? "league_withdraw_team" : "league_remove_team",
                      )}
                      accessibilityRole="button"
                      accessibilityState={{ busy: removingTeamID === team.id }}
                      disabled={removingTeamID !== undefined}
                      onPress={() => confirmRemove(team.id, team.name, canWithdrawTeam)}
                      style={styles.removeButton}
                    >
                      {removingTeamID === team.id ? (
                        <ActivityIndicator color={colors.text.primary} />
                      ) : Platform.OS === "web" ? (
                        <WebIcon color={colors.text.primary} name="close" size={control.iconSize} />
                      ) : (
                        <SymbolView
                          name={{ android: "close", ios: "xmark", web: "close" }}
                          size={control.iconSize}
                          tintColor={colors.text.primary}
                        />
                      )}
                    </Pressable>
                  ) : null}
                </View>
              </Card>
            ))}
          </ScrollView>
        )}
        <Modal animationType="fade" onRequestClose={dismissDialog} transparent visible={adding}>
          <View style={[styles.backdrop, { backgroundColor: colors.surface.canvas }]}>
            <Pressable
              accessibilityLabel={t("common_close")}
              accessibilityRole="button"
              disabled={saving}
              onPress={dismissDialog}
              style={StyleSheet.absoluteFill}
            />
            <View
              accessibilityViewIsModal
              style={[
                styles.dialog,
                { backgroundColor: colors.surface.default, borderColor: colors.border.default },
              ]}
            >
              <Text variant="title">{t("league_add_team_title")}</Text>
              <TextField label={t("league_add_team_name")} onChangeText={setName} value={name} />
              <Button
                disabled={!name.trim()}
                label={t("league_add_team_save")}
                loading={saving}
                onPress={() => void save()}
              />
              <Button
                disabled={saving}
                label={t("common_cancel")}
                onPress={dismissDialog}
                variant="secondary"
              />
            </View>
          </View>
        </Modal>
      </Screen>
    </>
  );
}

const styles = StyleSheet.create({
  backdrop: {
    ...StyleSheet.absoluteFill,
    alignItems: "center",
    justifyContent: "center",
    padding: space[5],
  },
  content: { gap: space[5], paddingBottom: space[5] },
  dialog: {
    borderRadius: radius.card,
    borderWidth: 1,
    gap: space[5],
    maxWidth: 440,
    padding: space[5],
    width: "100%",
  },
  loader: { alignItems: "center", flex: 1, justifyContent: "center" },
  navigationButton: {
    alignItems: "center",
    borderRadius: radius.pill,
    borderWidth: 1,
    height: control.minHeight,
    justifyContent: "center",
    width: control.minHeight,
  },
  removeButton: {
    alignItems: "center",
    height: control.minHeight,
    justifyContent: "center",
    width: control.minHeight,
  },
  teamName: { flex: 1 },
  withdrawn: { flex: 1 },
  teamRow: { alignItems: "center", flexDirection: "row", gap: space[3] },
});
