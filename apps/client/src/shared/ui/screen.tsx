import { type PropsWithChildren } from "react";
import { StyleSheet, View } from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";

import { space } from "@tournaments-manager/design-tokens";

import { usePreferences } from "@/shared/preferences/preferences-provider";

type ScreenProps = PropsWithChildren<{
  bottomInset?: "safe-area" | "none";
  topInset?: "safe-area" | "navigation-bar";
}>;

export function Screen({
  children,
  bottomInset = "safe-area",
  topInset = "safe-area",
}: ScreenProps) {
  const { colors } = usePreferences();
  const insets = useSafeAreaInsets();
  return (
    <View
      style={[
        styles.screen,
        {
          backgroundColor: colors.surface.canvas,
          paddingBottom: bottomInset === "safe-area" ? insets.bottom + space[4] : 0,
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
