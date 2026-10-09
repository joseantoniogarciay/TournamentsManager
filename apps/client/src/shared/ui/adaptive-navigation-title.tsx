import { Stack } from "expo-router";
import { Platform, StyleSheet, View } from "react-native";

import { space } from "@tournaments-manager/design-tokens";

import { useAdaptiveNavigationTitle } from "./use-adaptive-navigation-title";

export function AdaptiveNavigationTitle({ title }: { title: string }) {
  return Platform.OS === "ios" ? <IOSNavigationTitle title={title} /> : null;
}

function IOSNavigationTitle({ title }: { title: string }) {
  const navigationTitle = useAdaptiveNavigationTitle(title, true);
  return (
    <>
      <Stack.Screen
        options={{ title: navigationTitle.nativeTitle, headerTitle: navigationTitle.headerTitle }}
      />
      {navigationTitle.measurement}
      {navigationTitle.contentTitle ? (
        <View style={styles.contentTitle}>{navigationTitle.contentTitle}</View>
      ) : null}
    </>
  );
}

const styles = StyleSheet.create({ contentTitle: { paddingBottom: space[3] } });
