import { router, useFocusEffect } from "expo-router";
import { useCallback, useState } from "react";
import { Pressable, ScrollView, StyleSheet, View } from "react-native";

import { control, space } from "@tournaments-manager/design-tokens";

import {
  AccountDeletionError,
  deleteAccount,
  getAccountAccessMethods,
} from "@/features/account-access/api";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { useSession } from "@/shared/session/session-provider";
import {
  DisclosureIndicator,
  Screen,
  Text,
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
      void getAccountAccessMethods()
        .then(setAccessMethods)
        .catch(() => setAccessMethods(null));
    }, [user]),
  );

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
          <AccessNavigationRow
            label={t("account_password_change_title")}
            onPress={() => router.push("/account/password" as never)}
          />
          <AccessNavigationRow
            label={t(
              accessMethods?.methods.google
                ? "account_google_linked"
                : "account_access_google_action",
            )}
            onPress={() => router.push("/account/google-link" as never)}
          />
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
});
