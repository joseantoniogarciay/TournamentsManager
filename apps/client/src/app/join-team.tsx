import * as Linking from "expo-linking";
import { router, Stack } from "expo-router";
import { useCallback, useEffect, useState } from "react";
import { StyleSheet, View } from "react-native";

import { space } from "@tournaments-manager/design-tokens";

import {
  inspectTournamentTeamInvitationRequest,
  joinTournamentWithTeamInvitationRequest,
  TournamentTeamInvitationConflictError,
  TournamentTeamInvitationUnavailableError,
} from "@/features/league-creation/api";
import { maximumTeamNameLength } from "@/features/league-creation/draft";
import {
  clearPendingTeamInvitation,
  getPendingTeamInvitation,
  isTeamInvitationToken,
  rememberPendingTeamInvitation,
} from "@/features/league-creation/team-invitation";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { useSession } from "@/shared/session/session-provider";
import {
  Button,
  Card,
  KeyboardAwareScrollView,
  LoadingTransition,
  NavigationHeaderButton,
  RequestErrorCard,
  Screen,
  Text,
  TextField,
  usesLiquidGlassNavigation,
} from "@/shared/ui";

type Invitation = { tournamentId: string; tournamentName: string };

export default function JoinTeamScreen() {
  const t = getTranslator();
  const url = Linking.useURL();
  const { colors } = usePreferences();
  const { show } = useFeedback();
  const { rememberLastTeamName, syncUser, user } = useSession();
  const [token, setToken] = useState<string>();
  const [invitation, setInvitation] = useState<Invitation>();
  const [name, setName] = useState("");
  const [nameEdited, setNameEdited] = useState(false);
  const [initialized, setInitialized] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitted, setSubmitted] = useState(false);
  const [loadError, setLoadError] = useState<"unavailable" | "request">();

  useEffect(() => {
    if (!nameEdited) setName(user?.lastTeamName ?? "");
  }, [nameEdited, user?.id, user?.lastTeamName]);
  useEffect(() => {
    if (user) void syncUser().catch(() => undefined);
  }, [syncUser, user?.id]);

  useEffect(() => {
    let active = true;
    void (async () => {
      const incoming = invitationTokenFromURL(url);
      if (incoming.present) {
        try {
          if (incoming.token) await rememberPendingTeamInvitation(incoming.token);
          else await clearPendingTeamInvitation();
          if (!active) return;
          setToken(incoming.token ?? undefined);
        } catch {
          if (!active) return;
          setToken(undefined);
        } finally {
          if (active) setInitialized(true);
          if (active) router.replace("/join-team" as never);
        }
        return;
      }
      const pendingToken = await getPendingTeamInvitation();
      if (!active) return;
      setToken(pendingToken ?? undefined);
      setInitialized(true);
    })();
    return () => {
      active = false;
    };
  }, [url]);

  const load = useCallback(async () => {
    if (!token) {
      setLoadError("unavailable");
      return;
    }
    setIsLoading(true);
    setLoadError(undefined);
    try {
      setInvitation(await inspectTournamentTeamInvitationRequest(token));
    } catch (error) {
      setLoadError(
        error instanceof TournamentTeamInvitationUnavailableError ? "unavailable" : "request",
      );
      if (error instanceof TournamentTeamInvitationUnavailableError) {
        await clearPendingTeamInvitation();
      }
    } finally {
      setIsLoading(false);
    }
  }, [token]);

  useEffect(() => {
    if (initialized) void load();
  }, [initialized, load]);

  const close = async () => {
    await clearPendingTeamInvitation();
    if (router.canDismiss()) {
      router.dismiss();
      return;
    }
    router.replace("/");
  };
  const normalizedName = name.trim();
  const nameError = !normalizedName
    ? t("team_invitation_name_required")
    : normalizedName.length > maximumTeamNameLength
      ? t("team_invitation_name_too_long")
      : undefined;
  const join = async () => {
    setSubmitted(true);
    if (nameError || !token || !invitation) return;
    if (!user) {
      router.push("/account-authentication?destination=join-team" as never);
      return;
    }
    setIsSubmitting(true);
    try {
      const registration = await joinTournamentWithTeamInvitationRequest(token, normalizedName);
      await rememberLastTeamName(normalizedName).catch(() => undefined);
      await clearPendingTeamInvitation();
      router.replace(`/tournament/${registration.tournamentId}` as never);
    } catch (error) {
      if (error instanceof TournamentTeamInvitationUnavailableError) {
        setInvitation(undefined);
        setLoadError("unavailable");
        await clearPendingTeamInvitation();
      } else if (error instanceof TournamentTeamInvitationConflictError) {
        show({ kind: "generic-error", message: t("team_invitation_conflict") });
      } else {
        const failure = getRequestFailure(error);
        show({ kind: failure.kind, message: t(failure.messageKey) });
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <>
      <Stack.Screen
        options={{
          headerBackVisible: false,
          headerShadowVisible: false,
          headerStyle: { backgroundColor: colors.surface.canvas },
          headerTintColor: colors.text.primary,
          headerTitle: t("team_invitation_title"),
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
        {!initialized || isLoading ? (
          <LoadingTransition active message={t("common_loading")} />
        ) : loadError || !invitation ? (
          <RequestErrorCard
            actionLabel={t(loadError === "request" ? "common_retry" : "common_close")}
            loading={isLoading}
            message={t(
              loadError === "request" ? "common_request_error" : "team_invitation_unavailable",
            )}
            onRetry={loadError === "request" ? () => void load() : () => void close()}
          />
        ) : (
          <KeyboardAwareScrollView
            contentContainerStyle={styles.content}
            showsVerticalScrollIndicator={false}
          >
            <Card>
              <View style={styles.form}>
                <View style={styles.introduction}>
                  <Text variant="title">{invitation.tournamentName}</Text>
                  <Text color="secondary">{t("team_invitation_description")}</Text>
                  <Text color="secondary">{t("team_invitation_permissions")}</Text>
                </View>
                <TextField
                  error={nameError}
                  label={t("team_invitation_name_label")}
                  maxLength={maximumTeamNameLength}
                  onChangeText={(value) => {
                    setNameEdited(true);
                    setName(value);
                  }}
                  validationSubmitted={submitted}
                  validationTrigger="blur"
                  value={name}
                />
                <Button
                  label={t(user ? "team_invitation_join" : "team_invitation_sign_in")}
                  loading={isSubmitting}
                  onPress={() => void join()}
                />
              </View>
            </Card>
          </KeyboardAwareScrollView>
        )}
      </Screen>
    </>
  );
}

function invitationTokenFromURL(url: string | null): { present: boolean; token: string | null } {
  if (!url) return { present: false, token: null };
  const hashIndex = url.indexOf("#");
  if (hashIndex === -1) return { present: false, token: null };
  const value = url.slice(hashIndex + 1);
  return { present: true, token: isTeamInvitationToken(value) ? value : null };
}

const styles = StyleSheet.create({
  content: { gap: space[5] },
  form: { gap: space[4] },
  introduction: { gap: space[2] },
});
