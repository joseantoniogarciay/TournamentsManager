import { router } from "expo-router";
import { SymbolView } from "expo-symbols";
import { Platform, Pressable, StyleSheet, View } from "react-native";

import { color, control, space } from "@tournaments-manager/design-tokens";

import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { NavigationHeaderButton, Text } from "@/shared/ui";

import { useNotifications } from "./notification-provider";

/** Control de cabecera nativo con contador, sin depender de permisos push. */
export function NativeNotificationHeaderButton() {
  const t = getTranslator();
  const { colors } = usePreferences();
  const { count } = useNotifications();

  const badge =
    count > 0 ? (
      <View style={styles.badge}>
        <Text style={styles.badgeText}>{count > 99 ? "99+" : count}</Text>
      </View>
    ) : null;

  if (Platform.OS !== "ios") {
    return (
      <NavigationHeaderButton
        accessibilityLabel={t("account_notifications_accessibility_label")}
        badge={badge}
        icon="bell"
        nativeIcon={{ android: "notifications", ios: "bell", web: "notifications" }}
        onPress={() => router.push("/account/notifications" as never)}
      />
    );
  }

  return (
    <Pressable
      accessibilityLabel={t("account_notifications_accessibility_label")}
      accessibilityRole="button"
      onPress={() => router.push("/account/notifications" as never)}
      style={styles.button}
    >
      <SymbolView name="bell" size={control.iconSize} tintColor={colors.text.primary} />
      {badge}
    </Pressable>
  );
}

const styles = StyleSheet.create({
  badge: {
    alignItems: "center",
    backgroundColor: color.feedback.error,
    borderRadius: 10,
    minWidth: 20,
    paddingHorizontal: space[1],
    position: "absolute",
    right: 0,
    top: 0,
  },
  badgeText: { color: "white", fontSize: 11 },
  button: {
    alignItems: "center",
    height: control.minHeight,
    justifyContent: "center",
    width: control.minHeight,
  },
});
