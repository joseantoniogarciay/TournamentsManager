import { useEffect, useMemo, useRef } from "react";
import type { AnimationItem } from "lottie-web";
import lottie from "lottie-web/build/player/lottie_light";
import { illustration } from "@tournaments-manager/design-tokens";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import {
  createTournamentIllustration,
  type TournamentIllustrationVariant,
} from "./tournament-illustration-data";
import { useIllustrationPlayback } from "./use-illustration-playback";

export function TournamentIllustration({
  variant,
  animate = true,
}: {
  variant: TournamentIllustrationVariant;
  animate?: boolean;
}) {
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
  const container = useRef<HTMLDivElement>(null);
  const player = useRef<AnimationItem | null>(null);
  const svgReady = useRef(false);
  const consumed = useRef(false);
  const { reducedMotion, motionReady, playing } = useIllustrationPlayback();
  const playback = useRef({ reducedMotion, motionReady, playing });
  playback.current = {
    reducedMotion: reducedMotion || !animate,
    motionReady,
    playing: playing && animate,
  };
  const update = () => {
    const animation = player.current;
    const current = playback.current;
    if (!animation || !svgReady.current || !current.motionReady) return;
    if (current.reducedMotion || (!current.playing && consumed.current)) {
      consumed.current = true;
      animation.goToAndStop(source.op - 1, true);
    } else if (current.playing && !consumed.current) {
      consumed.current = true;
      animation.goToAndPlay(0, true);
    }
  };

  useEffect(() => {
    if (!container.current) return;
    svgReady.current = false;
    // loadAnimation mutates its data; keep each player instance isolated.
    const animation = lottie.loadAnimation({
      container: container.current,
      renderer: "svg",
      loop: false,
      autoplay: false,
      animationData: structuredClone(source),
      rendererSettings: { progressiveLoad: false, hideOnTransparent: true },
    });
    player.current = animation;
    const onReady = () => {
      if (player.current !== animation) return;
      animation.goToAndStop(source.op - 1, true);
      // isLoaded becomes true before Lottie queues DOMLoaded. Starting sooner
      // lets this initial still frame overwrite and consume the first playback.
      svgReady.current = true;
      update();
    };
    animation.addEventListener("DOMLoaded", onReady);
    return () => {
      svgReady.current = false;
      player.current = null;
      animation.removeEventListener("DOMLoaded", onReady);
      animation.destroy();
    };
  }, [source]);
  useEffect(update, [source, reducedMotion, motionReady, playing, animate]);

  return (
    <div
      aria-hidden="true"
      ref={container}
      style={{
        width:
          variant === "bracket" || variant === "welcome"
            ? illustration.emptyWidth
            : illustration.celebrationWidth,
        maxWidth: "100%",
        aspectRatio: illustration.aspectRatio,
        alignSelf: "center",
        pointerEvents: "none",
      }}
    />
  );
}
