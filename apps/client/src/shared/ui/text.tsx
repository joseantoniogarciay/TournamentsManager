import { type PropsWithChildren } from "react";
import { StyleSheet, Text as NativeText, type TextStyle } from "react-native";

import { color, typography } from "@tournaments-manager/design-tokens";

type Props = PropsWithChildren<{
  variant?: "body" | "bodyLarge" | "caption" | "title" | "display";
  color?: "primary" | "secondary" | "inverse" | "error";
}>;

export function Text({ children, variant = "body", color: textColor = "primary" }: Props) {
  return (
    <NativeText style={[styles.base, variants[variant], colors[textColor]]}>{children}</NativeText>
  );
}

const styles = StyleSheet.create({
  base: { fontFamily: typography.family.system },
  primary: { color: color.text.primary },
  secondary: { color: color.text.secondary },
  inverse: { color: color.text.inverse },
  error: { color: color.feedback.error },
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

const colors = {
  primary: styles.primary,
  secondary: styles.secondary,
  inverse: styles.inverse,
  error: styles.error,
};
