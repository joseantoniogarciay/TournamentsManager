import { router } from "expo-router";
import { StyleSheet, View } from "react-native";
import { color, control, space } from "@tournaments-manager/design-tokens";
import { useNotifications } from "./notification-provider";
import { getTranslator } from "@/shared/i18n/locale";
import { NavigationHeaderButton, Text } from "@/shared/ui";

export function NotificationHeaderButton() {
  const t = getTranslator();
  const { count } = useNotifications();
  return (
    <NavigationHeaderButton
      accessibilityLabel={t("account_notifications_accessibility_label")}
      badge={
        count > 0 ? (
          <View style={styles.badge}>
            <Text style={styles.badgeText}>{count > 99 ? "99+" : count}</Text>
          </View>
        ) : null
      }
      icon="bell"
      nativeIcon="bell"
      onPress={() => router.push("/(account-modals)/account/notifications")}
    />
  );
}
const styles = StyleSheet.create({
  button: {
    alignItems: "center",
    height: control.minHeight,
    justifyContent: "center",
    width: control.minHeight,
  },
  badge: {
    alignItems: "center",
    backgroundColor: color.feedback.error,
    borderRadius: control.minHeight / 2,
    minHeight: 18,
    minWidth: 20,
    paddingHorizontal: space[1],
    position: "absolute",
    right: -space[1],
    top: -space[1],
  },
  badgeText: { color: "white", fontSize: 11 },
});
