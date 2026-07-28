import { StyleSheet, TextInput, View, type TextInputProps } from "react-native";

import { color, control, radius, space, typography } from "@tournaments-manager/design-tokens";

import { Text } from "./text";

type Props = TextInputProps & { label: string; error?: string };

export function TextField({ label, error, accessibilityHint, ...inputProps }: Props) {
  return (
    <View style={styles.wrapper}>
      <Text variant="bodyLarge">{label}</Text>
      <TextInput
        accessibilityHint={accessibilityHint}
        accessibilityLabel={error ? `${label}. Error: ${error}` : label}
        placeholderTextColor={color.text.placeholder}
        style={[styles.input, error ? styles.inputError : undefined]}
        {...inputProps}
      />
      {error ? (
        <Text variant="caption" color="error">
          {error}
        </Text>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  wrapper: { gap: space[1] },
  input: {
    borderColor: color.border.default,
    borderRadius: radius.control,
    borderWidth: 1,
    color: color.text.primary,
    fontFamily: typography.family.system,
    fontSize: typography.size.bodyLarge,
    minHeight: control.minHeight,
    paddingHorizontal: control.horizontalPadding,
  },
  inputError: { borderColor: color.border.error },
});
