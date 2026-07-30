import { useState } from "react";
import { ScrollView, StyleSheet, View } from "react-native";

import { space } from "@tournaments-manager/design-tokens";

import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getTranslator } from "@/shared/i18n/locale";
import { Button, Card, Screen, Text, TextField } from "@/shared/ui";
import { useUsernameAvailability } from "@/features/registration/username-availability";

export default function RegisterScreen() {
  const t = getTranslator();
  const { show } = useFeedback();
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [submitted, setSubmitted] = useState(false);
  const { isValid: usernameIsValid, status: usernameAvailability } =
    useUsernameAvailability(username);
  const usernameError = !username.trim()
    ? t("validation_username_required")
    : !usernameIsValid
      ? t("validation_username_format")
      : undefined;
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
              feedback={usernameFeedback(t, usernameAvailability)}
              label={t("account_username_label")}
              autoCapitalize="none"
              autoCorrect={false}
              onBlur={() => setSubmitted(true)}
              onChangeText={(value) => setUsername(value.toLowerCase())}
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

function usernameFeedback(
  t: ReturnType<typeof getTranslator>,
  status: ReturnType<typeof useUsernameAvailability>["status"],
) {
  switch (status) {
    case "checking":
      return { message: t("account_username_checking"), tone: "help" as const };
    case "available":
      return { message: t("account_username_available"), tone: "success" as const };
    case "unavailable":
      return { message: t("account_username_unavailable"), tone: "help" as const };
    case "rate-limited":
      return { message: t("account_username_rate_limited"), tone: "help" as const };
    case "error":
      return { message: t("account_username_check_error"), tone: "help" as const };
    default:
      return undefined;
  }
}
