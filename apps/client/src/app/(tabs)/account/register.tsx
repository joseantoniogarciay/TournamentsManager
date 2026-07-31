import { router } from "expo-router";
import { useState } from "react";
import { ScrollView, StyleSheet, View } from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";

import { space } from "@tournaments-manager/design-tokens";

import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getCurrentLanguage, getTranslator } from "@/shared/i18n/locale";
import { Button, Card, Screen, Text, TextField } from "@/shared/ui";
import { useUsernameAvailability } from "@/features/registration/username-availability";
import { registerLocalAccountRequest } from "@/features/registration/api";
import { getRequestFailure } from "@/shared/feedback/request-failure";

export default function RegisterScreen() {
  const t = getTranslator();
  const { show } = useFeedback();
  const insets = useSafeAreaInsets();
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [passwordVisible, setPasswordVisible] = useState(false);
  const [submitted, setSubmitted] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { isValid: usernameIsValid, status: usernameAvailability } =
    useUsernameAvailability(username);
  const usernameError = !username.trim()
    ? t("validation_username_required")
    : !usernameIsValid
      ? t("validation_username_format")
      : undefined;
  const emailError = !isEmail(email) ? t("validation_email") : undefined;
  const passwordError = !password
    ? t("validation_password_required")
    : password.length < 8
      ? t("validation_password_length")
      : undefined;
  const register = async () => {
    setSubmitted(true);
    if (
      usernameError ||
      emailError ||
      passwordError ||
      usernameAvailability === "checking" ||
      usernameAvailability === "unavailable"
    ) {
      return;
    }

    setIsSubmitting(true);
    try {
      await registerLocalAccountRequest({
        email: email.trim(),
        locale: getCurrentLanguage(),
        password,
        username,
      });
      show({ kind: "success", message: t("account_registration_email_sent") });
      router.replace("/account");
    } catch (error) {
      const failure = getRequestFailure(error);
      show({ kind: failure.kind, message: t(failure.messageKey) });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Screen bottomInset="none" topInset="navigation-bar">
      <ScrollView
        contentContainerStyle={[styles.content, { paddingBottom: insets.bottom + space[12] }]}
        showsVerticalScrollIndicator={false}
      >
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
              passwordVisibility={{
                label: t(passwordVisible ? "password_hide" : "password_show"),
                onPress: () => setPasswordVisible(!passwordVisible),
              }}
              secureTextEntry={!passwordVisible}
              value={password}
            />
            <Text color="secondary">
              {password.length < 8
                ? t("password_strength_weak")
                : password.length < 15
                  ? t("password_strength_ok")
                  : t("password_strength_strong")}
            </Text>
            <Button
              disabled={
                usernameAvailability === "checking" || usernameAvailability === "unavailable"
              }
              label={t("account_register")}
              loading={isSubmitting}
              onPress={() => void register()}
            />
          </View>
        </Card>
      </ScrollView>
    </Screen>
  );
}

const styles = StyleSheet.create({
  content: {},
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
    case "network-error":
      return { message: t("common_network_error"), tone: "help" as const };
    case "error":
      return { message: t("common_request_error"), tone: "help" as const };
    default:
      return undefined;
  }
}
