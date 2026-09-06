import { router } from "expo-router";
import { Pressable, StyleSheet, View } from "react-native";

import { color } from "@tournaments-manager/design-tokens";

import { getTranslator } from "@/shared/i18n/locale";
import { Text } from "@/shared/ui";

type TermsAcceptanceProps = {
  checked: boolean;
  onChange: (checked: boolean) => void;
};

/** Casilla contractual común para el alta local y la primera alta con Google. */
export function TermsAcceptance({ checked, onChange }: TermsAcceptanceProps) {
  const t = getTranslator();

  return (
    <View style={styles.container}>
      <Pressable
        accessibilityLabel={t("account_terms_acceptance")}
        accessibilityRole="checkbox"
        accessibilityState={{ checked }}
        onPress={() => onChange(!checked)}
        style={styles.checkboxTarget}
      >
        <View style={[styles.checkbox, checked ? styles.checkboxSelected : undefined]} />
      </Pressable>
      <Text color="secondary" style={styles.copy}>
        {t("account_terms_acceptance_prefix")}
        <Text
          accessibilityLabel={t("terms_of_use_link")}
          accessibilityRole="link"
          color="secondary"
          onPress={() => router.push("/terms-of-use" as never)}
          style={styles.termsLink}
        >
          {t("terms_of_use_link")}
        </Text>
        {t("account_terms_acceptance_privacy_prefix")}
        <Text
          accessibilityLabel={t("privacy_policy_link")}
          accessibilityRole="link"
          color="secondary"
          onPress={() => router.push("/privacy-policy" as never)}
          style={styles.termsLink}
        >
          {t("privacy_policy_link")}
        </Text>
        {t("account_terms_acceptance_suffix")}
      </Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { alignItems: "center", flexDirection: "row", minHeight: 44 },
  checkboxTarget: { alignItems: "center", height: 44, justifyContent: "center", width: 44 },
  checkbox: { borderWidth: 1, height: 20, width: 20 },
  checkboxSelected: { backgroundColor: color.brand.primary },
  copy: { flex: 1 },
  termsLink: { textDecorationLine: "underline" },
});
