import { ActivityIndicator, Pressable, StyleSheet, View } from "react-native";

import { color, control, radius, space } from "@tournaments-manager/design-tokens";

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
  const isDisabled = disabled || loading;
  return (
    <Pressable
      accessibilityRole="button"
      accessibilityState={{ busy: loading, disabled: isDisabled }}
      disabled={isDisabled}
      onPress={onPress}
      style={[styles.base, styles[variant], isDisabled && styles.disabled]}
    >
      <View style={styles.content}>
        {loading ? (
          <ActivityIndicator
            color={variant === "primary" ? color.text.inverse : color.brand.primary}
          />
        ) : null}
        <Text color={variant === "primary" || variant === "destructive" ? "inverse" : "primary"}>
          {label}
        </Text>
      </View>
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
  primary: { backgroundColor: color.brand.primary },
  secondary: {
    backgroundColor: color.surface.default,
    borderColor: color.border.default,
    borderWidth: 1,
  },
  ghost: { backgroundColor: "transparent" },
  destructive: { backgroundColor: color.feedback.error },
  disabled: { opacity: 0.55 },
});
