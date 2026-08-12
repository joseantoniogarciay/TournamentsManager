import { router } from "expo-router";
import { Pressable, StyleSheet } from "react-native";

import { getTranslator } from "@/shared/i18n/locale";
import { Text } from "@/shared/ui";

export function PrivacyPolicyLink() {
  const t = getTranslator();

  return (
    <Pressable
      accessibilityLabel={t("privacy_policy_link")}
      accessibilityRole="link"
      onPress={() => router.push("/privacy-policy" as never)}
    >
      <Text color="secondary" style={styles.label}>
        {t("privacy_policy_link")}
      </Text>
    </Pressable>
  );
}

const styles = StyleSheet.create({ label: { textDecorationLine: "underline" } });
