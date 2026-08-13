import { router, useFocusEffect } from "expo-router";
import { useCallback, useEffect, useState } from "react";
import { Pressable, ScrollView, StyleSheet, View } from "react-native";

import { control, space } from "@tournaments-manager/design-tokens";

import {
  AccountDeletionError,
  deleteAccount,
  getAccountAccessMethods,
  reauthenticateWithPassword,
  reauthenticateWithGoogle,
  removeAccountPassword,
  unlinkGoogle,
} from "@/features/account-access/api";
import { GoogleLinkDialog } from "@/app/(tabs)/account/google-link";
import { useGoogleIdentityProof } from "@/features/federated-google/use-google-identity-proof";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { useSession } from "@/shared/session/session-provider";
import {
  DisclosureIndicator,
  Button,
  ModalDialog,
  Screen,
  Text,
  TextField,
  useConfirmationDialog,
  useTabContentBottomPadding,
} from "@/shared/ui";

export default function AccountAccessScreen() {
  const t = getTranslator();
  const { colors } = usePreferences();
  const { user } = useSession();
  const { completeAccountDeletion } = useSession();
  const { show } = useFeedback();
  const { confirm } = useConfirmationDialog();
  const tabContentBottomPadding = useTabContentBottomPadding();
  const [accessMethods, setAccessMethods] = useState<Awaited<
    ReturnType<typeof getAccountAccessMethods>
  > | null>(null);
  const [googleLinkVisible, setGoogleLinkVisible] = useState(false);
  const [googleUnlinkVisible, setGoogleUnlinkVisible] = useState(false);
  const [googleUnlinkPassword, setGoogleUnlinkPassword] = useState("");
  const [googleUnlinking, setGoogleUnlinking] = useState(false);
  const [passwordRemovalVisible, setPasswordRemovalVisible] = useState(false);
  const refreshAccessMethods = useCallback(() => {
    return getAccountAccessMethods().then(setAccessMethods);
  }, []);
  const confirmAccountDeletion = () =>
    confirm({
      acceptLabel: t("account_delete"),
      cancelLabel: t("common_cancel"),
      title: t("account_delete_title"),
      description: t("account_delete_description"),
      onCancel: () => undefined,
      onAccept: () =>
        void deleteAccount()
          .then(async () => {
            await completeAccountDeletion();
            show({ kind: "success", message: t("account_delete_success") });
          })
          .catch((error) =>
            show({
              kind: "generic-error",
              message: t(
                error instanceof AccountDeletionError && error.reason === "owned-leagues"
                  ? "account_delete_owned_leagues"
                  : "common_request_error",
              ),
            }),
          ),
    });

  useFocusEffect(
    useCallback(() => {
      if (!user) {
        setAccessMethods(null);
        return;
      }
      void refreshAccessMethods().catch(() => setAccessMethods(null));
    }, [refreshAccessMethods, user]),
  );
  const confirmGoogleUnlink = async () => {
    if (googleUnlinkPassword.length < 8 || googleUnlinking) return;
    setGoogleUnlinking(true);
    try {
      const ticket = await reauthenticateWithPassword(googleUnlinkPassword, "unlink-google");
      await unlinkGoogle(ticket);
      await refreshAccessMethods();
      setGoogleUnlinkVisible(false);
      setGoogleUnlinkPassword("");
    } catch {
      show({ kind: "generic-error", message: t("common_request_error") });
    } finally {
      setGoogleUnlinking(false);
    }
  };
  const requestGoogleUnlink = () => {
    if (!accessMethods?.methods.password) {
      show({ kind: "generic-error", message: t("account_google_unlink_requires_password") });
      return;
    }
    setGoogleUnlinkVisible(true);
  };
  const removePasswordProof = useCallback(
    async (challenge: { id: string }, idToken: string) => {
      const ticket = await reauthenticateWithGoogle(challenge.id, idToken, "remove-local-password");
      await removeAccountPassword(ticket);
      await refreshAccessMethods();
      setPasswordRemovalVisible(false);
    },
    [refreshAccessMethods],
  );
  const removePasswordGoogle = useGoogleIdentityProof(removePasswordProof);
  useEffect(() => {
    if (passwordRemovalVisible) removePasswordGoogle.prepare();
  }, [passwordRemovalVisible, removePasswordGoogle.prepare]);
  useEffect(() => {
    if (!removePasswordGoogle.error) return;
    show({ kind: "generic-error", message: t("common_request_error") });
    setPasswordRemovalVisible(false);
  }, [removePasswordGoogle.error, show, t]);

  return (
    <Screen bottomInset="none" topInset="navigation-bar">
      <ScrollView
        contentContainerStyle={[styles.content, { paddingBottom: tabContentBottomPadding }]}
        showsVerticalScrollIndicator={false}
      >
        <View style={styles.identity}>
          <Text variant="title">{accessMethods?.email ?? "…"}</Text>
          <Text color="secondary">{t("account_access_email_description")}</Text>
        </View>
        <View style={[styles.rows, { borderColor: colors.border.default }]}>
          {accessMethods?.methods.google && accessMethods.methods.password ? (
            <View style={[styles.row, { borderColor: colors.border.default }]}>
              <Pressable
                accessibilityLabel={t("account_password_change_title")}
                accessibilityRole="button"
                onPress={() => router.push("/account/password" as never)}
              >
                <Text variant="bodyLarge">{t("account_password_change_title")}</Text>
              </Pressable>
              <Button
                label={t("account_password_remove")}
                onPress={() => setPasswordRemovalVisible(true)}
                secondarySurfaceColor={colors.surface.canvas}
                variant="secondary"
              />
            </View>
          ) : (
            <AccessNavigationRow
              label={t(
                accessMethods?.methods.password
                  ? "account_password_change_title"
                  : "account_password_add_title",
              )}
              onPress={() => router.push("/account/password" as never)}
            />
          )}
          {accessMethods?.methods.google ? (
            <View style={[styles.row, { borderColor: colors.border.default }]}>
              <Text variant="bodyLarge">{t("account_google_linked")}</Text>
              <Button
                label={t("account_google_unlink")}
                onPress={requestGoogleUnlink}
                secondarySurfaceColor={colors.surface.canvas}
                variant="secondary"
              />
            </View>
          ) : (
            <AccessNavigationRow
              label={t("account_access_google_action")}
              onPress={() => setGoogleLinkVisible(true)}
            />
          )}
        </View>
        <Pressable
          accessibilityLabel={t("account_delete")}
          accessibilityRole="button"
          onPress={confirmAccountDeletion}
          style={styles.deleteAccount}
        >
          <Text style={styles.deleteAccountText}>{t("account_delete")}</Text>
        </Pressable>
      </ScrollView>
      <GoogleLinkDialog
        onDismiss={() => setGoogleLinkVisible(false)}
        onLinked={() => {
          setGoogleLinkVisible(false);
          void refreshAccessMethods();
        }}
        visible={googleLinkVisible}
      />
      <ModalDialog
        dismissAccessibilityLabel={t("common_cancel")}
        onDismiss={() => !googleUnlinking && setGoogleUnlinkVisible(false)}
        visible={googleUnlinkVisible}
      >
        <View style={styles.dialog}>
          <Text variant="title">{t("account_google_unlink_title")}</Text>
          <Text color="secondary">{t("account_google_unlink_description")}</Text>
          <TextField
            autoComplete="current-password"
            label={t("account_password_current_label")}
            onChangeText={setGoogleUnlinkPassword}
            secureTextEntry
            value={googleUnlinkPassword}
          />
          <Button
            disabled={googleUnlinkPassword.length < 8}
            label={t("account_google_unlink")}
            loading={googleUnlinking}
            onPress={() => void confirmGoogleUnlink()}
            variant="secondary"
          />
        </View>
      </ModalDialog>
      <ModalDialog
        dismissAccessibilityLabel={t("common_cancel")}
        onDismiss={() => setPasswordRemovalVisible(false)}
        visible={passwordRemovalVisible}
      >
        <View style={styles.dialog}>
          <Text variant="title">{t("account_password_remove_title")}</Text>
          <Text color="secondary">{t("account_password_remove_description")}</Text>
          <Button
            disabled={!removePasswordGoogle.isConfigured}
            label={t("account_password_google_reauthenticate")}
            loading={removePasswordGoogle.isLoading}
            onPress={() => void removePasswordGoogle.start()}
            variant="secondary"
          />
        </View>
      </ModalDialog>
    </Screen>
  );
}

function AccessNavigationRow({ label, onPress }: { label: string; onPress: () => void }) {
  const { colors } = usePreferences();
  return (
    <Pressable
      accessibilityLabel={label}
      accessibilityRole="button"
      onPress={onPress}
      style={[styles.row, { borderColor: colors.border.default }]}
    >
      <Text variant="bodyLarge">{label}</Text>
      <DisclosureIndicator />
    </Pressable>
  );
}

const styles = StyleSheet.create({
  content: { gap: space[6], marginHorizontal: space[5] },
  identity: { gap: space[2] },
  row: {
    alignItems: "center",
    borderBottomWidth: 1,
    flexDirection: "row",
    justifyContent: "space-between",
    minHeight: control.minHeight + space[5],
  },
  rows: { borderTopWidth: 1 },
  deleteAccount: { alignSelf: "flex-end", minHeight: control.minHeight, justifyContent: "center" },
  deleteAccountText: { textDecorationLine: "underline" },
  dialog: { gap: space[4] },
});
