import type { ReactNode } from "react";
import type { StyleProp, ViewStyle } from "react-native";
import { ActivityIndicator, Pressable, StyleSheet, View } from "react-native";

import { LinearGradient } from "expo-linear-gradient";

import { color, control, gradient, radius, space } from "@tournaments-manager/design-tokens";

import { usePreferences } from "@/shared/preferences/preferences-provider";

import { Text } from "./text";

type Props = {
  label: string;
  onPress: () => void;
  variant?: "primary" | "secondary" | "ghost" | "destructive";
  loading?: boolean;
  disabled?: boolean;
};

export function Button({
  label,
  onPress,
  variant = "primary",
  loading = false,
  disabled = false,
}: Props) {
  const { colors } = usePreferences();
  const isDisabled = disabled || loading;

  const content = (
    <View style={styles.content}>
      {loading ? (
        <ActivityIndicator
          color={variant === "primary" ? color.text.inverse : color.brand.primary}
        />
      ) : null}
      <Text
        color={
          variant === "primary" ? "onBrand" : variant === "destructive" ? "inverse" : "primary"
        }
      >
        {label}
      </Text>
    </View>
  );

  return (
    <Pressable
      accessibilityRole="button"
      accessibilityState={{ busy: loading, disabled: isDisabled }}
      disabled={isDisabled}
      onPress={onPress}
      style={isDisabled ? styles.disabled : undefined}
    >
      {variant === "primary" ? <BrandGradient style={styles.base}>{content}</BrandGradient> : null}
      {variant === "secondary" ? (
        <BrandGradient style={styles.outline}>
          <View style={[styles.outlineContent, { backgroundColor: colors.surface.default }]}>
            {content}
          </View>
        </BrandGradient>
      ) : null}
      {variant === "ghost" || variant === "destructive" ? (
        <View
          style={[
            styles.base,
            variant === "destructive"
              ? { backgroundColor: colors.feedback.error }
              : { backgroundColor: "transparent" },
          ]}
        >
          {content}
        </View>
      ) : null}
    </Pressable>
  );
}

function BrandGradient({ children, style }: { children: ReactNode; style: StyleProp<ViewStyle> }) {
  return (
    <View style={[styles.brandGradient, style]}>
      <LinearGradient
        {...gradient.brandButton}
        pointerEvents="none"
        style={StyleSheet.absoluteFill}
      />
      {children}
    </View>
  );
}

const styles = StyleSheet.create({
  brandGradient: { backgroundColor: color.brand.primary, overflow: "hidden" },
  base: {
    minHeight: control.minHeight,
    borderRadius: radius.control,
    justifyContent: "center",
    paddingHorizontal: control.horizontalPadding,
  },
  content: { alignItems: "center", flexDirection: "row", gap: space[2], justifyContent: "center" },
  outline: {
    borderRadius: radius.control,
    minHeight: control.minHeight,
    padding: 1,
  },
  outlineContent: {
    borderRadius: radius.control - 1,
    flex: 1,
    justifyContent: "center",
    paddingHorizontal: control.horizontalPadding - 1,
  },
  disabled: { opacity: 0.55 },
});
