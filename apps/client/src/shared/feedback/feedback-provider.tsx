import {
  createContext,
  type PropsWithChildren,
  type ReactElement,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import {
  AccessibilityInfo,
  Animated,
  Modal,
  PanResponder,
  Pressable,
  StyleSheet,
  View,
} from "react-native";
import { useFocusEffect } from "expo-router";
import { useSafeAreaInsets } from "react-native-safe-area-context";

import { banner, motion, radius, space } from "@tournaments-manager/design-tokens";

import { usePreferences } from "@/shared/preferences/preferences-provider";
import { Text } from "@/shared/ui/text";

type Feedback = { message: string; kind: "network-error" | "generic-error" | "success" };
type ActiveFeedback = Feedback & { id: number };
const navigationFeedbackDelayMs = 400;
type FeedbackContextValue = {
  banner: ReactElement | null;
  dismiss: () => void;
  setFeedbackHostFocused: (isFocused: boolean) => void;
  show: (feedback: Feedback) => void;
  showAfterNavigation: (feedback: Feedback) => void;
};
const FeedbackContext = createContext<FeedbackContextValue | null>(null);

export function FeedbackProvider({ children }: PropsWithChildren) {
  const { colors } = usePreferences();
  const [feedback, setFeedback] = useState<ActiveFeedback | null>(null);
  const [reducedMotion, setReducedMotion] = useState(false);
  const insets = useSafeAreaInsets();
  const visibility = useRef(new Animated.Value(0)).current;
  const dragY = useRef(new Animated.Value(0)).current;
  const activeFeedbackId = useRef<number | null>(null);
  const nextFeedbackId = useRef(0);
  const autoDismissTimeout = useRef<ReturnType<typeof setTimeout> | null>(null);
  const navigationFeedbackTimeout = useRef<ReturnType<typeof setTimeout> | null>(null);
  const queuedFeedback = useRef<Feedback | null>(null);
  const awaitingNextFocusedHost = useRef(false);

  useEffect(() => {
    void AccessibilityInfo.isReduceMotionEnabled().then(setReducedMotion);
    const subscription = AccessibilityInfo.addEventListener(
      "reduceMotionChanged",
      setReducedMotion,
    );
    return () => subscription.remove();
  }, []);

  const clearAutoDismiss = useCallback(() => {
    if (autoDismissTimeout.current) clearTimeout(autoDismissTimeout.current);
    autoDismissTimeout.current = null;
  }, []);

  const dismiss = useCallback(
    (id = activeFeedbackId.current) => {
      if (id === null || id !== activeFeedbackId.current) return;

      clearAutoDismiss();
      if (reducedMotion) {
        activeFeedbackId.current = null;
        setFeedback(null);
        return;
      }

      visibility.stopAnimation();
      Animated.timing(visibility, {
        duration: motion.enterExit,
        toValue: 0,
        useNativeDriver: true,
      }).start(({ finished }) => {
        if (finished && activeFeedbackId.current === id) {
          activeFeedbackId.current = null;
          setFeedback(null);
        }
      });
    },
    [clearAutoDismiss, reducedMotion, visibility],
  );

  const panResponder = useMemo(
    () =>
      PanResponder.create({
        onMoveShouldSetPanResponder: (_, gesture) =>
          gesture.dy < -space[2] && Math.abs(gesture.dy) > Math.abs(gesture.dx),
        onPanResponderMove: (_, gesture) => dragY.setValue(Math.min(gesture.dy, 0)),
        onPanResponderRelease: (_, gesture) => {
          if (gesture.dy <= -space[10] || gesture.vy <= -0.5) {
            dismiss();
            return;
          }

          Animated.timing(dragY, {
            duration: motion.feedback,
            toValue: 0,
            useNativeDriver: true,
          }).start();
        },
        onPanResponderTerminate: () => {
          Animated.timing(dragY, {
            duration: motion.feedback,
            toValue: 0,
            useNativeDriver: true,
          }).start();
        },
      }),
    [dismiss, dragY],
  );

  useEffect(() => {
    if (!feedback) return;

    visibility.stopAnimation();
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
  }, [feedback, reducedMotion, visibility]);

  useEffect(() => {
    return clearAutoDismiss;
  }, [clearAutoDismiss]);

  useEffect(() => {
    return () => {
      if (navigationFeedbackTimeout.current) clearTimeout(navigationFeedbackTimeout.current);
    };
  }, []);

  const show = useCallback(
    (next: Feedback) => {
      if (navigationFeedbackTimeout.current) clearTimeout(navigationFeedbackTimeout.current);
      navigationFeedbackTimeout.current = null;
      queuedFeedback.current = null;
      awaitingNextFocusedHost.current = false;
      clearAutoDismiss();
      visibility.stopAnimation();
      visibility.setValue(0);
      dragY.stopAnimation();
      dragY.setValue(0);

      const id = ++nextFeedbackId.current;
      activeFeedbackId.current = id;
      setFeedback({ ...next, id });
      autoDismissTimeout.current = setTimeout(() => dismiss(id), banner.autoDismissMs);
    },
    [clearAutoDismiss, dismiss, dragY, visibility],
  );
  const showAfterNavigation = useCallback((next: Feedback) => {
    queuedFeedback.current = next;
    awaitingNextFocusedHost.current = true;
  }, []);
  const setFeedbackHostFocused = useCallback(
    (isFocused: boolean) => {
      if (!isFocused || !awaitingNextFocusedHost.current || !queuedFeedback.current) return;

      const next = queuedFeedback.current;
      queuedFeedback.current = null;
      awaitingNextFocusedHost.current = false;
      navigationFeedbackTimeout.current = setTimeout(() => show(next), navigationFeedbackDelayMs);
    },
    [show],
  );
  const bannerNode = feedback ? (
    <Animated.View
      {...panResponder.panHandlers}
      style={[
        styles.banner,
        {
          backgroundColor: colors.surface.default,
          borderColor:
            feedback.kind === "success" ? colors.feedback.success : colors.feedback.error,
          opacity: visibility,
          top: insets.top + space[1],
          transform: [
            {
              translateY: Animated.add(
                visibility.interpolate({
                  inputRange: [0, 1],
                  outputRange: [-space[3], 0],
                }),
                dragY,
              ),
            },
          ],
        },
      ]}
    >
      <Pressable accessibilityRole="alert" onPress={() => dismiss(feedback.id)}>
        <Text>{feedback.message}</Text>
      </Pressable>
    </Animated.View>
  ) : null;
  const contextValue = useMemo(
    () => ({
      banner: bannerNode,
      dismiss,
      setFeedbackHostFocused,
      show,
      showAfterNavigation,
    }),
    [bannerNode, dismiss, setFeedbackHostFocused, show, showAfterNavigation],
  );

  return <FeedbackContext.Provider value={contextValue}>{children}</FeedbackContext.Provider>;
}

export function useFeedback() {
  const value = useContext(FeedbackContext);
  if (!value) throw new Error("useFeedback debe usarse dentro de FeedbackProvider");
  return value;
}

export function FeedbackBanner() {
  const { banner: feedbackBanner, dismiss, setFeedbackHostFocused } = useFeedback();
  const [isFocused, setIsFocused] = useState(false);
  useFocusEffect(
    useCallback(() => {
      setIsFocused(true);
      setFeedbackHostFocused(true);
      return () => {
        setIsFocused(false);
        setFeedbackHostFocused(false);
      };
    }, [setFeedbackHostFocused]),
  );
  if (!feedbackBanner || !isFocused) return null;

  return (
    <Modal
      animationType="none"
      onRequestClose={() => dismiss()}
      presentationStyle="overFullScreen"
      statusBarTranslucent
      transparent
      visible
    >
      <View pointerEvents="box-none" style={styles.modalHost}>
        {feedbackBanner}
      </View>
    </Modal>
  );
}

const styles = StyleSheet.create({
  modalHost: { flex: 1 },
  banner: {
    borderRadius: radius.card,
    borderWidth: 1,
    left: space[5],
    padding: space[3],
    position: "absolute",
    right: space[5],
    zIndex: 1,
  },
});
