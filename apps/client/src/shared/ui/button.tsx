import type { ReactNode } from "react";
import type { StyleProp, ViewStyle } from "react-native";
import { ActivityIndicator, StyleSheet, View } from "react-native";

import { LinearGradient } from "expo-linear-gradient";

import {
  color,
  control,
  gradient,
  radius,
  space,
  typography,
} from "@tournaments-manager/design-tokens";

import { usePreferences } from "@/shared/preferences/preferences-provider";

import { Text } from "./text";
import { ActionPressable, type ActionAvailability } from "./action-pressable";
import { getTranslator } from "@/shared/i18n/locale";

type Props = ActionAvailability & {
  label: string;
  onPress: () => void;
  variant?: "primary" | "secondary" | "ghost" | "destructive";
  loading?: boolean;
  secondarySurfaceColor?: string;
};

export function Button({
  label,
  onPress,
  variant = "primary",
  loading = false,
  disabled = false,
  disabledReason,
  secondarySurfaceColor,
}: Props) {
  const { colors } = usePreferences();
  const isDisabled = disabled || loading;
  const t = getTranslator();

  const content = (
    <View style={styles.content}>
      {loading ? <ActivityIndicator color={colors.text.secondary} /> : null}
      <Text
        color={
          isDisabled
            ? "secondary"
            : variant === "destructive"
              ? "error"
              : variant === "primary"
                ? "onBrand"
                : "primary"
        }
        style={styles.label}
      >
        {label}
      </Text>
    </View>
  );

  return (
    <ActionPressable
      accessibilityRole="button"
      accessibilityState={{ busy: loading, disabled: isDisabled }}
      disabled={isDisabled}
      disabledReason={
        loading
          ? t("common_action_in_progress")
          : (disabledReason ?? t("common_action_in_progress"))
      }
      onPress={onPress}
      style={styles.pressable}
    >
      {isDisabled ? (
        <View
          style={[
            styles.base,
            styles.destructiveOutline,
            { backgroundColor: colors.surface.subtle, borderColor: colors.border.default },
          ]}
        >
          {content}
        </View>
      ) : (
        <>
          {variant === "primary" ? (
            <BrandGradient style={styles.base}>{content}</BrandGradient>
          ) : null}
          {variant === "secondary" ? (
            <BrandGradient outline style={[styles.base, styles.outline]}>
              <View
                style={[
                  styles.outlineContent,
                  { backgroundColor: secondarySurfaceColor ?? colors.surface.default },
                ]}
              >
                {content}
              </View>
            </BrandGradient>
          ) : null}
          {variant === "destructive" ? (
            <View
              style={[
                styles.base,
                styles.destructiveOutline,
                { borderColor: colors.feedback.error },
              ]}
            >
              {content}
            </View>
          ) : null}
          {variant === "ghost" ? (
            <View style={[styles.base, { backgroundColor: "transparent" }]}>{content}</View>
          ) : null}
        </>
      )}
    </ActionPressable>
  );
}

function BrandGradient({
  children,
  outline = false,
  style,
}: {
  children: ReactNode;
  outline?: boolean;
  style: StyleProp<ViewStyle>;
}) {
  const { colors, resolvedTheme } = usePreferences();
  const usesDarkOutline = outline && resolvedTheme === "dark";

  return (
    <View
      style={[
        styles.brandGradient,
        usesDarkOutline && { backgroundColor: colors.border.focus },
        style,
      ]}
    >
      {usesDarkOutline ? null : (
        <LinearGradient
          {...gradient.brandButton}
          pointerEvents="none"
          style={StyleSheet.absoluteFill}
        />
      )}
      {children}
    </View>
  );
}

const styles = StyleSheet.create({
  brandGradient: { backgroundColor: color.brand.primary, overflow: "hidden" },
  base: {
    minHeight: control.minHeight,
    borderRadius: radius.pill,
    justifyContent: "center",
  },
  content: {
    alignItems: "center",
    flexDirection: "row",
    gap: space[2],
    justifyContent: "center",
    paddingHorizontal: control.horizontalPadding,
  },
  outline: { padding: 1 },
  destructiveOutline: { borderWidth: 1 },
  outlineContent: {
    borderRadius: radius.pill,
    justifyContent: "center",
    minHeight: control.minHeight - 2,
  },
  label: { fontFamily: typography.family.semibold, textAlign: "center" },
  pressable: { borderRadius: radius.pill },
});
