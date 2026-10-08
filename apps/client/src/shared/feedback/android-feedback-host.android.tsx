import { requireNativeModule, requireNativeView } from "expo";
import { useEffect, useRef } from "react";
import { StyleSheet, useWindowDimensions, type ViewProps } from "react-native";

import { banner, motion, radius, space, typography } from "@tournaments-manager/design-tokens";

import { usePreferences } from "@/shared/preferences/preferences-provider";

import type { AndroidFeedbackHostProps } from "./android-feedback-host.types";

type BannerSpec = {
  id: number;
  message: string;
  expiresAt: number;
  backgroundColor: string;
  borderColor: string;
  textColor: string;
  fontFamily: string;
  fontSize: number;
  lineHeight: number;
  padding: number;
  borderRadius: number;
  borderWidth: number;
  left: number;
  top: number;
  width: number;
  enterExitMs: number;
  feedbackMs: number;
  swipeDistance: number;
  swipeSlop: number;
  reducedMotion: boolean;
};

type GlobalFeedbackModule = {
  show: (spec: BannerSpec) => Promise<void>;
  hide: (reducedMotion: boolean, duration: number) => Promise<void>;
  addListener: (
    name: "onDismiss",
    listener: (event: { id: number }) => void,
  ) => { remove: () => void };
};

const nativeHost = requireNativeModule<GlobalFeedbackModule>("TMGlobalFeedback");
const NativeWindowAnchor = requireNativeView<ViewProps>("TMGlobalFeedback");

export function FeedbackWindowAnchor() {
  return (
    <NativeWindowAnchor
      accessible={false}
      collapsable={false}
      pointerEvents="none"
      style={styles.anchor}
    />
  );
}

export function AndroidFeedbackHost({
  feedback,
  onDismiss,
  reducedMotion,
  topInset,
}: AndroidFeedbackHostProps) {
  const { colors } = usePreferences();
  const { width, fontScale } = useWindowDimensions();
  const deadline = useRef<{ id: number; expiresAt: number } | null>(null);

  useEffect(() => {
    const subscription = nativeHost.addListener("onDismiss", ({ id }) => onDismiss(id));
    return () => subscription.remove();
  }, [onDismiss]);

  useEffect(() => {
    if (!feedback) {
      deadline.current = null;
      void nativeHost.hide(reducedMotion, motion.enterExit);
      return;
    }
    if (deadline.current?.id !== feedback.id) {
      deadline.current = { id: feedback.id, expiresAt: Date.now() + banner.autoDismissMs };
    }
    void nativeHost.show({
      id: feedback.id,
      message: feedback.message,
      expiresAt: deadline.current.expiresAt,
      backgroundColor: colors.surface.default,
      borderColor: feedback.kind === "success" ? colors.feedback.success : colors.feedback.error,
      textColor: colors.text.primary,
      fontFamily: typography.family.regular,
      fontSize: typography.size.body,
      lineHeight: typography.size.body * typography.lineHeight.default,
      padding: space[3],
      borderRadius: radius.card,
      borderWidth: 1,
      left: space[5],
      top: topInset + space[1],
      width: Math.max(0, width - space[5] * 2),
      enterExitMs: motion.enterExit,
      feedbackMs: motion.feedback,
      swipeDistance: space[10],
      swipeSlop: space[2],
      reducedMotion,
    });
  }, [colors, feedback, fontScale, reducedMotion, topInset, width]);

  useEffect(() => () => void nativeHost.hide(true, 0), []);

  return <FeedbackWindowAnchor />;
}

const styles = StyleSheet.create({
  anchor: { height: 0, position: "absolute", width: 0 },
});
