import LottieView from "lottie-react-native";
import { useEffect, useMemo, useRef } from "react";
import { StyleSheet, View } from "react-native";
import { illustration } from "@tournaments-manager/design-tokens";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import {
  createTournamentIllustration,
  type TournamentIllustrationVariant,
} from "./tournament-illustration-data";
import { useIllustrationPlayback } from "./use-illustration-playback";

export function TournamentIllustration({ variant }: { variant: TournamentIllustrationVariant }) {
  const { colors } = usePreferences();
  const source = useMemo(
    () =>
      createTournamentIllustration(variant, {
        accent: colors.indicator.default,
        ink: colors.text.secondary,
        subtle: colors.surface.subtle,
      }),
    [colors, variant],
  );
  const ref = useRef<LottieView>(null);
  const consumed = useRef(false);
  const ready = useRef(false);
  const { reducedMotion, motionReady, playing } = useIllustrationPlayback();
  const playback = useRef({ reducedMotion, motionReady, playing });
  playback.current = { reducedMotion, motionReady, playing };
  const update = () => {
    if (!ready.current) return;
    const current = playback.current;
    if (!current.motionReady) return;
    if (current.reducedMotion || (!current.playing && consumed.current)) {
      consumed.current = true;
      ref.current?.play(source.op - 1, source.op - 1);
    } else if (current.playing && !consumed.current) {
      consumed.current = true;
      ref.current?.play(0, source.op - 1);
    }
  };
  useEffect(update, [source, reducedMotion, motionReady, playing]);

  return (
    <View
      accessible={false}
      accessibilityElementsHidden
      importantForAccessibility="no-hide-descendants"
      pointerEvents="none"
      style={[
        styles.frame,
        { width: variant === "trophy" ? illustration.celebrationWidth : illustration.emptyWidth },
      ]}
    >
      <LottieView
        ref={ref}
        source={source}
        autoPlay={false}
        loop={false}
        progress={(source.op - 1) / source.op}
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
  frame: { maxWidth: "100%", aspectRatio: illustration.aspectRatio, alignSelf: "center" },
  art: { width: "100%", height: "100%" },
});
