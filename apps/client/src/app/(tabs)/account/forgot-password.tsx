import { router } from "expo-router";
import { useState } from "react";
import { StyleSheet, View } from "react-native";
import { space } from "@tournaments-manager/design-tokens";
import { requestRecovery } from "@/features/password-recovery/api";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getTranslator } from "@/shared/i18n/locale";
import { Button, Card, KeyboardAwareScrollView, Screen, Text, TextField } from "@/shared/ui";

export default function ForgotPasswordScreen() {
  const t = getTranslator();
  const { show } = useFeedback();
  const [email, setEmail] = useState("");
  const [sending, setSending] = useState(false);
  const valid = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
  const submit = async () => {
    if (!valid) return;
    setSending(true);
    try {
      await requestRecovery(email.trim());
      show({ kind: "success", message: t("password_recovery_sent") });
      router.replace("/account");
    } catch {
      show({ kind: "generic-error", message: t("common_request_error") });
    } finally {
      setSending(false);
    }
  };
  return (
    <Screen topInset="navigation-bar">
      <KeyboardAwareScrollView>
        <Card>
          <View style={styles.form}>
            <Text color="secondary">{t("password_recovery_description")}</Text>
            <TextField
              autoCapitalize="none"
              autoComplete="email"
              error={email && !valid ? t("validation_email") : undefined}
              keyboardType="email-address"
              label={t("account_email_label")}
              onChangeText={setEmail}
              value={email}
            />
            <Button
              disabled={!valid}
              label={t("password_recovery_submit")}
              loading={sending}
              onPress={() => void submit()}
            />
          </View>
        </Card>
      </KeyboardAwareScrollView>
    </Screen>
  );
}
const styles = StyleSheet.create({ form: { gap: space[4] } });
