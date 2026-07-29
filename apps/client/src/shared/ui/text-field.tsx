import { StyleSheet, TextInput, View, type TextInputProps } from "react-native";

import { control, radius, space, typography } from "@tournaments-manager/design-tokens";

import { usePreferences } from "@/shared/preferences/preferences-provider";

import { Text } from "./text";

type Props = TextInputProps & { label: string; error?: string };

export function TextField({ label, error, accessibilityHint, ...inputProps }: Props) {
  const { colors } = usePreferences();
  return (
    <View style={styles.wrapper}>
      <Text variant="bodyLarge">{label}</Text>
      <TextInput
        accessibilityHint={accessibilityHint}
        accessibilityLabel={error ? `${label}. ${error}` : label}
        placeholderTextColor={colors.text.placeholder}
        style={[
          styles.input,
          {
            borderColor: error ? colors.border.error : colors.border.default,
            color: colors.text.primary,
          },
        ]}
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
    borderRadius: radius.control,
    borderWidth: 1,
    fontFamily: typography.family.system,
    fontSize: typography.size.bodyLarge,
    minHeight: control.minHeight,
    paddingHorizontal: control.horizontalPadding,
  },
});
