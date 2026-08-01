import { type PropsWithChildren } from "react";
import { type StyleProp, StyleSheet, type ViewStyle, View } from "react-native";

import { radius, space } from "@tournaments-manager/design-tokens";

import { usePreferences } from "@/shared/preferences/preferences-provider";

type CardProps = PropsWithChildren<{ style?: StyleProp<ViewStyle> }>;

export function Card({ children, style }: CardProps) {
  const { colors } = usePreferences();
  return (
    <View
      style={[
        styles.card,
        { backgroundColor: colors.surface.default, borderColor: colors.border.default },
        style,
      ]}
    >
      {children}
    </View>
  );
}

const styles = StyleSheet.create({
  card: {
    borderRadius: radius.card,
    borderWidth: 1,
    marginHorizontal: space[5],
    padding: space[5],
  },
});
