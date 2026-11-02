import { useFocusEffect } from "expo-router";
import { useCallback, useEffect, useState } from "react";
import { AccessibilityInfo, AppState, Platform } from "react-native";

/** Default to a still until the system preference is known. Never animate offscreen. */
export function useIllustrationPlayback() {
  const [reducedMotion, setReducedMotion] = useState(true);
  const [motionReady, setMotionReady] = useState(false);
  const [focused, setFocused] = useState(false);
  const [active, setActive] = useState(AppState.currentState === "active");

  useFocusEffect(
    useCallback(() => {
      setFocused(true);
      return () => setFocused(false);
    }, []),
  );

  useEffect(() => {
    const subscription = AppState.addEventListener("change", (state) =>
      setActive(state === "active"),
    );
    if (Platform.OS === "web") {
      const query = window.matchMedia("(prefers-reduced-motion: reduce)");
      const update = () => {
        setReducedMotion(query.matches);
        setMotionReady(true);
      };
      update();
      query.addEventListener("change", update);
      return () => {
        query.removeEventListener("change", update);
        subscription.remove();
      };
    }
    let live = true;
    let changed = false;
    const motionSubscription = AccessibilityInfo.addEventListener(
      "reduceMotionChanged",
      (value) => {
        changed = true;
        setReducedMotion(value);
        setMotionReady(true);
      },
    );
    void AccessibilityInfo.isReduceMotionEnabled()
      .then((value) => {
        if (live && !changed) {
          setReducedMotion(value);
          setMotionReady(true);
        }
      })
      .catch(() => {
        if (live) setMotionReady(true);
      });
    return () => {
      live = false;
      motionSubscription.remove();
      subscription.remove();
    };
  }, []);

  return {
    reducedMotion,
    motionReady,
    playing: motionReady && !reducedMotion && focused && active,
  };
}
