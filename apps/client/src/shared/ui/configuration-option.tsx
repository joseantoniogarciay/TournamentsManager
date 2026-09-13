import { Pressable, StyleSheet } from "react-native";

import { color, control, radius } from "@tournaments-manager/design-tokens";

import { usePreferences } from "@/shared/preferences/preferences-provider";
import { Text } from "./text";

export function ConfigurationOption({
  disabled = false,
  label,
  onPress,
  selected,
}: {
  disabled?: boolean;
  label: string;
  onPress: () => void;
  selected: boolean;
}) {
  const { colors } = usePreferences();

  return (
    <Pressable
      accessibilityRole="button"
      accessibilityState={{ disabled, selected }}
      disabled={disabled}
      onPress={onPress}
      style={[
        styles.option,
        selected
          ? { backgroundColor: color.brand.primary, borderColor: color.brand.primary }
          : { backgroundColor: colors.surface.default, borderColor: colors.border.default },
        disabled ? styles.disabled : undefined,
      ]}
    >
      <Text color={selected ? "onBrand" : "primary"}>{label}</Text>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  disabled: { opacity: 0.55 },
  option: {
    alignItems: "center",
    borderRadius: radius.pill,
    borderWidth: 1,
    justifyContent: "center",
    minHeight: control.minHeight,
    paddingHorizontal: control.horizontalPadding,
  },
});
