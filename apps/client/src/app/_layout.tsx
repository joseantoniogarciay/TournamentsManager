import { Stack } from "expo-router";
import { Platform } from "react-native";

import { FeedbackProvider } from "@/shared/feedback/feedback-provider";
import { getTranslator } from "@/shared/i18n/locale";
import { PreferencesProvider } from "@/shared/preferences/preferences-provider";

export default function RootLayout() {
  const t = getTranslator();
  return (
    <PreferencesProvider>
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
    </PreferencesProvider>
  );
}
