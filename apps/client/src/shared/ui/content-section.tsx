import type { PropsWithChildren } from "react";
import { StyleSheet, View, type StyleProp, type ViewStyle } from "react-native";

import { space } from "@tournaments-manager/design-tokens";

import { usePreferences } from "@/shared/preferences/preferences-provider";

type ContentSectionProps = PropsWithChildren<{
  separated?: boolean;
  style?: StyleProp<ViewStyle>;
}>;

/** Open content owns its outer margin; Screen and scroll containers do not add it. */
export function ContentSection({ children, separated = false, style }: ContentSectionProps) {
  const { colors } = usePreferences();
  return (
    <View
      style={[
        styles.section,
        separated && { borderTopWidth: 1, borderTopColor: colors.border.default },
        style,
      ]}
    >
      {children}
    </View>
  );
}

const styles = StyleSheet.create({
  section: { marginHorizontal: space[5], paddingVertical: space[3], gap: space[3] },
});
