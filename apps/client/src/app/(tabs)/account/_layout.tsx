import { Stack, router } from "expo-router";

import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { useSession } from "@/shared/session/session-provider";
import { Text } from "@/shared/ui";

export default function AccountLayout() {
  const t = getTranslator();
  const { colors } = usePreferences();
  const { user } = useSession();
  return (
    <Stack
      screenOptions={{
        headerShadowVisible: false,
        headerBackButtonDisplayMode: "minimal",
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
          <Stack.Toolbar.Button
            accessibilityLabel={t("account_settings_accessibility_label")}
            icon="gearshape"
            onPress={() => router.push("/account/settings")}
          />
        </Stack.Toolbar>
      </Stack.Screen>
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
        options={{ presentation: "modal", title: t("account_settings_title") }}
      >
        <Stack.Toolbar placement="left">
          <Stack.Toolbar.Button
            accessibilityLabel={t("common_close")}
            icon="xmark"
            onPress={() => {
              if (router.canGoBack()) {
                router.back();
                return;
              }
              router.replace("/account");
            }}
          />
        </Stack.Toolbar>
      </Stack.Screen>
    </Stack>
  );
}
