import { router } from "expo-router";
import { Pressable, StyleSheet, View } from "react-native";
import { color, control, space } from "@tournaments-manager/design-tokens";
import { useNotifications } from "./notification-provider";
import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { Text } from "@/shared/ui";
import { WebIcon } from "@/shared/ui/web-icon";

export function NotificationHeaderButton() {
  const t = getTranslator();
  const { colors } = usePreferences();
  const { count } = useNotifications();
  return (
    <Pressable
      accessibilityLabel={t("account_notifications_accessibility_label")}
      accessibilityRole="button"
      onPress={() => router.push("/account/notifications" as never)}
      style={styles.button}
    >
      <WebIcon color={colors.text.primary} name="bell" size={control.iconSize} />
      {count > 0 ? (
        <View style={[styles.badge, { backgroundColor: color.brand.primary }]}>
          <Text style={styles.badgeText}>{count > 99 ? "99+" : count}</Text>
        </View>
      ) : null}
    </Pressable>
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
    borderRadius: 10,
    minWidth: 20,
    paddingHorizontal: space[1],
    position: "absolute",
    right: 0,
    top: 0,
  },
  badgeText: { color: "white", fontSize: 11 },
});
