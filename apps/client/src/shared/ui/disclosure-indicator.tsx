import { SymbolView } from "expo-symbols";
import { Platform, StyleSheet, View } from "react-native";

import { space } from "@tournaments-manager/design-tokens";

import { usePreferences } from "@/shared/preferences/preferences-provider";

import { WebIcon } from "./web-icon";

const disclosureContainerSize = space[6];
const disclosureIconSize = disclosureContainerSize * 0.8;

/** Indicador visual de que toda la fila abre su siguiente nivel. */
export function DisclosureIndicator() {
  const { colors } = usePreferences();

  return (
    <View accessible={false} style={styles.container}>
      {Platform.OS === "web" ? (
        <WebIcon color={colors.text.secondary} name="chevronRight" size={disclosureIconSize} />
      ) : (
        <SymbolView
          name="chevron.right"
          size={disclosureIconSize}
          tintColor={colors.text.secondary}
        />
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    alignItems: "center",
    height: disclosureContainerSize,
    justifyContent: "center",
    width: disclosureContainerSize,
  },
});
