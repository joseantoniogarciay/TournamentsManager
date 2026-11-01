import type { PropsWithChildren } from "react";
import { StyleSheet, View, type StyleProp, type ViewStyle } from "react-native";

import { space } from "@tournaments-manager/design-tokens";

import { usePreferences } from "@/shared/preferences/preferences-provider";

type ListRowProps = PropsWithChildren<{ style?: StyleProp<ViewStyle> }>;

/** A row is a separator and spacing, never another card around a list item. */
export function ListRow({ children, style }: ListRowProps) {
  const { colors } = usePreferences();
  return (
    <View style={[styles.row, { borderBottomColor: colors.border.default }, style]}>
      {children}
    </View>
  );
}

const styles = StyleSheet.create({
  row: { marginHorizontal: space[5], paddingVertical: space[3], borderBottomWidth: 1 },
});
