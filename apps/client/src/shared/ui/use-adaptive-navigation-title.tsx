import { useHeaderHeight } from "expo-router/react-navigation";
import { useState } from "react";
import { Platform, StyleSheet, View, useWindowDimensions } from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";

import { control, space, typography } from "@tournaments-manager/design-tokens";

import { Text } from "./text";

export function useAdaptiveNavigationTitle(title: string | undefined, emphasized = false) {
  const headerHeight = useHeaderHeight();
  const insets = useSafeAreaInsets();
  const { width, fontScale } = useWindowDimensions();
  const maxWidth = Math.max(
    0,
    width - 2 * (Math.max(insets.left, insets.right) + space[5] + control.minHeight + space[5]),
  );
  const measurementKey = JSON.stringify([title, maxWidth, fontScale, emphasized]);
  const [measurement, setMeasurement] = useState<{ key: string; height: number }>();
  // HeaderHeight includes the status-bar inset for this fullscreen route.
  const availableHeight = Math.max(0, headerHeight - insets.top);
  const reflows =
    Platform.OS === "ios" &&
    measurement?.key === measurementKey &&
    measurement.height > availableHeight;

  return {
    nativeTitle: reflows ? "" : title,
    headerTitle: () =>
      reflows ? (
        <View />
      ) : (
        <Text
          accessibilityRole="header"
          numberOfLines={2}
          style={[
            styles.title,
            emphasized && { fontFamily: typography.family.semibold },
            Platform.OS === "ios" ? { maxWidth } : { marginHorizontal: space[5] },
          ]}
          variant="bodyLarge"
        >
          {title}
        </Text>
      ),
    measurement:
      Platform.OS === "ios" && title ? (
        <View
          accessibilityElementsHidden
          importantForAccessibility="no-hide-descendants"
          pointerEvents="none"
          style={[styles.measurement, { width: maxWidth }]}
        >
          <Text
            onTextLayout={({ nativeEvent: { lines } }) => {
              const height = lines.reduce(
                (bottom, line) => Math.max(bottom, line.y + line.height),
                0,
              );
              setMeasurement((previous) =>
                previous?.key === measurementKey && previous.height === height
                  ? previous
                  : { key: measurementKey, height },
              );
            }}
            style={[styles.title, emphasized && { fontFamily: typography.family.semibold }]}
            variant="bodyLarge"
          >
            {title}
          </Text>
        </View>
      ) : null,
    contentTitle: reflows ? (
      <Text
        accessibilityRole="header"
        style={[
          styles.title,
          styles.contentTitle,
          emphasized && { fontFamily: typography.family.semibold },
        ]}
        variant="bodyLarge"
      >
        {title}
      </Text>
    ) : null,
  };
}

const styles = StyleSheet.create({
  title: { flexShrink: 1, textAlign: "center" },
  measurement: { opacity: 0, position: "absolute" },
  contentTitle: { marginHorizontal: space[5] },
});
