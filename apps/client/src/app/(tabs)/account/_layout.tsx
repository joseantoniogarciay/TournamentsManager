import { Stack, router } from "expo-router";
import { StyleSheet, View } from "react-native";

import { space, typography } from "@tournaments-manager/design-tokens";

import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { useSession } from "@/shared/session/session-provider";
import { NavigationHeaderButton, Text, usesLiquidGlassNavigation } from "@/shared/ui";
import { NativeNotificationHeaderButton } from "@/features/notifications/native-header-button";

export default function AccountLayout() {
  const t = getTranslator();
  const { colors } = usePreferences();
  const { user } = useSession();
  const goBackToAccount = () => {
    if (router.canGoBack()) {
      router.back();
      return;
    }
    router.replace("/account");
  };
  const goBackToAccess = () => {
    if (router.canGoBack()) {
      router.back();
      return;
    }
    router.replace("/account/access");
  };
  return (
    <Stack
      screenOptions={{
        headerShadowVisible: false,
        headerBackButtonDisplayMode: "minimal",
        headerBackVisible: usesLiquidGlassNavigation,
        ...(!usesLiquidGlassNavigation
          ? {
              headerLeft: () => (
                <NavigationHeaderButton
                  accessibilityLabel={t("common_back")}
                  icon="back"
                  nativeIcon={{ android: "arrow_back", ios: "chevron.left", web: "arrow_back" }}
                  onPress={goBackToAccount}
                />
              ),
            }
          : {}),
        headerStyle: { backgroundColor: colors.surface.canvas },
        headerTintColor: colors.text.primary,
        headerTitleAlign: "center",
        headerTitleStyle: { color: colors.text.primary, fontFamily: typography.family.semibold },
      }}
    >
      <Stack.Screen
        name="index"
        options={{
          headerBackVisible: false,
          headerLeft: usesLiquidGlassNavigation
            ? undefined
            : () => <Text variant="title">{user?.username}</Text>,
          headerTitle: () => null,
          title: "",
          ...(!usesLiquidGlassNavigation
            ? {
                headerRight: () => (
                  <View style={styles.headerActions}>
                    {user ? <NativeNotificationHeaderButton /> : null}
                    <NavigationHeaderButton
                      accessibilityLabel={t("account_settings_accessibility_label")}
                      icon="settings"
                      nativeIcon={{ android: "settings", ios: "gearshape", web: "settings" }}
                      onPress={() => router.push("/(account-modals)/account/settings")}
                      side="right"
                    />
                  </View>
                ),
              }
            : {}),
        }}
      >
        {usesLiquidGlassNavigation ? (
          <Stack.Toolbar placement="left">
            <Stack.Toolbar.View hidesSharedBackground>
              <Text variant="title">{user?.username}</Text>
            </Stack.Toolbar.View>
          </Stack.Toolbar>
        ) : null}
        {usesLiquidGlassNavigation ? (
          <Stack.Toolbar placement="right">
            {user ? (
              <Stack.Toolbar.Button
                accessibilityLabel={t("account_notifications_accessibility_label")}
                icon="bell"
                onPress={() => router.push("/(account-modals)/account/notifications")}
              />
            ) : null}
            <Stack.Toolbar.Button
              accessibilityLabel={t("account_settings_accessibility_label")}
              icon="gearshape"
              onPress={() => router.push("/(account-modals)/account/settings")}
            />
          </Stack.Toolbar>
        ) : null}
      </Stack.Screen>
      <Stack.Screen name="access" options={{ title: t("account_access_data_title") }} />
      <Stack.Screen name="register" options={{ title: t("account_register_title") }} />
      <Stack.Screen name="forgot-password" options={{ title: t("password_recovery_title") }} />
      <Stack.Screen
        name="password"
        options={{
          headerBackVisible: false,
          headerLeft: () => (
            <NavigationHeaderButton
              accessibilityLabel={t("common_back")}
              icon="back"
              nativeIcon={{ android: "arrow_back", ios: "chevron.left", web: "arrow_back" }}
              onPress={goBackToAccess}
            />
          ),
          title: t("account_password_change_title"),
        }}
      />
      <Stack.Screen
        name="google-link"
        options={{ presentation: "modal", title: t("account_google_link_title") }}
      />
    </Stack>
  );
}

const styles = StyleSheet.create({
  headerActions: { alignItems: "center", flexDirection: "row", gap: space[3] },
});
