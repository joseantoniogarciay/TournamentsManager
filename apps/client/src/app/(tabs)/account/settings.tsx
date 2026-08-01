import { Pressable, ScrollView, StyleSheet, Switch, View } from "react-native";

import { color, control, radius, space } from "@tournaments-manager/design-tokens";

import { getTranslator } from "@/shared/i18n/locale";
import { type ThemePreference, usePreferences } from "@/shared/preferences/preferences-provider";
import { Card, Screen, Text, useTabContentBottomPadding } from "@/shared/ui";

const themeOptions: ThemePreference[] = ["system", "light", "dark"];

export default function AccountSettingsScreen() {
  const t = getTranslator();
  const { colors, setThemePreference, themePreference } = usePreferences();
  const tabContentBottomPadding = useTabContentBottomPadding();

  return (
    <Screen bottomInset="none" topInset="navigation-bar">
      <ScrollView
        contentContainerStyle={[styles.content, { paddingBottom: tabContentBottomPadding }]}
        showsVerticalScrollIndicator={false}
      >
        <Card>
          <View style={styles.section}>
            <Text variant="title">{t("settings_appearance_title")}</Text>
            <Text color="secondary">{t("settings_theme_description")}</Text>
            <View style={styles.options}>
              {themeOptions.map((theme) => {
                const selected = themePreference === theme;
                return (
                  <Pressable
                    accessibilityRole="radio"
                    accessibilityState={{ checked: selected }}
                    key={theme}
                    onPress={() => setThemePreference(theme)}
                    style={[
                      styles.option,
                      { borderColor: selected ? color.brand.primary : colors.border.default },
                    ]}
                  >
                    <Text color={selected ? "primary" : "secondary"}>
                      {t(`settings_theme_${theme}`)}
                    </Text>
                  </Pressable>
                );
              })}
            </View>
          </View>
        </Card>

        <Card>
          <View style={styles.notificationRow}>
            <View style={styles.notificationCopy}>
              <Text variant="title">{t("settings_notifications_title")}</Text>
              <Text color="secondary">{t("settings_notifications_unavailable")}</Text>
            </View>
            <Switch accessibilityLabel={t("settings_notifications_title")} disabled value={false} />
          </View>
        </Card>
      </ScrollView>
    </Screen>
  );
}

const styles = StyleSheet.create({
  content: { gap: space[5] },
  section: { gap: space[3] },
  options: { gap: space[2] },
  option: {
    borderRadius: radius.control,
    borderWidth: 1,
    justifyContent: "center",
    minHeight: control.minHeight,
    paddingHorizontal: control.horizontalPadding,
  },
  notificationRow: { alignItems: "center", flexDirection: "row", gap: space[4] },
  notificationCopy: { flex: 1, gap: space[1] },
});
