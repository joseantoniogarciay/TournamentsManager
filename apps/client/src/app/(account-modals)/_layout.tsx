import { router, Stack } from "expo-router";
import { Platform } from "react-native";

import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { NavigationHeaderButton } from "@/shared/ui";

export default function AccountModalLayout() {
  const t = getTranslator();
  const { colors } = usePreferences();
  const closeToAccount = () => router.replace("/account");

  return (
    <Stack
      screenOptions={{
        headerShadowVisible: false,
        headerStyle: { backgroundColor: colors.surface.canvas },
        headerTintColor: colors.text.primary,
        headerTitleAlign: "center",
        headerTitleStyle: { color: colors.text.primary },
        ...(Platform.OS !== "ios"
          ? {
              headerBackVisible: false,
              headerLeft: () => (
                <NavigationHeaderButton
                  accessibilityLabel={t("common_close")}
                  icon="close"
                  nativeIcon={{ android: "close", ios: "xmark", web: "close" }}
                  onPress={closeToAccount}
                />
              ),
            }
          : {}),
      }}
    >
      <Stack.Screen name="account/settings" options={{ title: t("account_settings_title") }}>
        <Stack.Toolbar placement="left">
          <Stack.Toolbar.Button
            accessibilityLabel={t("common_close")}
            icon="xmark"
            onPress={closeToAccount}
          />
        </Stack.Toolbar>
      </Stack.Screen>
      <Stack.Screen name="account/notifications" options={{ title: t("notifications_title") }}>
        <Stack.Toolbar placement="left">
          <Stack.Toolbar.Button
            accessibilityLabel={t("common_close")}
            icon="xmark"
            onPress={closeToAccount}
          />
        </Stack.Toolbar>
      </Stack.Screen>
    </Stack>
  );
}
