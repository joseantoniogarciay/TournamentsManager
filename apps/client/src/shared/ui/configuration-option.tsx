import { StyleSheet } from "react-native";

import { color, control, radius } from "@tournaments-manager/design-tokens";

import { usePreferences } from "@/shared/preferences/preferences-provider";
import { Text } from "./text";
import { ActionPressable, type ActionAvailability } from "./action-pressable";

export function ConfigurationOption({
  disabled = false,
  disabledReason,
  label,
  onPress,
  selected,
}: ActionAvailability & {
  label: string;
  onPress: () => void;
  selected: boolean;
}) {
  const { colors } = usePreferences();

  return (
    <ActionPressable
      accessibilityRole="button"
      accessibilityState={{ disabled, selected }}
      disabled={disabled}
      disabledReason={disabledReason ?? ""}
      onPress={onPress}
      style={[
        styles.option,
        disabled
          ? { backgroundColor: colors.surface.subtle, borderColor: colors.border.default }
          : selected
            ? { backgroundColor: color.brand.primary, borderColor: color.brand.primary }
            : { backgroundColor: colors.surface.default, borderColor: colors.border.default },
      ]}
    >
      <Text color={disabled ? "secondary" : selected ? "onBrand" : "primary"}>{label}</Text>
    </ActionPressable>
  );
}

const styles = StyleSheet.create({
  option: {
    alignItems: "center",
    borderRadius: radius.pill,
    borderWidth: 1,
    justifyContent: "center",
    minHeight: control.minHeight,
    maxWidth: "100%",
    paddingHorizontal: control.horizontalPadding,
  },
});
