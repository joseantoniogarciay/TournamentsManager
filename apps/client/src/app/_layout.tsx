import { DarkTheme, DefaultTheme, Stack, ThemeProvider } from "expo-router";
import { type PropsWithChildren } from "react";
import { Platform } from "react-native";

import { FeedbackProvider } from "@/shared/feedback/feedback-provider";
import { getTranslator } from "@/shared/i18n/locale";
import { PreferencesProvider, usePreferences } from "@/shared/preferences/preferences-provider";

export default function RootLayout() {
  const t = getTranslator();
  return (
    <PreferencesProvider>
      <NavigationTheme>
        <FeedbackProvider>
          <Stack screenOptions={{ headerShown: false }}>
            <Stack.Screen name="(tabs)" />
            <Stack.Screen
              name="link/confirm"
              options={{
                headerShown: true,
                headerTitleAlign: "center",
                presentation: Platform.OS === "web" ? "card" : "modal",
                title: t("link_confirmation_title"),
              }}
            />
          </Stack>
        </FeedbackProvider>
      </NavigationTheme>
    </PreferencesProvider>
  );
}

function NavigationTheme({ children }: PropsWithChildren) {
  const { resolvedTheme } = usePreferences();

  return (
    <ThemeProvider value={resolvedTheme === "dark" ? DarkTheme : DefaultTheme}>
      {children}
    </ThemeProvider>
  );
}
