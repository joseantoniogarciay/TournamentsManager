import { router, Stack, useLocalSearchParams } from "expo-router";
import { SymbolView } from "expo-symbols";
import { useCallback, useEffect, useState } from "react";
import {
  ActivityIndicator,
  Platform,
  Pressable,
  ScrollView,
  Share,
  StyleSheet,
  View,
} from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";

import { control, space } from "@tournaments-manager/design-tokens";

import {
  addTournamentTeamRequest,
  createTournamentTeamInvitationRequest,
  getTournamentRelationship,
  revokeTournamentTeamInvitationRequest,
  TournamentUnavailableError,
  removeTournamentTeamRequest,
  withdrawTournamentTeamRequest,
} from "@/features/league-creation/api";
import { useTournament, useTournamentStore } from "@/features/league-creation/league-store";
import { maximumTeamNameLength, maximumTournamentTeams } from "@/features/league-creation/draft";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { useSession } from "@/shared/session/session-provider";
import {
  Button,
  Card,
  LoadingTransition,
  ModalDialog,
  NavigationHeaderButton,
  RequestErrorCard,
  Screen,
  Text,
  TextField,
  useConfirmationDialog,
  usesLiquidGlassNavigation,
} from "@/shared/ui";
import { WebIcon } from "@/shared/ui/web-icon";

