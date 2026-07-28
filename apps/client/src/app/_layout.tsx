import { Stack } from "expo-router";
import { Platform } from "react-native";

import { FeedbackProvider } from "@/shared/feedback/feedback-provider";

export default function RootLayout() {
  return (
    <FeedbackProvider>
      <Stack screenOptions={{ headerShown: false }}>
        <Stack.Screen name="(tabs)" />
        <Stack.Screen
          name="link/confirm"
          options={{ presentation: Platform.OS === "web" ? "card" : "modal" }}
        />
      </Stack>
    </FeedbackProvider>
  );
}
