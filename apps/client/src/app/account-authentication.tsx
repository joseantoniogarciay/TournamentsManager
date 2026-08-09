import { router, Stack } from "expo-router";
import { SymbolView } from "expo-symbols";
import { Platform, Pressable, StyleSheet } from "react-native";

import { control, radius } from "@tournaments-manager/design-tokens";

import { AccountScreen } from "@/app/(tabs)/account";
import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { WebIcon } from "@/shared/ui/web-icon";

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
          headerLeft: () => (
            <Pressable
              accessibilityLabel={t("common_close")}
              accessibilityRole="button"
              onPress={close}
              style={[
                styles.navigationButton,
                { backgroundColor: colors.surface.default, borderColor: colors.border.default },
              ]}
            >
              {Platform.OS === "web" ? (
                <WebIcon color={colors.text.primary} name="close" size={control.iconSize} />
              ) : (
                <SymbolView
                  name={{ android: "close", ios: "xmark", web: "close" }}
                  size={control.iconSize}
                  tintColor={colors.text.primary}
                />
              )}
            </Pressable>
          ),
          headerShadowVisible: false,
          headerStyle: { backgroundColor: colors.surface.canvas },
          headerTintColor: colors.text.primary,
          headerTitle: t("account_title"),
          headerTitleAlign: "center",
        }}
      />
      <AccountScreen sessionReplacementDestination="/create-tournament" />
    </>
  );
}

const styles = StyleSheet.create({
  navigationButton: {
    alignItems: "center",
    borderRadius: radius.pill,
    borderWidth: 1,
    height: control.minHeight,
    justifyContent: "center",
    width: control.minHeight,
  },
});
