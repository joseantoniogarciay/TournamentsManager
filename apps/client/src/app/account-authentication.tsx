import { router, Stack } from "expo-router";

import { AccountScreen } from "@/app/(tabs)/account";
import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { NavigationHeaderButton, usesLiquidGlassNavigation } from "@/shared/ui";

export default function AccountAuthenticationScreen() {
  const t = getTranslator();
  const { colors } = usePreferences();
  const close = () => {
    if (router.canDismiss()) {
      router.dismiss();
      return;
    }
    router.replace("/");
  };

  return (
    <>
      <Stack.Screen
        options={{
          headerBackVisible: false,
          ...(!usesLiquidGlassNavigation
            ? {
                headerLeft: () => (
                  <NavigationHeaderButton
                    accessibilityLabel={t("common_close")}
                    icon="close"
                    nativeIcon={{ android: "close", ios: "xmark", web: "close" }}
                    onPress={close}
                  />
                ),
              }
            : {}),
          headerShadowVisible: false,
          headerStyle: { backgroundColor: colors.surface.canvas },
          headerTintColor: colors.text.primary,
          headerTitle: t("account_title"),
          headerTitleAlign: "center",
        }}
      >
        {usesLiquidGlassNavigation ? (
          <Stack.Toolbar placement="left">
            <Stack.Toolbar.Button
              accessibilityLabel={t("common_close")}
              icon="xmark"
              onPress={close}
            />
          </Stack.Toolbar>
        ) : null}
      </Stack.Screen>
      <AccountScreen sessionReplacementDestination="/create-tournament" />
    </>
  );
}
