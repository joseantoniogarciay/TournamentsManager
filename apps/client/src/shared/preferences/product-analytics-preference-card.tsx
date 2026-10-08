import { Pressable, StyleSheet, Switch, View } from "react-native";

import { control, space } from "@tournaments-manager/design-tokens";

import { getTranslator } from "@/shared/i18n/locale";
import { Card, Text } from "@/shared/ui";

import { usePreferences } from "./preferences-provider";

type ProductAnalyticsPreferenceCardProps = {
  onEnabled?: () => void;
};

/** Reusable consent control; it never controls essential reliability capture. */
export function ProductAnalyticsPreferenceCard({ onEnabled }: ProductAnalyticsPreferenceCardProps) {
  const t = getTranslator();
  const { productAnalyticsEnabled, setProductAnalyticsEnabled } = usePreferences();

  const updateProductAnalytics = (enabled: boolean) => {
    const wasEnabled = productAnalyticsEnabled;
    setProductAnalyticsEnabled(enabled);
    if (enabled && !wasEnabled) onEnabled?.();
  };

  return (
    <Card>
      <View style={styles.row}>
        <View style={styles.copy}>
          <Text variant="title">{t("product_analytics_home_title")}</Text>
          <Text color="secondary">{t("product_analytics_home_description")}</Text>
        </View>
        <Pressable
          accessibilityRole="switch"
          accessibilityState={{ checked: productAnalyticsEnabled }}
          aria-checked={productAnalyticsEnabled}
          accessibilityHint={t("product_analytics_switch_hint")}
          accessibilityLabel={t("product_analytics_switch_label")}
          onPress={() => updateProductAnalytics(!productAnalyticsEnabled)}
          style={styles.switchTarget}
        >
          <View
            aria-hidden
            pointerEvents="none"
            accessible={false}
            accessibilityElementsHidden
            importantForAccessibility="no-hide-descendants"
          >
            <Switch accessible={false} tabIndex={-1} value={productAnalyticsEnabled} />
          </View>
        </Pressable>
      </View>
    </Card>
  );
}

const styles = StyleSheet.create({
  switchTarget: {
    alignItems: "center",
    justifyContent: "center",
    minHeight: control.minHeight,
    minWidth: control.minHeight,
  },
  copy: { flex: 1, gap: space[1] },
  row: { alignItems: "center", flexDirection: "row", gap: space[4] },
});
