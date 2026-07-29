import { type PropsWithChildren } from "react";
import { StyleSheet, View } from "react-native";

import { radius, space } from "@tournaments-manager/design-tokens";

import { usePreferences } from "@/shared/preferences/preferences-provider";

export function Card({ children }: PropsWithChildren) {
  const { colors } = usePreferences();
  return (
    <View
      style={[
        styles.card,
        { backgroundColor: colors.surface.default, borderColor: colors.border.default },
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
