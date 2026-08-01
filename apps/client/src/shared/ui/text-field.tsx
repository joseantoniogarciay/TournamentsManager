import { SymbolView } from "expo-symbols";
import { useState } from "react";
import { Pressable, StyleSheet, TextInput, View, type TextInputProps } from "react-native";

import { control, radius, space, typography } from "@tournaments-manager/design-tokens";

import { usePreferences } from "@/shared/preferences/preferences-provider";

import { Text } from "./text";

type Props = TextInputProps & {
  label: string;
  error?: string;
  feedback?: { message: string; tone: "help" | "success" };
  passwordVisibility?: { isVisible: boolean; label: string; onPress: () => void };
  validationSubmitted?: boolean;
  validationTrigger?: "blur" | "change";
};

export function TextField({
  label,
  error,
  feedback,
  accessibilityHint,
  passwordVisibility,
  validationSubmitted = false,
  validationTrigger,
  onBlur,
  onChangeText,
  onFocus,
  ...inputProps
}: Props) {
  const { colors } = usePreferences();
  const [isFocused, setIsFocused] = useState(false);
  const [hasTriggeredValidation, setHasTriggeredValidation] = useState(false);
  const visibleError =
    validationTrigger && !validationSubmitted && !hasTriggeredValidation ? undefined : error;
  const message = visibleError ?? feedback?.message;
  const messageColor = visibleError
    ? "error"
    : feedback?.tone === "success"
      ? "success"
      : "secondary";
  return (
    <View style={styles.wrapper}>
      <Text variant="bodyLarge">{label}</Text>
      <View
        style={[
          styles.input,
          {
            borderColor: isFocused
              ? colors.border.focus
              : visibleError
                ? colors.border.error
                : colors.border.default,
          },
        ]}
      >
        <TextInput
          accessibilityHint={accessibilityHint}
          accessibilityLabel={message ? `${label}. ${message}` : label}
          onBlur={(event) => {
            setIsFocused(false);
            if (validationTrigger === "blur") setHasTriggeredValidation(true);
            onBlur?.(event);
          }}
          onChangeText={(value) => {
            if (validationTrigger === "change") setHasTriggeredValidation(true);
            onChangeText?.(value);
          }}
          onFocus={(event) => {
            setIsFocused(true);
            onFocus?.(event);
          }}
          placeholderTextColor={colors.text.placeholder}
          style={[styles.textInput, { color: colors.text.primary }]}
          {...inputProps}
        />
        {passwordVisibility ? (
          <Pressable
            accessibilityLabel={passwordVisibility.label}
            accessibilityRole="button"
            onPress={passwordVisibility.onPress}
            style={styles.visibilityButton}
          >
            <SymbolView
              name={
                passwordVisibility.isVisible
                  ? { android: "visibility_off", ios: "eye.slash", web: "visibility_off" }
                  : { android: "visibility", ios: "eye", web: "visibility" }
              }
              size={control.iconSize}
              tintColor={colors.text.secondary}
            />
          </Pressable>
        ) : null}
      </View>
      {message ? (
        <Text variant="caption" color={messageColor}>
          {message}
        </Text>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  wrapper: { gap: space[1] },
  input: {
    alignItems: "center",
    borderRadius: radius.control,
    borderWidth: 1,
    flexDirection: "row",
    minHeight: control.minHeight,
    paddingHorizontal: control.horizontalPadding,
  },
  textInput: {
    borderWidth: 0,
    flex: 1,
    fontFamily: typography.family.system,
    fontSize: typography.size.bodyLarge,
    minHeight: control.minHeight,
    outlineStyle: "solid",
    outlineWidth: 0,
  },
  visibilityButton: {
    alignItems: "center",
    height: control.minHeight,
    justifyContent: "center",
    marginRight: -space[2],
    width: control.minHeight,
  },
});
