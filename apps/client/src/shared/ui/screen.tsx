import { type PropsWithChildren } from "react";
import { SafeAreaView, StyleSheet } from "react-native";

import { color, space } from "@tournaments-manager/design-tokens";

export function Screen({ children }: PropsWithChildren) {
  return <SafeAreaView style={styles.screen}>{children}</SafeAreaView>;
}

const styles = StyleSheet.create({
  screen: { backgroundColor: color.surface.canvas, flex: 1, padding: space[4] },
});
