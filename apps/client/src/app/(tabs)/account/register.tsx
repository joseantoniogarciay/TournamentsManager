import { useState } from "react";
import { ScrollView, StyleSheet, View } from "react-native";

import { space } from "@tournaments-manager/design-tokens";

import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getTranslator } from "@/shared/i18n/locale";
import { Button, Card, Screen, Text, TextField } from "@/shared/ui";

export default function RegisterScreen() {
  const t = getTranslator();
  const { show } = useFeedback();
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [submitted, setSubmitted] = useState(false);
  const usernameError = username.trim() ? undefined : t("validation_username_required");
  const emailError = !isEmail(email) ? t("validation_email") : undefined;
  const passwordError = password ? undefined : t("validation_password_required");
  const register = () => {
    setSubmitted(true);
    if (usernameError || emailError || passwordError) return;
    show({ kind: "generic-error", message: t("account_registration_unavailable") });
  };

  return (
    <Screen bottomInset="none" topInset="navigation-bar">
      <ScrollView contentContainerStyle={styles.content} showsVerticalScrollIndicator={false}>
        <Card>
          <View style={styles.form}>
            <Text color="secondary">{t("account_register_description")}</Text>
            <TextField
              error={submitted ? usernameError : undefined}
              label={t("account_username_label")}
              onBlur={() => setSubmitted(true)}
              onChangeText={setUsername}
              value={username}
            />
            <TextField
              autoCapitalize="none"
              autoComplete="email"
              error={submitted ? emailError : undefined}
              keyboardType="email-address"
              label={t("account_email_label")}
              onBlur={() => setSubmitted(true)}
              onChangeText={setEmail}
              value={email}
            />
            <TextField
              autoComplete="new-password"
              error={submitted ? passwordError : undefined}
              label={t("account_password_label")}
              onBlur={() => setSubmitted(true)}
              onChangeText={setPassword}
              secureTextEntry
              value={password}
            />
            <Button label={t("account_register")} onPress={register} />
          </View>
        </Card>
      </ScrollView>
    </Screen>
  );
}

const styles = StyleSheet.create({
  content: { paddingBottom: space[12] },
  form: { gap: space[4] },
});

function isEmail(value: string) {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value);
}
