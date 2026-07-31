import { DarkTheme, DefaultTheme, router, Stack, ThemeProvider } from "expo-router";
import * as SplashScreen from "expo-splash-screen";
import { type PropsWithChildren, useEffect } from "react";
import { Platform } from "react-native";

import { FeedbackProvider } from "@/shared/feedback/feedback-provider";
import { PendingVerificationProvider } from "@/features/registration/pending-verification";
import { SessionProvider, useSession } from "@/shared/session/session-provider";
import { PreferencesProvider, usePreferences } from "@/shared/preferences/preferences-provider";

if (Platform.OS !== "web") {
  SplashScreen.setOptions({ duration: 240, fade: true });
  void SplashScreen.preventAutoHideAsync();
}

export default function RootLayout() {
  return (
    <PreferencesProvider>
      <NavigationTheme>
        <SessionProvider>
          <PendingVerificationProvider>
            <FeedbackProvider>
              <RootNavigator />
            </FeedbackProvider>
          </PendingVerificationProvider>
        </SessionProvider>
      </NavigationTheme>
    </PreferencesProvider>
  );
}

function RootNavigator() {
  const { finishSessionReplacement, revision, transition } = useSession();

  useEffect(() => {
    if (transition !== "resetting") return;

    router.replace("/");
    let secondFrame: number | undefined;
    const firstFrame = requestAnimationFrame(() => {
      secondFrame = requestAnimationFrame(finishSessionReplacement);
    });
    return () => {
      cancelAnimationFrame(firstFrame);
      if (secondFrame !== undefined) cancelAnimationFrame(secondFrame);
    };
  }, [finishSessionReplacement, revision, transition]);

  return (
    <Stack key={revision} screenOptions={{ headerShown: false }}>
      <Stack.Screen name="(tabs)" />
      <Stack.Screen
        name="link/confirm"
        options={{
          animation: "fade",
          headerShown: false,
          presentation: Platform.OS === "web" ? "card" : "modal",
        }}
      />
    </Stack>
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
