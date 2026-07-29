import { type PropsWithChildren } from "react";
import { StyleSheet, Text as NativeText, type TextStyle } from "react-native";

import { color, typography } from "@tournaments-manager/design-tokens";

import { usePreferences } from "@/shared/preferences/preferences-provider";

type Props = PropsWithChildren<{
  variant?: "body" | "bodyLarge" | "caption" | "title" | "display";
  color?: "primary" | "secondary" | "inverse" | "onBrand" | "error" | "success";
}>;

export function Text({ children, variant = "body", color: textColor = "primary" }: Props) {
  const { colors } = usePreferences();
  const textColors = {
    primary: { color: colors.text.primary },
    secondary: { color: colors.text.secondary },
    inverse: { color: colors.text.inverse },
    onBrand: { color: color.text.inverse },
    error: { color: colors.feedback.error },
    success: { color: colors.feedback.success },
  };
  return (
    <NativeText style={[styles.base, variants[variant], textColors[textColor]]}>
      {children}
    </NativeText>
  );
}

const styles = StyleSheet.create({
  base: { fontFamily: typography.family.system },
});

const variants: Record<NonNullable<Props["variant"]>, TextStyle> = {
  body: {
    fontSize: typography.size.body,
    lineHeight: typography.size.body * typography.lineHeight.default,
  },
  bodyLarge: {
    fontSize: typography.size.bodyLarge,
    lineHeight: typography.size.bodyLarge * typography.lineHeight.default,
  },
  caption: {
    fontSize: typography.size.caption,
    lineHeight: typography.size.caption * typography.lineHeight.default,
  },
  title: {
    fontSize: typography.size.title,
    fontWeight: typography.weight.semibold,
    lineHeight: typography.size.title * typography.lineHeight.compact,
  },
  display: {
    fontSize: typography.size.display,
    fontWeight: typography.weight.bold,
    lineHeight: typography.size.display * typography.lineHeight.compact,
  },
};
