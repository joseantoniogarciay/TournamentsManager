import { SymbolView } from "expo-symbols";
import { Pressable, StyleSheet, TextInput, View, type TextInputProps } from "react-native";

import { control, radius, space, typography } from "@tournaments-manager/design-tokens";

import { usePreferences } from "@/shared/preferences/preferences-provider";

import { Text } from "./text";

type Props = TextInputProps & {
  label: string;
  error?: string;
  feedback?: { message: string; tone: "help" | "success" };
  passwordVisibility?: { isVisible: boolean; label: string; onPress: () => void };
};

export function TextField({
  label,
  error,
  feedback,
  accessibilityHint,
  passwordVisibility,
  ...inputProps
}: Props) {
  const { colors } = usePreferences();
  const message = error ?? feedback?.message;
  const messageColor = error ? "error" : feedback?.tone === "success" ? "success" : "secondary";
  return (
    <View style={styles.wrapper}>
      <Text variant="bodyLarge">{label}</Text>
      <View
        style={[styles.input, { borderColor: error ? colors.border.error : colors.border.default }]}
      >
        <TextInput
          accessibilityHint={accessibilityHint}
          accessibilityLabel={message ? `${label}. ${message}` : label}
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
    flex: 1,
    fontFamily: typography.family.system,
    fontSize: typography.size.bodyLarge,
    minHeight: control.minHeight,
  },
  visibilityButton: {
    alignItems: "center",
    height: control.minHeight,
    justifyContent: "center",
    width: control.minHeight,
  },
});
