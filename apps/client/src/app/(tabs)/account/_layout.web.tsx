import { Stack, router } from "expo-router";
import { StyleSheet, View } from "react-native";

import { space } from "@tournaments-manager/design-tokens";

import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { useSession } from "@/shared/session/session-provider";
import { NavigationHeaderButton, Text } from "@/shared/ui";
import { NotificationHeaderButton } from "@/features/notifications/header-button";

export default function AccountLayout() {
  const t = getTranslator();
  const { colors } = usePreferences();
  const { user } = useSession();
  const goBackToAccount = () => {
    // Tras recargar una ruta de Cuenta, el historial del navegador puede
    // pertenecer a otra tab. El cierre de este flujo siempre vuelve a su raíz.
    router.replace("/account");
  };
  const headerOptions = {
    headerBackButtonDisplayMode: "minimal" as const,
    headerBackVisible: false,
    headerLeft: () => (
      <NavigationHeaderButton
        accessibilityLabel={t("common_back")}
        icon="back"
        nativeIcon={{ android: "arrow_back", ios: "chevron.left", web: "arrow_back" }}
        onPress={goBackToAccount}
      />
    ),
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
          headerLeft: () => (
            <View style={styles.username}>
              <Text variant="title">{user?.username}</Text>
            </View>
          ),
          headerRight: () => (
            <View style={styles.headerActions}>
              {user ? <NotificationHeaderButton /> : null}
              <NavigationHeaderButton
                accessibilityLabel={t("account_settings_accessibility_label")}
                icon="settings"
                nativeIcon="gearshape"
                onPress={() => router.push("/account/settings")}
                side="right"
              />
            </View>
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
    </Stack>
  );
}

const styles = StyleSheet.create({
  headerActions: { alignItems: "center", flexDirection: "row", gap: space[3] },
  username: { marginLeft: space[5] },
});
