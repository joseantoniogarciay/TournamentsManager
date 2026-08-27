import { router, Stack } from "expo-router";
import { StyleSheet, View } from "react-native";

import { color, space } from "@tournaments-manager/design-tokens";

import { getTranslator } from "@/shared/i18n/locale";
import { NavigationHeaderButton, Text, usesLiquidGlassNavigation } from "@/shared/ui";

import { useNotifications } from "./notification-provider";

/** Control de cabecera nativo con contador, sin depender de permisos push. */
export function NativeNotificationHeaderButton() {
  const t = getTranslator();
  const { count } = useNotifications();

  const badge =
    count > 0 ? (
      <View style={styles.badge}>
        <Text style={styles.badgeText}>{count > 99 ? "99+" : count}</Text>
      </View>
    ) : null;

  if (!usesLiquidGlassNavigation) {
    return (
      <NavigationHeaderButton
        accessibilityLabel={t("account_notifications_accessibility_label")}
        badge={badge}
        icon="bell"
        nativeIcon={{ android: "notifications", ios: "bell", web: "notifications" }}
        onPress={() => router.push("/(account-modals)/account/notifications")}
      />
    );
  }

  return (
    <Stack.Toolbar.Button
      accessibilityLabel={t("account_notifications_accessibility_label")}
      icon="bell"
      onPress={() => router.push("/(account-modals)/account/notifications")}
    />
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
});
