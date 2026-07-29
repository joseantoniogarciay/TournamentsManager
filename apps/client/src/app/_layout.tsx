import { DarkTheme, DefaultTheme, Stack, ThemeProvider } from "expo-router";
import * as SplashScreen from "expo-splash-screen";
import { type PropsWithChildren, useEffect } from "react";
import { Platform } from "react-native";

import { FeedbackProvider } from "@/shared/feedback/feedback-provider";
import { getTranslator } from "@/shared/i18n/locale";
import { PreferencesProvider, usePreferences } from "@/shared/preferences/preferences-provider";

if (Platform.OS !== "web") {
  SplashScreen.setOptions({ duration: 240, fade: true });
  void SplashScreen.preventAutoHideAsync();
}

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

  useEffect(() => {
    if (Platform.OS !== "web") SplashScreen.hide();
  }, []);

  return (
    <ThemeProvider value={resolvedTheme === "dark" ? DarkTheme : DefaultTheme}>
      {children}
    </ThemeProvider>
  );
}
