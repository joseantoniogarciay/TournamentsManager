import { NativeTabs } from "expo-router/unstable-native-tabs";

import { color } from "@tournaments-manager/design-tokens";

import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { useSession } from "@/shared/session/session-provider";

export default function TabLayout() {
  const t = getTranslator();
  const { resolvedTheme } = usePreferences();
  const { revision } = useSession();

  return (
    <NativeTabs
      key={revision}
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
      </NativeTabs.Trigger>
    </NativeTabs>
  );
}
