import { StyleSheet, View } from "react-native";

import { radius, space } from "@tournaments-manager/design-tokens";

import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { Text } from "@/shared/ui";

/** Identifica de forma consistente a la cuenta que creó la liga. */
export function LeagueCreatorChip() {
  const t = getTranslator();
  const { colors } = usePreferences();

  return (
    <View
      accessibilityLabel={t("league_creator")}
      accessibilityRole="text"
      style={[
        styles.chip,
        { backgroundColor: colors.surface.subtle, borderColor: colors.border.default },
      ]}
    >
      <Text variant="caption">{t("league_creator")}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  chip: {
    borderRadius: radius.pill,
    borderWidth: 1,
    paddingHorizontal: space[2],
    paddingVertical: space[1],
  },
});
