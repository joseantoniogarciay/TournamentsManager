import { router } from "expo-router";
import { useState } from "react";
import { StyleSheet, View } from "react-native";

import { space } from "@tournaments-manager/design-tokens";

import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getCurrentLanguage, getTranslator } from "@/shared/i18n/locale";
import {
  Button,
  Card,
  InteractionBlocker,
  KeyboardAwareScrollView,
  Screen,
  Text,
  TextField,
  useTabContentBottomPadding,
} from "@/shared/ui";
import { useUsernameAvailability } from "@/features/registration/username-availability";
import { registerLocalAccountRequest } from "@/features/registration/api";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import {
  clearLocalLeagueDraft,
  getLocalLeagueDraft,
  toLeagueInput,
} from "@/features/league-creation/draft";

export default function RegisterScreen() {
  const t = getTranslator();
  const { show } = useFeedback();
  const tabContentBottomPadding = useTabContentBottomPadding();
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [passwordVisible, setPasswordVisible] = useState(false);
  const [hasAttemptedSubmit, setHasAttemptedSubmit] = useState(false);
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
    setHasAttemptedSubmit(true);
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
      const draft = toLeagueInput(await getLocalLeagueDraft());
      await registerLocalAccountRequest({
        ...(draft ? { draft } : {}),
        email: email.trim(),
        locale: getCurrentLanguage(),
        password,
        username,
      });
      if (draft) await clearLocalLeagueDraft();
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
      <KeyboardAwareScrollView
        contentContainerStyle={[styles.content, { paddingBottom: tabContentBottomPadding }]}
        showsVerticalScrollIndicator={false}
      >
        <Card>
          <View style={styles.form}>
            <Text color="secondary">{t("account_register_description")}</Text>
            <TextField
              error={usernameError}
              feedback={usernameFeedback(t, usernameAvailability)}
              label={t("account_username_label")}
              autoCapitalize="none"
              autoCorrect={false}
              onChangeText={(value) => setUsername(value.toLowerCase())}
              validationSubmitted={hasAttemptedSubmit}
              validationTrigger="change"
              value={username}
            />
            <TextField
              autoCapitalize="none"
              autoComplete="email"
              error={emailError}
              keyboardType="email-address"
              label={t("account_email_label")}
              onChangeText={setEmail}
              validationSubmitted={hasAttemptedSubmit}
              validationTrigger="blur"
              value={email}
            />
            <TextField
              autoComplete="new-password"
              error={passwordError}
              label={t("account_password_label")}
              onChangeText={setPassword}
              passwordVisibility={{
                isVisible: passwordVisible,
                label: t(passwordVisible ? "password_hide" : "password_show"),
                onPress: () => setPasswordVisible(!passwordVisible),
              }}
              secureTextEntry={!passwordVisible}
              validationSubmitted={hasAttemptedSubmit}
              validationTrigger="change"
              value={password}
            />
            {password.length >= 8 ? (
              <Text color="secondary">
                {password.length < 15 ? t("password_strength_ok") : t("password_strength_strong")}
              </Text>
            ) : null}
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
      </KeyboardAwareScrollView>
      {isSubmitting ? (
        <InteractionBlocker accessibilityLabel={t("account_registration_submitting")} />
      ) : null}
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
