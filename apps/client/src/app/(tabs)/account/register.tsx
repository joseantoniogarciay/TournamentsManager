import { router } from "expo-router";
import { useState } from "react";
import { Pressable, StyleSheet, View } from "react-native";

import { color, space } from "@tournaments-manager/design-tokens";

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
import { APIUnexpectedResponseError } from "@/api/fetch";
import { PrivacyPolicyLink } from "@/shared/legal/privacy-policy-link";
import { TermsOfUseLink } from "@/shared/legal/terms-of-use-link";
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
  const [acceptedTerms, setAcceptedTerms] = useState(false);
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
      !acceptedTerms ||
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
        termsVersion: "2026-08-22",
        username,
      });
      if (draft) await clearLocalLeagueDraft();
      show({ kind: "success", message: t("account_registration_email_sent") });
      router.replace("/account");
    } catch (error) {
      if (error instanceof APIUnexpectedResponseError && error.status === 429) {
        show({ kind: "generic-error", message: t("account_rate_limited") });
        return;
      }
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
            <Pressable
              accessibilityRole="checkbox"
              accessibilityState={{ checked: acceptedTerms }}
              onPress={() => setAcceptedTerms((value) => !value)}
              style={styles.terms}
            >
              <View
                style={[styles.checkbox, acceptedTerms ? styles.checkboxSelected : undefined]}
              />
              <Text color="secondary">{t("account_terms_acceptance")}</Text>
            </Pressable>
            <TermsOfUseLink />
            <Button
              disabled={
                !acceptedTerms ||
                usernameAvailability === "checking" ||
                usernameAvailability === "unavailable"
              }
              label={t("account_register")}
              loading={isSubmitting}
              onPress={() => void register()}
            />
          </View>
        </Card>
        <View style={styles.privacyPolicyLink}>
          <PrivacyPolicyLink />
        </View>
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
  terms: { alignItems: "center", flexDirection: "row", gap: space[2], minHeight: 44 },
  checkbox: { borderWidth: 1, height: 20, width: 20 },
  checkboxSelected: { backgroundColor: color.brand.primary },
  privacyPolicyLink: {
    alignSelf: "flex-end",
    marginHorizontal: space[5],
    marginTop: space[12] + space[5],
  },
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
