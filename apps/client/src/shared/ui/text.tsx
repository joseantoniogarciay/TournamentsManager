import { type PropsWithChildren } from "react";
import { StyleSheet, Text as NativeText, type StyleProp, type TextStyle } from "react-native";

import { color, typography } from "@tournaments-manager/design-tokens";

import { usePreferences } from "@/shared/preferences/preferences-provider";

type Props = PropsWithChildren<{
  variant?: "body" | "bodyLarge" | "caption" | "title" | "display";
  color?: "primary" | "secondary" | "inverse" | "onBrand" | "error" | "success";
  numberOfLines?: number;
  style?: StyleProp<TextStyle>;
}>;

export function Text({
  children,
  variant = "body",
  color: textColor = "primary",
  numberOfLines,
  style,
}: Props) {
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
    <NativeText
      numberOfLines={numberOfLines}
      style={[styles.base, variants[variant], textColors[textColor], style]}
    >
      {children}
    </NativeText>
  );
}

const styles = StyleSheet.create({
  base: { fontFamily: typography.family.regular },
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
    fontFamily: typography.family.semibold,
    fontSize: typography.size.title,
    lineHeight: typography.size.title * typography.lineHeight.compact,
  },
  display: {
    fontFamily: typography.family.bold,
    fontSize: typography.size.display,
    lineHeight: typography.size.display * typography.lineHeight.compact,
  },
};
