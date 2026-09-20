import { SymbolView } from "expo-symbols";
import { useState } from "react";
import { ActivityIndicator, Platform, Pressable, Share, StyleSheet, View } from "react-native";

import { control, space } from "@tournaments-manager/design-tokens";

import type { PublicTournament } from "@/api/generated/models";
import {
  addTournamentTeamRequest,
  createTournamentTeamInvitationRequest,
  removeTournamentTeamRequest,
  revokeTournamentTeamInvitationRequest,
  withdrawTournamentTeamRequest,
} from "@/features/league-creation/api";
import { useTournamentStore } from "@/features/league-creation/league-store";
import { maximumTeamNameLength, maximumTournamentTeams } from "@/features/league-creation/draft";
import { isShareCancellation } from "@/features/league-creation/share";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { Button, Card, ModalDialog, Text, TextField, useConfirmationDialog } from "@/shared/ui";
import { WebIcon } from "@/shared/ui/web-icon";

const localAppLinkURL = "http://localhost:8082";

export function TournamentTeamManagement({
  relationship,
  tournament,
}: {
  relationship: string | null;
  tournament: PublicTournament;
}) {
  const t = getTranslator();
  const { colors } = usePreferences();
  const { show } = useFeedback();
  const { confirm } = useConfirmationDialog();
  const { updateTournament } = useTournamentStore();
  const [adding, setAdding] = useState(false);
  const [name, setName] = useState("");
  const [saving, setSaving] = useState(false);
  const [removingTeamID, setRemovingTeamID] = useState<string>();
  const [isSharingInvitation, setIsSharingInvitation] = useState(false);
  const [isRevokingInvitation, setIsRevokingInvitation] = useState(false);
  const canAddTeam = relationship === "organizer" && tournament.state === "published";
  const teamLimitReached = tournament.teams.length >= maximumTournamentTeams;
  const canRemoveTeam = canAddTeam && tournament.teams.length > 1;
  const canWithdrawTeam =
    relationship === "organizer" &&
    tournament.state === "in_progress" &&
    tournament.format === "league";

  const openAddTeam = () => {
    if (teamLimitReached) {
      show({ kind: "generic-error", message: t("league_team_limit_reached") });
      return;
    }
    setAdding(true);
  };
  const dismissDialog = () => {
    if (saving) return;
    setAdding(false);
    setName("");
  };
  const save = async () => {
    if (!name.trim()) return;
    if (teamLimitReached) {
      show({ kind: "generic-error", message: t("league_team_limit_reached") });
      return;
    }
    setSaving(true);
    try {
      const team = await addTournamentTeamRequest(tournament.id, { name: name.trim() });
      updateTournament(tournament.id, (current) => ({
        ...current,
        teams: [...current.teams, team],
      }));
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
    if (isSharingInvitation) return;
    if (teamLimitReached) {
      show({ kind: "generic-error", message: t("league_team_limit_reached") });
      return;
    }
    setIsSharingInvitation(true);
    try {
      const token = await createTournamentTeamInvitationRequest(tournament.id);
      const base = (
        process.env.EXPO_PUBLIC_APP_LINK_URL ??
        (process.env.APP_ENV === "production" ? undefined : localAppLinkURL)
      )?.replace(/\/$/, "");
      if (!base) throw new Error("missing app link URL");
      const url = `${base}/join-team#${token}`;
      await Share.share({
        message: t("league_invitation_share_message")
          .replace("{tournament}", tournament.name)
          .replace("{url}", url),
      });
    } catch (error) {
      if (isShareCancellation(error)) return;
      const failure = getRequestFailure(error);
      show({ kind: failure.kind, message: t(failure.messageKey) });
    } finally {
      setIsSharingInvitation(false);
    }
  };
  const revokeInvitation = async () => {
    if (isRevokingInvitation) return;
    setIsRevokingInvitation(true);
    try {
      await revokeTournamentTeamInvitationRequest(tournament.id);
      show({ kind: "success", message: t("league_invitation_revoked") });
    } catch (error) {
      const failure = getRequestFailure(error);
      show({ kind: failure.kind, message: t(failure.messageKey) });
    } finally {
      setIsRevokingInvitation(false);
    }
  };
  const remove = async (teamID: string, withdrawn: boolean) => {
    if ((!canRemoveTeam && !canWithdrawTeam) || removingTeamID) return;
    setRemovingTeamID(teamID);
    try {
      if (withdrawn) {
        const updatedTournament = await withdrawTournamentTeamRequest(tournament.id, teamID);
        updateTournament(tournament.id, () => updatedTournament);
      } else {
        await removeTournamentTeamRequest(tournament.id, teamID);
        updateTournament(tournament.id, (current) => ({
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
          ? tournament.sport === "basketball"
            ? "basketball_withdraw_team_description"
            : "league_withdraw_team_description"
          : "league_remove_team_description",
      ).replace("{name}", teamName),
      acceptLabel: t(withdrawn ? "league_withdraw_team" : "league_remove_team"),
      cancelLabel: t("common_cancel"),
      onAccept: () => void remove(teamID, withdrawn),
      onCancel: () => undefined,
    });

  return (
    <>
      {canAddTeam ? (
        <Card>
          <View style={styles.invitationCard}>
            <View style={styles.invitationCopy}>
              <Text variant="title">{t("league_complete_teams_title")}</Text>
              <Text color="secondary">{t("league_complete_teams_description")}</Text>
              {tournament.teams.length < 2 ? (
                <Text color="secondary">{t("league_start_requires_two_teams")}</Text>
              ) : null}
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
      {tournament.teams.map((team) => (
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
    </>
  );
}

const styles = StyleSheet.create({
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
