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
          color={variant === "primary" ? colors.text.inverse : color.brand.primary}
        />
      ) : null}
      <Text color={variant === "primary" || variant === "destructive" ? "inverse" : "primary"}>
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
      {variant === "primary" ? (
        <LinearGradient colors={[...gradient.brand]} style={styles.base}>
          {content}
        </LinearGradient>
      ) : null}
      {variant === "secondary" ? (
        <LinearGradient colors={[...gradient.brand]} style={styles.outline}>
          <View style={[styles.outlineContent, { backgroundColor: colors.surface.default }]}>
            {content}
          </View>
        </LinearGradient>
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

const styles = StyleSheet.create({
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
