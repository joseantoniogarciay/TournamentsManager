import { type PropsWithChildren } from "react";
import { type StyleProp, StyleSheet, type ViewStyle, View } from "react-native";

import { radius, space } from "@tournaments-manager/design-tokens";

import { usePreferences } from "@/shared/preferences/preferences-provider";

type CardProps = PropsWithChildren<{
  density?: "compact" | "standard";
  style?: StyleProp<ViewStyle>;
}>;

export function Card({ children, density = "standard", style }: CardProps) {
  const { colors } = usePreferences();
  return (
    <View
      style={[
        styles[density],
        { backgroundColor: colors.surface.default, borderColor: colors.border.default },
        style,
      ]}
    >
      {children}
    </View>
  );
}

const styles = StyleSheet.create({
  compact: {
    borderRadius: radius.card,
    borderWidth: 1,
    marginHorizontal: space[5],
    padding: space[3],
  },
  standard: {
    borderRadius: radius.card,
    borderWidth: 1,
    marginHorizontal: space[5],
    padding: space[5],
  },
});
