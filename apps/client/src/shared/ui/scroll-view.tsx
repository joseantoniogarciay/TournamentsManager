import { useRef, useState, type Ref } from "react";
import {
  ScrollView as NativeScrollView,
  StyleSheet,
  View,
  type ScrollViewInstance,
  type ScrollViewProps,
} from "react-native";
import { LinearGradient } from "expo-linear-gradient";
import { space } from "@tournaments-manager/design-tokens";
import { usePreferences } from "@/shared/preferences/preferences-provider";

/** Fade only the vertical content that continues beyond the viewport. */
export function ScrollView({
  ref,
  style,
  onLayout,
  onScroll,
  onContentSizeChange,
  scrollEventThrottle = 16,
  ...props
}: ScrollViewProps & { ref?: Ref<ScrollViewInstance> }) {
  const { colors } = usePreferences();
  const metrics = useRef({
    height: 0,
    content: 0,
    offset: props.contentOffset?.y ?? 0,
    top: props.contentInset?.top ?? 0,
    bottom: props.contentInset?.bottom ?? 0,
  });
  const [edges, setEdges] = useState({ top: 0, bottom: 0 });
  const update = () => {
    const { height, content, offset, top, bottom } = metrics.current;
    const next = {
      top: Math.min(1, Math.max(0, offset + top) / space[3]),
      bottom:
        height > 0 ? Math.min(1, Math.max(0, content + bottom - height - offset) / space[3]) : 0,
    };
    setEdges((previous) =>
      previous.top === next.top && previous.bottom === next.bottom ? previous : next,
    );
  };
  if (props.horizontal)
    return (
      <NativeScrollView
        {...props}
        ref={ref}
        style={style}
        onLayout={onLayout}
        onScroll={onScroll}
        onContentSizeChange={onContentSizeChange}
        scrollEventThrottle={scrollEventThrottle}
      />
    );
  const background = colors.surface.canvas;
  return (
    <View style={[styles.viewport, style]}>
      <NativeScrollView
        {...props}
        ref={ref}
        style={styles.scroll}
        scrollEventThrottle={scrollEventThrottle}
        onLayout={(event) => {
          metrics.current.height = event.nativeEvent.layout.height;
          update();
          onLayout?.(event);
        }}
        onContentSizeChange={(width, height) => {
          metrics.current.content = height;
          update();
          onContentSizeChange?.(width, height);
        }}
        onScroll={(event) => {
          const { contentOffset, contentInset, layoutMeasurement, contentSize } = event.nativeEvent;
          metrics.current = {
            offset: contentOffset.y,
            top: contentInset?.top ?? 0,
            bottom: contentInset?.bottom ?? 0,
            height: layoutMeasurement.height,
            content: contentSize.height,
          };
          update();
          onScroll?.(event);
        }}
      />
      <LinearGradient
        pointerEvents="none"
        accessible={false}
        importantForAccessibility="no-hide-descendants"
        colors={[background, `${background}00`]}
        style={[styles.edge, styles.top, { opacity: edges.top }]}
      />
      <LinearGradient
        pointerEvents="none"
        accessible={false}
        importantForAccessibility="no-hide-descendants"
        colors={[`${background}00`, background]}
        style={[styles.edge, styles.bottom, { opacity: edges.bottom }]}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  viewport: { flexGrow: 1, flexShrink: 1, minHeight: 0 },
  scroll: { flex: 1 },
  edge: { position: "absolute", left: 0, right: 0, height: space[3] },
  top: { top: 0 },
  bottom: { bottom: 0 },
});
