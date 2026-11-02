import { useEffect, useState } from "react";
import { AccessibilityInfo, AppState, Platform } from "react-native";

/** Motion/app state shared by route illustrations and root loading overlays. */
export function useIllustrationEnvironment() {
  const [reducedMotion, setReducedMotion] = useState(true);
  const [motionReady, setMotionReady] = useState(false);
  const [active, setActive] = useState(AppState.currentState === "active");

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

  return { reducedMotion, motionReady, active };
}
