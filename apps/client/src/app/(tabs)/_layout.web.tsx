import { router, Tabs, usePathname } from "expo-router";
import { StyleSheet, View } from "react-native";

import { color, control, typography } from "@tournaments-manager/design-tokens";

import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { WebIcon } from "@/shared/ui/web-icon";
import { useNotifications } from "@/features/notifications/notification-provider";

export default function TabLayout() {
  const t = getTranslator();
  const { colors } = usePreferences();
  const { count } = useNotifications();
  const pathname = usePathname();

  return (
    <Tabs
      screenOptions={{
        headerShown: false,
        tabBarActiveTintColor: color.brand.primary,
        tabBarInactiveTintColor: colors.text.secondary,
        tabBarLabelStyle: { fontFamily: typography.family.medium },
        tabBarStyle: {
          backgroundColor: colors.surface.default,
          borderTopColor: colors.border.default,
        },
      }}
    >
      <Tabs.Screen
        name="index"
        options={{
          tabBarIcon: ({ color }) => (
            <WebIcon color={color as string} name="home" size={control.iconSize} />
          ),
          title: t("nav_home"),
        }}
        listeners={{ tabPress: resetToTabRoot(pathname, "/") }}
      />
      <Tabs.Screen
        name="tournaments"
        options={{
          tabBarIcon: ({ color }) => (
            <WebIcon color={color as string} name="tournament" size={control.iconSize} />
          ),
          title: t("nav_tournaments"),
        }}
        listeners={{ tabPress: resetToTabRoot(pathname, "/tournaments") }}
      />
      <Tabs.Screen
        name="account"
        options={{
          tabBarIcon: ({ color }) => (
            <View style={styles.accountIcon}>
              <WebIcon color={color as string} name="account" size={control.iconSize} />
              {count > 0 ? <View style={styles.unreadDot} /> : null}
            </View>
          ),
          title: t("nav_account"),
        }}
        listeners={{ tabPress: resetToTabRoot(pathname, "/account") }}
      />
    </Tabs>
  );
}

const styles = StyleSheet.create({
  accountIcon: { height: control.iconSize, position: "relative", width: control.iconSize },
  unreadDot: {
    backgroundColor: color.feedback.error,
    borderRadius: 4,
    height: 8,
    position: "absolute",
    right: -4,
    top: -4,
    width: 8,
  },
});

function resetToTabRoot(pathname: string, rootPath: "/" | "/tournaments" | "/account") {
  return (event: { preventDefault: () => void }) => {
    const isActiveTab =
      pathname === rootPath || (rootPath !== "/" && pathname.startsWith(`${rootPath}/`));

    if (!isActiveTab) return;

    event.preventDefault();
    router.replace(rootPath);
  };
}
