import { type PropsWithChildren } from "react";
import { StyleSheet, View } from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";

import { space } from "@tournaments-manager/design-tokens";

import { usePreferences } from "@/shared/preferences/preferences-provider";

type ScreenProps = PropsWithChildren<{
  topInset?: "safe-area" | "navigation-bar";
}>;

export function Screen({ children, topInset = "safe-area" }: ScreenProps) {
  const { colors } = usePreferences();
  const insets = useSafeAreaInsets();
  return (
    <View
      style={[
        styles.screen,
        {
          backgroundColor: colors.surface.canvas,
          paddingBottom: insets.bottom + space[4],
          paddingTop: (topInset === "safe-area" ? insets.top : 0) + space[3],
        },
      ]}
    >
      {children}
    </View>
  );
}

const styles = StyleSheet.create({
  screen: {
    flex: 1,
  },
});
