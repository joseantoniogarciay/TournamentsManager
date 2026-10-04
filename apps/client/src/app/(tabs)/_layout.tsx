import { NativeTabs } from "expo-router/unstable-native-tabs";
import { Platform } from "react-native";

import { color, typography } from "@tournaments-manager/design-tokens";

import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { useSession } from "@/shared/session/session-provider";
import { useNotifications } from "@/features/notifications/notification-provider";

export default function TabLayout() {
  const t = getTranslator();
  const { colors, resolvedTheme } = usePreferences();
  const { revision } = useSession();
  const { count } = useNotifications();
  const usesAndroidAppearance = Platform.OS === "android";
  const usesLegacyIOSAppearance = Platform.OS === "ios" && Number(Platform.Version) < 26;

  return (
    <NativeTabs
      key={revision}
      labelStyle={{
        default: {
          fontFamily: typography.family.medium,
          color: usesAndroidAppearance ? colors.text.secondary : undefined,
        },
        selected: {
          fontFamily: typography.family.semibold,
          color: usesAndroidAppearance ? color.brand.primary : undefined,
        },
      }}
      backgroundColor={
        usesLegacyIOSAppearance || usesAndroidAppearance ? colors.surface.default : undefined
      }
      iconColor={
        usesAndroidAppearance
          ? { default: colors.text.secondary, selected: color.brand.primary }
          : undefined
      }
      indicatorColor={usesAndroidAppearance ? colors.surface.subtle : undefined}
      blurEffect={usesLegacyIOSAppearance ? "none" : undefined}
      disableTransparentOnScrollEdge={usesLegacyIOSAppearance}
      minimizeBehavior="never"
      tintColor={color.brand.primary}
      unstable_nativeProps={{ colorScheme: resolvedTheme }}
    >
      <NativeTabs.Trigger name="index">
        <NativeTabs.Trigger.Label>{t("nav_home")}</NativeTabs.Trigger.Label>
        <NativeTabs.Trigger.Icon
          md={{ default: "home", selected: "home_filled" }}
          sf={{ default: "house", selected: "house.fill" }}
        />
      </NativeTabs.Trigger>
      <NativeTabs.Trigger name="tournaments">
        <NativeTabs.Trigger.Label>{t("nav_tournaments")}</NativeTabs.Trigger.Label>
        <NativeTabs.Trigger.Icon
          md={{ default: "emoji_events", selected: "emoji_events" }}
          sf={{ default: "trophy", selected: "trophy.fill" }}
        />
      </NativeTabs.Trigger>
      <NativeTabs.Trigger name="account">
        <NativeTabs.Trigger.Label>{t("nav_account")}</NativeTabs.Trigger.Label>
        <NativeTabs.Trigger.Icon
          md={{ default: "person", selected: "person" }}
          sf={{ default: "person", selected: "person.fill" }}
        />
        {count > 0 ? <NativeTabs.Trigger.Badge>•</NativeTabs.Trigger.Badge> : null}
      </NativeTabs.Trigger>
    </NativeTabs>
  );
}
