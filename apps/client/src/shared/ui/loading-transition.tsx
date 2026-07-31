import { useEffect, useRef, useState } from "react";
import { AccessibilityInfo, ActivityIndicator, Animated, StyleSheet, View } from "react-native";

import { motion, space } from "@tournaments-manager/design-tokens";

import { usePreferences } from "@/shared/preferences/preferences-provider";

import { Text } from "./text";

type LoadingTransitionProps = { active: boolean; message: string };

/** Capa modal reutilizable para transiciones que bloquean la interacción. */
export function LoadingTransition({ active, message }: LoadingTransitionProps) {
  const { colors } = usePreferences();
  const [rendered, setRendered] = useState(active);
  const [reducedMotion, setReducedMotion] = useState(false);
  const visibility = useRef(new Animated.Value(active ? 1 : 0)).current;

  useEffect(() => {
    void AccessibilityInfo.isReduceMotionEnabled().then(setReducedMotion);
    const subscription = AccessibilityInfo.addEventListener(
      "reduceMotionChanged",
      setReducedMotion,
    );
    return () => subscription.remove();
  }, []);

  useEffect(() => {
    visibility.stopAnimation();
    if (active) {
      setRendered(true);
      if (reducedMotion) {
        visibility.setValue(1);
        return;
      }
      visibility.setValue(0);
      Animated.timing(visibility, {
        duration: motion.enterExit,
        toValue: 1,
        useNativeDriver: true,
      }).start();
      return;
    }

    if (!rendered) return;
    if (reducedMotion) {
      visibility.setValue(0);
      setRendered(false);
      return;
    }
    Animated.timing(visibility, {
      duration: motion.enterExit,
      toValue: 0,
      useNativeDriver: true,
    }).start(({ finished }) => {
      if (finished) setRendered(false);
    });
  }, [active, reducedMotion, rendered, visibility]);

  if (!rendered) return null;

  return (
    <Animated.View
      accessibilityLabel={message}
      accessibilityRole="progressbar"
      accessibilityViewIsModal
      pointerEvents="auto"
      style={[styles.overlay, { backgroundColor: colors.surface.canvas, opacity: visibility }]}
    >
      <View style={styles.content}>
        <Text variant="bodyLarge">{message}</Text>
        <ActivityIndicator color={colors.text.primary} size="large" />
      </View>
    </Animated.View>
  );
}

const styles = StyleSheet.create({
  overlay: {
    alignItems: "center",
    bottom: 0,
    elevation: 2,
    justifyContent: "center",
    left: 0,
    position: "absolute",
    right: 0,
    top: 0,
    zIndex: 2,
  },
  content: { alignItems: "center", gap: space[4] },
});
