import { useEffect, useMemo, useRef } from "react";
import type { AnimationItem } from "lottie-web";
import lottie from "lottie-web/build/player/lottie_light";
import { illustration } from "@tournaments-manager/design-tokens";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { createCupLoadingAnimation, cupLoadingStillFrame } from "./tournament-illustration-data";

/** Deliberately independent of navigation: also mounted by the root session overlay. */
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
  const container = useRef<HTMLDivElement>(null);
  const player = useRef<AnimationItem | null>(null);
  const ready = useRef(false);
  const animationAllowed = useRef(animate);
  animationAllowed.current = animate;
  const update = () => {
    if (!ready.current || !player.current) return;
    if (animationAllowed.current) player.current.goToAndPlay(0, true);
    else player.current.goToAndStop(cupLoadingStillFrame, true);
  };

  useEffect(() => {
    if (!container.current) return;
    ready.current = false;
    const animation = lottie.loadAnimation({
      container: container.current,
      renderer: "svg",
      loop: true,
      autoplay: false,
      animationData: structuredClone(source),
    });
    player.current = animation;
    const onReady = () => {
      if (player.current !== animation) return;
      animation.goToAndStop(cupLoadingStillFrame, true);
      ready.current = true;
      update();
    };
    animation.addEventListener("DOMLoaded", onReady);
    return () => {
      ready.current = false;
      player.current = null;
      animation.removeEventListener("DOMLoaded", onReady);
      animation.destroy();
    };
  }, [source]);
  useEffect(update, [animate, source]);

  return (
    <div
      aria-hidden="true"
      ref={container}
      style={{
        width: illustration.celebrationWidth,
        maxWidth: "100%",
        aspectRatio: illustration.aspectRatio,
        alignSelf: "center",
        pointerEvents: "none",
      }}
    />
  );
}
