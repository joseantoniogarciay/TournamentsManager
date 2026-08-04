import { Stack, router } from "expo-router";
import { Pressable, StyleSheet } from "react-native";

import { control } from "@tournaments-manager/design-tokens";

import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { useSession } from "@/shared/session/session-provider";
import { Text } from "@/shared/ui";
import { WebIcon } from "@/shared/ui/web-icon";

export default function AccountLayout() {
  const t = getTranslator();
  const { colors } = usePreferences();
  const { user } = useSession();
  const headerOptions = {
    headerBackButtonDisplayMode: "minimal" as const,
    headerShadowVisible: false,
    headerStyle: { backgroundColor: colors.surface.canvas },
    headerTintColor: colors.text.primary,
    headerTitleAlign: "center" as const,
    headerTitleStyle: { color: colors.text.primary },
  };

  return (
    <Stack screenOptions={headerOptions}>
      <Stack.Screen
        name="index"
        options={{
          headerLeft: () => <Text variant="title">{user?.username}</Text>,
          headerRight: () => (
            <Pressable
              accessibilityLabel={t("account_settings_accessibility_label")}
              accessibilityRole="button"
              onPress={() => router.push("/account/settings")}
              style={styles.headerButton}
            >
              <WebIcon color={colors.text.primary} name="settings" size={control.iconSize} />
            </Pressable>
          ),
          headerTitle: () => null,
        }}
      />
      <Stack.Screen name="access" options={{ title: t("account_access_data_title") }} />
      <Stack.Screen name="register" options={{ title: t("account_register_title") }} />
      <Stack.Screen name="forgot-password" options={{ title: t("password_recovery_title") }} />
      <Stack.Screen name="password" options={{ title: t("account_password_change_title") }} />
      <Stack.Screen
        name="google-link"
        options={{ presentation: "modal", title: t("account_google_link_title") }}
      />
      <Stack.Screen
        name="settings"
        options={{
          headerLeft: () => (
            <Pressable
              accessibilityLabel={t("common_close")}
              accessibilityRole="button"
              onPress={() => {
                if (router.canGoBack()) {
                  router.back();
                  return;
                }
                router.replace("/account");
              }}
              style={styles.headerButton}
            >
              <WebIcon color={colors.text.primary} name="close" size={control.iconSize} />
            </Pressable>
          ),
          presentation: "card",
          title: t("account_settings_title"),
        }}
      />
    </Stack>
  );
}

const styles = StyleSheet.create({
  headerButton: {
    alignItems: "center",
    justifyContent: "center",
    minHeight: control.minHeight,
    minWidth: control.minHeight,
  },
});
