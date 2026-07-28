import { createContext, type PropsWithChildren, useCallback, useContext, useState } from "react";
import { Pressable, StyleSheet, View } from "react-native";

import { banner, color, space } from "@tournaments-manager/design-tokens";

import { Text } from "@/shared/ui";

type Feedback = { message: string; kind: "network-error" | "generic-error" | "success" };
type FeedbackContextValue = { show: (feedback: Feedback) => void };
const FeedbackContext = createContext<FeedbackContextValue | null>(null);

export function FeedbackProvider({ children }: PropsWithChildren) {
  const [feedback, setFeedback] = useState<Feedback | null>(null);
  const show = useCallback((next: Feedback) => {
    setFeedback(next);
    setTimeout(() => setFeedback(null), banner.autoDismissMs);
  }, []);
  return (
    <FeedbackContext.Provider value={{ show }}>
      <View style={styles.root}>
        {feedback ? (
          <Pressable
            accessibilityRole="alert"
            onPress={() => setFeedback(null)}
            style={[styles.banner, feedback.kind === "success" ? styles.success : styles.error]}
          >
            <Text color="inverse">{feedback.message}</Text>
          </Pressable>
        ) : null}
        {children}
      </View>
    </FeedbackContext.Provider>
  );
}

export function useFeedback() {
  const value = useContext(FeedbackContext);
  if (!value) throw new Error("useFeedback debe usarse dentro de FeedbackProvider");
  return value;
}

const styles = StyleSheet.create({
  root: { flex: 1 },
  banner: {
    left: space[4],
    padding: space[3],
    position: "absolute",
    right: space[4],
    top: space[4],
    zIndex: 1,
  },
  error: { backgroundColor: color.feedback.error },
  success: { backgroundColor: color.feedback.success },
});
