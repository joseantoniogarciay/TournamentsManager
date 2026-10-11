import { type PropsWithChildren } from "react";
import { Platform, StyleSheet, View } from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";

import { space, webLayout } from "@tournaments-manager/design-tokens";

import { FeedbackBanner } from "@/shared/feedback/feedback-provider";
import { usePreferences } from "@/shared/preferences/preferences-provider";

import { AdaptiveNavigationTitle } from "./adaptive-navigation-title";
import { ConfirmationDialogHost } from "./confirmation-dialog";

type ScreenProps = PropsWithChildren<{
  navigationTitle?: string;
  bottomInset?: "safe-area" | "none";
  topInset?: "safe-area" | "navigation-bar";
  webWidth?: "reading" | "wide";
}>;

export function Screen({
  children,
  navigationTitle,
  bottomInset = "safe-area",
  topInset = "safe-area",
  webWidth = "reading",
}: ScreenProps) {
  const { colors } = usePreferences();
  const insets = useSafeAreaInsets();
  const content = (
    <>
      {navigationTitle ? <AdaptiveNavigationTitle title={navigationTitle} /> : null}
      {children}
    </>
  );
  return (
    <View
      style={[
        styles.screen,
        {
          backgroundColor: colors.surface.canvas,
          paddingBottom:
            bottomInset === "safe-area"
              ? (Platform.OS === "web" ? 0 : insets.bottom) + space[4]
              : 0,
          paddingTop: (topInset === "safe-area" ? insets.top : 0) + space[3],
        },
      ]}
    >
      {Platform.OS === "web" ? (
        <View
          style={[
            styles.webContent,
            { maxWidth: webWidth === "wide" ? webLayout.maxWidth : webLayout.readingMaxWidth },
          ]}
        >
          {content}
        </View>
      ) : (
        content
      )}
      {Platform.OS === "web" ? <FeedbackBanner /> : null}
      <ConfirmationDialogHost />
    </View>
  );
}

const styles = StyleSheet.create({
  screen: {
    flex: 1,
  },
  webContent: { alignSelf: "center", flex: 1, minWidth: 0, width: "100%" },
});
