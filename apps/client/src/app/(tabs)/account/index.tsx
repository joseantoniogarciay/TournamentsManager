import { router } from "expo-router";
import { useState } from "react";
import { ScrollView, StyleSheet, View } from "react-native";

import { space } from "@tournaments-manager/design-tokens";

import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getTranslator } from "@/shared/i18n/locale";
import { Button, Card, Screen, Text, TextField } from "@/shared/ui";

export default function AccountScreen() {
  const t = getTranslator();
  const { show } = useFeedback();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showEmailError, setShowEmailError] = useState(false);
  const [showPasswordError, setShowPasswordError] = useState(false);
  const emailError = !isEmail(email) ? t("validation_email") : undefined;
  const passwordError = password ? undefined : t("validation_password_required");

  const signIn = () => {
    setShowEmailError(true);
    setShowPasswordError(true);
    if (emailError || passwordError) return;
    show({ kind: "generic-error", message: t("account_local_login_unavailable") });
  };

  return (
    <Screen topInset="navigation-bar">
      <ScrollView contentContainerStyle={styles.content} showsVerticalScrollIndicator={false}>
        <Card>
          <View style={styles.form}>
            <Text variant="title">{t("account_sign_in_title")}</Text>
            <Text color="secondary">{t("account_sign_in_description")}</Text>
            <TextField
              autoCapitalize="none"
              autoComplete="email"
              error={showEmailError ? emailError : undefined}
              keyboardType="email-address"
              label={t("account_email_label")}
              onBlur={() => setShowEmailError(true)}
              onChangeText={setEmail}
              value={email}
            />
            <TextField
              autoComplete="current-password"
              error={showPasswordError ? passwordError : undefined}
              label={t("account_password_label")}
              onBlur={() => setShowPasswordError(true)}
              onChangeText={setPassword}
              secureTextEntry
              value={password}
            />
            <Button label={t("account_sign_in")} onPress={signIn} />
          </View>
        </Card>

        <Card>
          <View style={styles.form}>
            <Text variant="title">{t("account_social_title")}</Text>
            <Text color="secondary">{t("account_social_description")}</Text>
            <Button
              disabled
              label={t("account_google_unavailable")}
              onPress={() => undefined}
              variant="secondary"
            />
          </View>
        </Card>

        <View style={styles.register}>
          <Text color="secondary">{t("account_register_prompt")}</Text>
          <Button
            label={t("account_register")}
            onPress={() => router.push("/account/register")}
            variant="secondary"
          />
        </View>
      </ScrollView>
    </Screen>
  );
}

const styles = StyleSheet.create({
  content: { gap: space[5], paddingBottom: space[12] },
  form: { gap: space[4] },
  register: { gap: space[3], marginHorizontal: space[5] },
});

function isEmail(value: string) {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value);
}