export default function TournamentTeamsScreen() {
  const t = getTranslator();
  const { id } = useLocalSearchParams<{ id: string }>();
  const { user } = useSession();
  const { colors } = usePreferences();
  const insets = useSafeAreaInsets();
  const { show } = useFeedback();
  const { confirm } = useConfirmationDialog();
  const league = useTournament(id);
  const { loadTournament, refreshTournament, updateTournament } = useTournamentStore();
  const [relationship, setRelationship] = useState<string | null>();
  const [loadErrorMessage, setLoadErrorMessage] = useState<string>();
  const [leagueUnavailable, setTournamentUnavailable] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [adding, setAdding] = useState(false);
  const [name, setName] = useState("");
  const [saving, setSaving] = useState(false);
  const [removingTeamID, setRemovingTeamID] = useState<string>();
  const [isSharingInvitation, setIsSharingInvitation] = useState(false);
  const [isRevokingInvitation, setIsRevokingInvitation] = useState(false);

  const load = useCallback(
    async (force = false) => {
      if (!id) {
        setLoadErrorMessage(t("common_request_error"));
        return;
      }
      setIsLoading(true);
      setLoadErrorMessage(undefined);
      setTournamentUnavailable(false);
      setRelationship(undefined);
      try {
        await (force ? refreshTournament(id) : loadTournament(id));
        if (user) setRelationship(await getTournamentRelationship(id));
      } catch (error) {
        const unavailable = error instanceof TournamentUnavailableError;
        setTournamentUnavailable(unavailable);
        setLoadErrorMessage(
          t(unavailable ? "league_unavailable" : getRequestFailure(error).messageKey),
        );
      } finally {
        setIsLoading(false);
      }
    },
    [id, loadTournament, refreshTournament, t, user],
  );
  useEffect(() => {
    void load();
  }, [load]);

  const canAddTeam = relationship === "organizer" && league?.state === "published";
  const teamLimitReached = (league?.teams.length ?? 0) >= maximumTournamentTeams;
  const canRemoveTeam = canAddTeam && (league?.teams.length ?? 0) > 1;
  const canWithdrawTeam =
    relationship === "organizer" && league?.state === "in_progress" && league.format === "league";
  const openAddTeam = () => {
    if ((league?.teams.length ?? 0) >= maximumTournamentTeams) {
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
    router.replace(id ? `/tournament/${id}` : "/");
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
    if ((league?.teams.length ?? 0) >= maximumTournamentTeams) {
      show({ kind: "generic-error", message: t("league_team_limit_reached") });
      return;
    }
    setSaving(true);
    try {
      const team = await addTournamentTeamRequest(id, { name: name.trim() });
      updateTournament(id, (current) => ({ ...current, teams: [...current.teams, team] }));
      setAdding(false);
      setName("");
    } catch (error) {
      const failure = getRequestFailure(error);
      show({ kind: failure.kind, message: t(failure.messageKey) });
    } finally {
      setSaving(false);
    }
  };
  const shareInvitation = async () => {
    if (!id || !league || isSharingInvitation) return;
    if (teamLimitReached) {
      show({ kind: "generic-error", message: t("league_team_limit_reached") });
      return;
    }
    setIsSharingInvitation(true);
    try {
      const token = await createTournamentTeamInvitationRequest(id);
      const base = (
        process.env.EXPO_PUBLIC_APP_LINK_URL ??
        (process.env.APP_ENV === "production" ? undefined : "http://localhost:8082")
      )?.replace(/\/$/, "");
      if (!base) throw new Error("missing app link URL");
      const url = `${base}/join-team#${token}`;
      await Share.share({
        message: t("league_invitation_share_message")
          .replace("{tournament}", league.name)
          .replace("{url}", url),
      });
    } catch (error) {
      const failure = getRequestFailure(error);
      show({ kind: failure.kind, message: t(failure.messageKey) });
    } finally {
      setIsSharingInvitation(false);
    }
  };
  const revokeInvitation = async () => {
    if (!id || isRevokingInvitation) return;
    setIsRevokingInvitation(true);
    try {
      await revokeTournamentTeamInvitationRequest(id);
      show({ kind: "success", message: t("league_invitation_revoked") });
    } catch (error) {
      const failure = getRequestFailure(error);
      show({ kind: failure.kind, message: t(failure.messageKey) });
    } finally {
      setIsRevokingInvitation(false);
    }
  };
  const remove = async (teamID: string, withdrawn: boolean) => {
    if (!id || (!canRemoveTeam && !canWithdrawTeam)) return;
    setRemovingTeamID(teamID);
    try {
      if (withdrawn) {
        const updatedTournament = await withdrawTournamentTeamRequest(id, teamID);
        updateTournament(id, () => updatedTournament);
      } else {
        await removeTournamentTeamRequest(id, teamID);
        updateTournament(id, (current) => ({
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
        withdrawn
          ? league?.sport === "basketball"
            ? "basketball_withdraw_team_description"
            : "league_withdraw_team_description"
          : "league_remove_team_description",
      ).replace("{name}", teamName),
      acceptLabel: t(withdrawn ? "league_withdraw_team" : "league_remove_team"),
      cancelLabel: t("common_cancel"),
      onAccept: () => void remove(teamID, withdrawn),
      onCancel: () => undefined,
    });
  const navigationButton = (onPress: () => void, label: string, side: "left") => (
    <NavigationHeaderButton
      accessibilityLabel={label}
      icon="close"
      nativeIcon={{ android: "close", ios: "xmark", web: "close" }}
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
          ...(!usesLiquidGlassNavigation
            ? {
                headerLeft: () => navigationButton(close, t("common_back"), "left"),
              }
            : {}),
        }}
      />
      {usesLiquidGlassNavigation ? (
        <>
          <Stack.Toolbar placement="left">
            <Stack.Toolbar.Button
              accessibilityLabel={t("common_back")}
              icon="xmark"
              onPress={close}
            />
          </Stack.Toolbar>
        </>
      ) : null}
      <Screen bottomInset="none" topInset="navigation-bar">
        {loadErrorMessage ? (
          <RequestErrorCard
            actionLabel={t(leagueUnavailable ? "common_close" : "common_retry")}
            loading={leagueUnavailable ? false : isLoading}
            message={loadErrorMessage}
            onRetry={leagueUnavailable ? closeUnavailable : () => void load(true)}
          />
        ) : !league || (user && relationship === undefined) ? (
          <LoadingTransition active message={t("common_loading")} />
        ) : (
          <ScrollView
            contentContainerStyle={[styles.content, { paddingBottom: insets.bottom + space[5] }]}
            showsVerticalScrollIndicator={false}
          >
            {canAddTeam ? (
              <Card>
                <View style={styles.invitationCard}>
                  <View style={styles.invitationCopy}>
                    <Text variant="title">{t("league_complete_teams_title")}</Text>
                    <Text color="secondary">{t("league_complete_teams_description")}</Text>
                    <Text color="secondary">{t("league_invitation_rotation_hint")}</Text>
                  </View>
                  <Button label={t("league_add_team")} onPress={openAddTeam} variant="secondary" />
                  <Button
                    disabled={teamLimitReached}
                    label={t("league_share_invitation")}
                    loading={isSharingInvitation}
                    onPress={() => void shareInvitation()}
                  />
                  <Button
                    label={t("league_revoke_invitation")}
                    loading={isRevokingInvitation}
                    onPress={() => void revokeInvitation()}
                    variant="secondary"
                  />
                </View>
              </Card>
            ) : null}
            {league.teams.map((team) => (
              <Card density="compact" key={team.id}>
                <View style={styles.teamRow}>
                  <Text style={styles.teamName} variant="bodyLarge">
                    {team.name}
                  </Text>
                  {team.withdrawn ? (
                    <Text style={styles.withdrawn} variant="bodyLarge">
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
                        <ActivityIndicator color={colors.indicator.default} />
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
        <ModalDialog
          dismissAccessibilityLabel={t("common_close")}
          onDismiss={dismissDialog}
          visible={adding}
        >
          <Text variant="title">{t("league_add_team_title")}</Text>
          <TextField
            label={t("league_add_team_name")}
            maxLength={maximumTeamNameLength}
            onChangeText={setName}
            value={name}
          />
          <Button
            disabled={!name.trim()}
            label={t("league_add_team_save")}
            loading={saving}
            onPress={() => void save()}
          />
        </ModalDialog>
      </Screen>
    </>
  );
}

const styles = StyleSheet.create({
  content: { gap: space[5], paddingBottom: space[5] },
  invitationCard: { gap: space[4] },
  invitationCopy: { gap: space[2] },
  removeButton: {
    alignItems: "center",
    height: control.minHeight,
    justifyContent: "center",
    width: control.minHeight,
  },
  teamName: { flex: 1, minWidth: 0 },
  withdrawn: { flexShrink: 0, textAlign: "right" },
  teamRow: {
    alignItems: "center",
    flexDirection: "row",
    gap: space[5],
    minHeight: control.minHeight,
  },
});
