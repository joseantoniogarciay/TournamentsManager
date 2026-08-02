import { DarkTheme, DefaultTheme, router, Stack, ThemeProvider } from "expo-router";
import Head from "expo-router/head";
import * as SplashScreen from "expo-splash-screen";
import { type PropsWithChildren, useEffect } from "react";
import { Platform } from "react-native";

import { FeedbackProvider } from "@/shared/feedback/feedback-provider";
import { PendingVerificationProvider } from "@/features/registration/pending-verification";
import { SessionProvider, useSession } from "@/shared/session/session-provider";
import { PreferencesProvider, usePreferences } from "@/shared/preferences/preferences-provider";
import { ConfirmationDialogProvider } from "@/shared/ui";

if (Platform.OS !== "web") {
  SplashScreen.setOptions({ duration: 240, fade: true });
  void SplashScreen.preventAutoHideAsync();
}

export default function RootLayout() {
  return (
    <PreferencesProvider>
      <NavigationTheme>
        <FeedbackProvider>
          <SessionProvider>
            <PendingVerificationProvider>
              <ConfirmationDialogProvider>
                <RootNavigator />
              </ConfirmationDialogProvider>
            </PendingVerificationProvider>
          </SessionProvider>
        </FeedbackProvider>
      </NavigationTheme>
    </PreferencesProvider>
  );
}

function RootNavigator() {
  const { finishSessionReplacement, replacementDestination, revision, transition } = useSession();

  useEffect(() => {
    if (transition !== "resetting" && transition !== "signing-out") return;

    router.replace(transition === "signing-out" ? "/account" : replacementDestination);
    let secondFrame: number | undefined;
    const firstFrame = requestAnimationFrame(() => {
      secondFrame = requestAnimationFrame(finishSessionReplacement);
    });
    return () => {
      cancelAnimationFrame(firstFrame);
      if (secondFrame !== undefined) cancelAnimationFrame(secondFrame);
    };
  }, [finishSessionReplacement, replacementDestination, revision, transition]);

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
  const { colors, resolvedTheme } = usePreferences();

  useEffect(() => {
    if (Platform.OS !== "web") SplashScreen.hide();
  }, []);

  return (
    <ThemeProvider value={resolvedTheme === "dark" ? DarkTheme : DefaultTheme}>
      <WebPageAppearance backgroundColor={colors.surface.canvas} />
      {children}
    </ThemeProvider>
  );
}

function WebPageAppearance({ backgroundColor }: { backgroundColor: string }) {
  useEffect(() => {
    if (Platform.OS !== "web") return;

    document.documentElement.style.backgroundColor = backgroundColor;
    document.body.style.backgroundColor = backgroundColor;
  }, [backgroundColor]);

  useEffect(() => {
    if (Platform.OS !== "web" || !window.visualViewport || !isSafari()) return;

    const viewport = window.visualViewport;
    let previousHeight = viewport.height;
    let remeasureTimeout: ReturnType<typeof setTimeout> | undefined;
    const remeasureAfterKeyboardCloses = () => {
      const height = viewport.height;
      if (height > previousHeight) {
        if (remeasureTimeout) clearTimeout(remeasureTimeout);
        remeasureTimeout = setTimeout(() => {
          // Safari puede notificar una altura intermedia al terminar de ocultar
          // el teclado. Reemitir la medida fuerza a React Native Web a usar la
          // altura final, sin desplazar el contenido.
          viewport.dispatchEvent(new Event("resize"));
        }, 250);
      }
      previousHeight = height;
    };

    viewport.addEventListener("resize", remeasureAfterKeyboardCloses);
    return () => {
      viewport.removeEventListener("resize", remeasureAfterKeyboardCloses);
      if (remeasureTimeout) clearTimeout(remeasureTimeout);
    };
  }, []);

  if (Platform.OS !== "web") return null;

  return (
    <Head>
      <meta
        name="viewport"
        content="width=device-width, initial-scale=1, shrink-to-fit=no, viewport-fit=cover"
      />
      <meta name="theme-color" content={backgroundColor} />
      <style>{`
        @supports selector(div:has(> [role="tablist"])) {
          div:has(> [role="tablist"]) {
            bottom: env(safe-area-inset-bottom) !important;
            left: 0 !important;
            position: fixed !important;
            right: 0 !important;
            z-index: 1 !important;
          }
        }
      `}</style>
    </Head>
  );
}

function isSafari() {
  const userAgent = navigator.userAgent;
  return /Safari/.test(userAgent) && !/Chrome|Chromium|CriOS|EdgiOS|FxiOS/.test(userAgent);
}
