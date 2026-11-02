import LottieView from "lottie-react-native";
import { useEffect, useMemo, useRef } from "react";
import { StyleSheet, View } from "react-native";
import { illustration } from "@tournaments-manager/design-tokens";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { createCupLoadingAnimation, cupLoadingStillFrame } from "./tournament-illustration-data";

export function CupLoader({ animate }: { animate: boolean }) {
  const { colors } = usePreferences();
  const source = useMemo(
    () =>
      createCupLoadingAnimation({
        accent: colors.indicator.default,
        ink: colors.text.secondary,
        subtle: colors.surface.subtle,
      }),
    [colors],
  );
  const ref = useRef<LottieView>(null);
  const ready = useRef(false);
  const animationAllowed = useRef(animate);
  animationAllowed.current = animate;
  const update = () => {
    if (!ready.current) return;
    if (animationAllowed.current) ref.current?.play(0, source.op - 1);
    else ref.current?.play(cupLoadingStillFrame, cupLoadingStillFrame);
  };
  useEffect(update, [animate, source]);
  return (
    <View
      accessible={false}
      accessibilityElementsHidden
      importantForAccessibility="no-hide-descendants"
      pointerEvents="none"
      style={styles.frame}
    >
      <LottieView
        ref={ref}
        source={source}
        autoPlay={false}
        loop={animate}
        progress={cupLoadingStillFrame / source.op}
        style={styles.art}
        resizeMode="contain"
        onAnimationLoaded={() => {
          ready.current = true;
          update();
        }}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  frame: {
    width: illustration.celebrationWidth,
    maxWidth: "100%",
    aspectRatio: illustration.aspectRatio,
    alignSelf: "center",
  },
  art: { width: "100%", height: "100%" },
});
