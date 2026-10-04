import { useCallback, useEffect, useRef, useState } from "react";
import {
  Keyboard,
  Platform,
  ScrollView,
  TextInput,
  View,
  type KeyboardEvent,
  type ScrollViewProps,
} from "react-native";

import { space } from "@tournaments-manager/design-tokens";

/** Reserva el espacio ocluido y mantiene visible el campo enfocado, también con autoFocus. */
export function KeyboardAwareScrollView({
  children,
  keyboardShouldPersistTaps = "handled",
  keyboardDismissMode = "none",
  onFocus,
  onBlur,
  onLayout,
  onScroll,
  onContentSizeChange,
  scrollEventThrottle = 16,
  ...props
}: ScrollViewProps) {
  const scroll = useRef<ScrollView>(null);
  const focusedInput = useRef<ReturnType<typeof TextInput.State.currentlyFocusedInput>>(null);
  const keyboardTop = useRef<number | undefined>(undefined);
  const offset = useRef(0);
  const pendingFrame = useRef<number | undefined>(undefined);
  const [bottomPadding, setBottomPadding] = useState(0);

  const revealFocusedInput = useCallback(() => {
    if (pendingFrame.current !== undefined) cancelAnimationFrame(pendingFrame.current);
    pendingFrame.current = requestAnimationFrame(() => {
      pendingFrame.current = undefined;
      const input = focusedInput.current;
      const view = scroll.current;
      const top = keyboardTop.current;
      if (!view || !input || top === undefined || TextInput.State.currentlyFocusedInput() !== input)
        return;
      view.getNativeScrollRef()?.measureInWindow((_x, y, _width, height) => {
        if (focusedInput.current !== input || keyboardTop.current !== top) return;
        // Android suele reducir la ventana: no sumamos de nuevo toda la altura del teclado.
        const overlap = Math.max(0, y + height - top);
        setBottomPadding(space[5] + (Platform.OS === "android" ? overlap : 0));
        input.measureInWindow((_inputX, inputY, _inputWidth, inputHeight) => {
          if (focusedInput.current !== input || keyboardTop.current !== top) return;
          const obscured = inputY + inputHeight + space[5] - Math.min(y + height, top);
          if (obscured > 0) view.scrollTo({ y: offset.current + obscured, animated: true });
        });
      });
    });
  }, []);

  useEffect(() => {
    if (Platform.OS === "web") return;
    const show = (event: KeyboardEvent) => {
      keyboardTop.current = event.endCoordinates.screenY;
      revealFocusedInput();
    };
    const hide = () => {
      keyboardTop.current = undefined;
      setBottomPadding(0);
    };
    const subscriptions = [
      Keyboard.addListener("keyboardDidShow", show),
      Keyboard.addListener("keyboardDidHide", hide),
      ...(Platform.OS === "ios" ? [Keyboard.addListener("keyboardDidChangeFrame", show)] : []),
    ];
    return () => {
      subscriptions.forEach((subscription) => subscription.remove());
      if (pendingFrame.current !== undefined) cancelAnimationFrame(pendingFrame.current);
    };
  }, [revealFocusedInput]);

  return (
    <ScrollView
      {...props}
      ref={scroll}
      automaticallyAdjustKeyboardInsets={Platform.OS === "ios"}
      keyboardDismissMode={keyboardDismissMode}
      keyboardShouldPersistTaps={keyboardShouldPersistTaps}
      scrollEventThrottle={scrollEventThrottle}
      onFocus={(event) => {
        if (Platform.OS !== "web") {
          focusedInput.current = TextInput.State.currentlyFocusedInput();
          keyboardTop.current = Keyboard.metrics()?.screenY;
          revealFocusedInput();
        }
        onFocus?.(event);
      }}
      onBlur={(event) => {
        focusedInput.current = null;
        onBlur?.(event);
      }}
      onLayout={(event) => {
        revealFocusedInput();
        onLayout?.(event);
      }}
      onScroll={(event) => {
        offset.current = event.nativeEvent.contentOffset.y;
        onScroll?.(event);
      }}
      onContentSizeChange={(width, height) => {
        revealFocusedInput();
        onContentSizeChange?.(width, height);
      }}
    >
      {children}
      {bottomPadding > 0 ? <View style={{ height: bottomPadding }} /> : null}
    </ScrollView>
  );
}
