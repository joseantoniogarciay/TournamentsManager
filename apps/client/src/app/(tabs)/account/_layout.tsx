import { Stack, router } from "expo-router";
import { Platform } from "react-native";

import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { useSession } from "@/shared/session/session-provider";
import { NavigationHeaderButton, Text } from "@/shared/ui";
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
  const goBackToAccess = () => router.replace("/account/access");
  return (
    <Stack
      screenOptions={{
        headerShadowVisible: false,
        headerBackButtonDisplayMode: "minimal",
        headerBackVisible: Platform.OS === "ios",
        ...(Platform.OS !== "ios"
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
        headerTitleStyle: { color: colors.text.primary },
      }}
    >
      <Stack.Screen name="index" options={{ title: "" }}>
        <Stack.Toolbar placement="left">
          <Stack.Toolbar.View hidesSharedBackground>
            <Text variant="title">{user?.username}</Text>
          </Stack.Toolbar.View>
        </Stack.Toolbar>
        <Stack.Toolbar placement="right">
          {user ? (
            <Stack.Toolbar.View hidesSharedBackground>
              <NativeNotificationHeaderButton />
            </Stack.Toolbar.View>
          ) : null}
          <Stack.Toolbar.Button
            accessibilityLabel={t("account_settings_accessibility_label")}
            icon="gearshape"
            onPress={() => router.push("/(account-modals)/account/settings")}
          />
        </Stack.Toolbar>
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
