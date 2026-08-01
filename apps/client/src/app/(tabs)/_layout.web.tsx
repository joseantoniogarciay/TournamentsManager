import { Tabs } from "expo-router";

import { color, control } from "@tournaments-manager/design-tokens";

import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { WebIcon } from "@/shared/ui/web-icon";

export default function TabLayout() {
  const t = getTranslator();
  const { colors } = usePreferences();

  return (
    <Tabs
      screenOptions={{
        headerShown: false,
        tabBarActiveTintColor: color.brand.primary,
        tabBarInactiveTintColor: colors.text.secondary,
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
      />
      <Tabs.Screen
        name="tournaments"
        options={{
          tabBarIcon: ({ color }) => (
            <WebIcon color={color as string} name="tournament" size={control.iconSize} />
          ),
          title: t("nav_tournaments"),
        }}
      />
      <Tabs.Screen
        name="account"
        options={{
          tabBarIcon: ({ color }) => (
            <WebIcon color={color as string} name="account" size={control.iconSize} />
          ),
          title: t("nav_account"),
        }}
      />
    </Tabs>
  );
}
