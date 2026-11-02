import { useFocusEffect } from "expo-router";
import { useCallback, useState } from "react";
import { useIllustrationEnvironment } from "./use-illustration-environment";

/** Route moments also require focus; global loading overlays use the environment directly. */
export function useIllustrationPlayback() {
  const { reducedMotion, motionReady, active } = useIllustrationEnvironment();
  const [focused, setFocused] = useState(false);
  useFocusEffect(
    useCallback(() => {
      setFocused(true);
      return () => setFocused(false);
    }, []),
  );
  return {
    reducedMotion,
    motionReady,
    playing: motionReady && !reducedMotion && focused && active,
  };
}
