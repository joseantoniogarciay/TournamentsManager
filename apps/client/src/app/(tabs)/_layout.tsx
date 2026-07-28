import { NativeTabs } from "expo-router/unstable-native-tabs";

import { getTranslator } from "@/shared/i18n/locale";

export default function TabLayout() {
  const t = getTranslator();

  return (
    <NativeTabs minimizeBehavior="onScrollDown">
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
