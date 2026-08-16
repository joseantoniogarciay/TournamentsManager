import { SymbolView } from "expo-symbols";
import { type ComponentProps, type ReactNode } from "react";
import { Platform, Pressable, StyleSheet } from "react-native";

import { control, radius, space } from "@tournaments-manager/design-tokens";

import { usePreferences } from "@/shared/preferences/preferences-provider";

import { WebIcon } from "./web-icon";

type NavigationHeaderButtonProps = {
  accessibilityLabel: string;
  icon: ComponentProps<typeof WebIcon>["name"];
  nativeIcon: ComponentProps<typeof SymbolView>["name"];
  onPress: () => void;
  badge?: ReactNode;
  side?: "left" | "right";
};

export const usesLiquidGlassNavigation = Platform.OS === "ios" && Number(Platform.Version) >= 26;

export function NavigationHeaderButton({
  accessibilityLabel,
  badge,
  icon,
  nativeIcon,
  onPress,
  side = "left",
}: NavigationHeaderButtonProps) {
  const { colors } = usePreferences();

  return (
    <Pressable
      accessibilityLabel={accessibilityLabel}
      accessibilityRole="button"
      onPress={onPress}
      style={[
        styles.button,
        side === "left"
          ? Platform.OS === "ios" && !usesLiquidGlassNavigation
            ? undefined
            : styles.left
          : Platform.OS === "ios" && !usesLiquidGlassNavigation
            ? undefined
            : styles.right,
        { backgroundColor: colors.surface.default, borderColor: colors.border.default },
      ]}
    >
      {Platform.OS === "web" ? (
        <WebIcon color={colors.text.primary} name={icon} size={control.iconSize} />
      ) : (
        <SymbolView name={nativeIcon} size={control.iconSize} tintColor={colors.text.primary} />
      )}
      {badge}
    </Pressable>
  );
}

const styles = StyleSheet.create({
  button: {
    alignItems: "center",
    borderRadius: radius.pill,
    borderWidth: 1,
    height: control.minHeight,
    justifyContent: "center",
    width: control.minHeight,
  },
  left: { marginLeft: space[5] },
  right: { marginRight: space[5] },
});
