import { type PropsWithChildren } from "react";
import { StyleSheet, View } from "react-native";

import { color, radius, space } from "@tournaments-manager/design-tokens";

export function Card({ children }: PropsWithChildren) {
  return <View style={styles.card}>{children}</View>;
}

const styles = StyleSheet.create({
  card: {
    backgroundColor: color.surface.default,
    borderColor: color.border.default,
    borderRadius: radius.card,
    borderWidth: 1,
    marginHorizontal: space[5],
    padding: space[5],
  },
});
