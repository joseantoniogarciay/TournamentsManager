import { router } from "expo-router";
import { Pressable, StyleSheet } from "react-native";

import { getTranslator } from "@/shared/i18n/locale";
import { Text } from "@/shared/ui";

export function TermsOfUseLink() {
  const t = getTranslator();

  return (
    <Pressable
      accessibilityLabel={t("terms_of_use_link")}
      accessibilityRole="link"
      onPress={() => router.push("/terms-of-use" as never)}
    >
      <Text color="secondary" style={styles.label}>
        {t("terms_of_use_link")}
      </Text>
    </Pressable>
  );
}

const styles = StyleSheet.create({ label: { textDecorationLine: "underline" } });
