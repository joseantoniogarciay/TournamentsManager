import { router, useFocusEffect } from "expo-router";
import { useCallback, useEffect, useState } from "react";
import {
  ActivityIndicator,
  Image,
  Platform,
  Pressable,
  ScrollView,
  StyleSheet,
  View,
  useWindowDimensions,
} from "react-native";

import { control, radius, space } from "@tournaments-manager/design-tokens";

import googleLogo from "../../../../assets/google-g.png";

import { GoogleAuthenticationError } from "@/features/federated-google/api";
import {
  authenticateLocalAccount,
  LocalAuthenticationError,
} from "@/features/local-authentication/api";
import { useGoogleAuthentication } from "@/features/federated-google/use-google-authentication";
import { useUsernameAvailability } from "@/features/registration/username-availability";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import { APIUnexpectedResponseError } from "@/api/fetch";
import { getCurrentLanguage, getTranslator } from "@/shared/i18n/locale";
import { TermsAcceptance } from "@/shared/legal/terms-acceptance";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { useSession } from "@/shared/session/session-provider";
import {
  Button,
  Card,
  DialogCloseButton,
  DisclosureIndicator,
  KeyboardAwareScrollView,
  ModalDialog,
  Screen,
  Text,
  TextField,
  useConfirmationDialog,
  useTabContentBottomPadding,
} from "@/shared/ui";

type AccountScreenProps = {
  sessionReplacementDestination?: "/account" | "/create-tournament";
};

type SocialLegalDocument = "privacy" | "terms";

export function AccountScreen({ sessionReplacementDestination = "/account" }: AccountScreenProps) {
  const t = getTranslator();
  const { show } = useFeedback();
  const { colors } = usePreferences();
  const { height: viewportHeight } = useWindowDimensions();
  const { completeSessionReplacement, signOut, user } = useSession();
  const { confirm } = useConfirmationDialog();
  const completeAccountSessionReplacement = useCallback(
    (nextUser: Parameters<typeof completeSessionReplacement>[0]) =>
      completeSessionReplacement(nextUser, sessionReplacementDestination),
    [completeSessionReplacement, sessionReplacementDestination],
  );
  const tabContentBottomPadding = useTabContentBottomPadding();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [passwordVisible, setPasswordVisible] = useState(false);
  const [isSigningIn, setIsSigningIn] = useState(false);
  const [showEmailError, setShowEmailError] = useState(false);
  const [showPasswordError, setShowPasswordError] = useState(false);
  const [googleUsername, setGoogleUsername] = useState("");
  const [googleUsernameSubmitted, setGoogleUsernameSubmitted] = useState(false);
  const [googleTermsAccepted, setGoogleTermsAccepted] = useState(false);
  const [socialLegalDocument, setSocialLegalDocument] = useState<SocialLegalDocument | null>(null);
  const { isValid: googleUsernameIsValid, status: googleUsernameAvailability } =
    useUsernameAvailability(googleUsername);
  const {
    chooseUsername,
    dismissPendingAccount,
    dismissError: dismissGoogleError,
    error: googleError,
    isAuthenticating: isGoogleAuthenticating,
    isConfigured: isGoogleConfigured,
    isPreparing: isGooglePreparing,
    isSubmitting: isGoogleSubmitting,
    prepare: prepareGoogleAuthentication,
    requiresUsername,
    start: startGoogleAuthentication,
  } = useGoogleAuthentication({
    locale: getCurrentLanguage(),
    onSession: completeAccountSessionReplacement,
  });
  const emailError = !isEmail(email) ? t("validation_email") : undefined;
  const passwordError = password ? undefined : t("validation_password_required");
  const googleUsernameError = !googleUsername.trim()
    ? t("validation_username_required")
    : !googleUsernameIsValid
      ? t("validation_username_format")
      : undefined;

  useEffect(() => {
    if (!googleError) return;
    if (googleError instanceof GoogleAuthenticationError) {
      show({
        kind: "generic-error",
        message: t(
          googleError.failure === "rate-limited"
            ? "account_google_rate_limited"
            : "account_existing_access_tip",
        ),
      });
    } else {
      const failure = getRequestFailure(googleError);
      show({ kind: failure.kind, message: t(failure.messageKey) });
    }
    dismissGoogleError();
  }, [dismissGoogleError, googleError, show, t]);

  useFocusEffect(
    useCallback(() => {
      if (!user) prepareGoogleAuthentication();
    }, [prepareGoogleAuthentication, user]),
  );

  const signIn = async () => {
    setShowEmailError(true);
    setShowPasswordError(true);
    if (emailError || passwordError) return;
    setIsSigningIn(true);
    try {
      const result = await authenticateLocalAccount({
        email,
        password,
        sessionTransport: Platform.OS === "web" ? "cookie" : "bearer",
      });
      if (result.kind === "pending-verification") {
        show({ kind: "success", message: t("account_login_verification_sent") });
        return;
      }
      completeAccountSessionReplacement(result.user);
    } catch (error) {
      if (error instanceof LocalAuthenticationError) {
        show({ kind: "generic-error", message: t("account_login_invalid_credentials") });
        return;
      }
      if (error instanceof APIUnexpectedResponseError && error.status === 429) {
        show({ kind: "generic-error", message: t("account_rate_limited") });
        return;
      }
      const failure = getRequestFailure(error);
      show({ kind: failure.kind, message: t(failure.messageKey) });
    } finally {
      setIsSigningIn(false);
    }
  };

  const createGoogleAccount = () => {
    setGoogleUsernameSubmitted(true);
    if (
      googleUsernameError ||
      googleUsernameAvailability === "checking" ||
      googleUsernameAvailability === "unavailable"
    ) {
      return;
    }
    if (!googleTermsAccepted) return;
    void chooseUsername(googleUsername as never);
  };

  const dismissGoogleAccountCreation = () => {
    if (isGoogleSubmitting) return;
    dismissPendingAccount();
    setGoogleUsername("");
    setGoogleUsernameSubmitted(false);
    setGoogleTermsAccepted(false);
    setSocialLegalDocument(null);
  };

  const confirmSignOut = () => {
    confirm({
      acceptLabel: t("account_logout"),
      cancelLabel: t("common_cancel"),
      description: t("account_logout_description"),
      onAccept: () => void signOut(),
      onCancel: () => undefined,
      title: t("account_logout_title"),
    });
  };

  if (user) {
    return (
      <Screen bottomInset="none" topInset="navigation-bar">
        <ScrollView
          contentContainerStyle={[styles.content, { paddingBottom: tabContentBottomPadding }]}
          showsVerticalScrollIndicator={false}
        >
          <View style={styles.authenticatedContent}>
            <Pressable
              accessibilityLabel={t("account_access_data_title")}
              accessibilityRole="button"
              onPress={() => router.push("/account/access" as never)}
              style={[styles.navigationRow, { borderColor: colors.border.default }]}
            >
              <Text variant="bodyLarge">{t("account_access_data_title")}</Text>
              <DisclosureIndicator />
            </Pressable>
            <Button
              label={t("account_logout")}
              onPress={confirmSignOut}
              secondarySurfaceColor={colors.surface.canvas}
              variant="secondary"
            />
          </View>
        </ScrollView>
      </Screen>
    );
  }

  return (
    <Screen bottomInset="none" topInset="navigation-bar">
      <KeyboardAwareScrollView
        contentContainerStyle={[styles.content, { paddingBottom: tabContentBottomPadding }]}
        showsVerticalScrollIndicator={false}
      >
        <Card>
          <View style={styles.form}>
            <Text variant="title">{t("account_sign_in_title")}</Text>
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
              passwordVisibility={{
                isVisible: passwordVisible,
                label: t(passwordVisible ? "password_hide" : "password_show"),
                onPress: () => setPasswordVisible(!passwordVisible),
              }}
              secureTextEntry={!passwordVisible}
              value={password}
            />
            <Pressable
              accessibilityLabel={t("password_recovery_title")}
              accessibilityRole="button"
              onPress={() => router.push("/account/forgot-password" as never)}
              style={styles.forgotPassword}
            >
              <Text color="secondary" style={styles.forgotPasswordText}>
                {t("password_recovery_title")}
              </Text>
            </Pressable>
            <Button
              disabled={isSigningIn}
              label={t("account_sign_in")}
              loading={isSigningIn}
              onPress={() => void signIn()}
            />
          </View>
        </Card>

        <Card>
          <View style={styles.form}>
            <Text variant="title">{t("account_social_title")}</Text>
            <Pressable
              accessibilityLabel={t(
                isGoogleConfigured ? "account_google_continue" : "account_google_unavailable",
              )}
              accessibilityRole="button"
              accessibilityState={{
                busy: isGooglePreparing || isGoogleAuthenticating,
                disabled: !isGoogleConfigured || isGooglePreparing || isGoogleAuthenticating,
              }}
              disabled={!isGoogleConfigured || isGooglePreparing || isGoogleAuthenticating}
              onPress={() => void startGoogleAuthentication()}
              style={[
                styles.googleButton,
                { borderColor: colors.border.default },
                !isGoogleConfigured || isGooglePreparing || isGoogleAuthenticating
                  ? styles.googleButtonDisabled
                  : undefined,
              ]}
            >
              {isGooglePreparing || isGoogleAuthenticating ? (
                <ActivityIndicator color={colors.indicator.default} />
              ) : (
                <Image source={googleLogo} style={styles.googleLogo} />
              )}
            </Pressable>
          </View>
        </Card>

        <ModalDialog
          dismissAccessibilityLabel={t("common_cancel")}
          onDismiss={dismissGoogleAccountCreation}
          visible={requiresUsername}
        >
          <View style={styles.form}>
            <Text variant="title">{t("account_google_new_account_title")}</Text>
            <Text color="secondary">{t("account_google_new_account_description")}</Text>
            <TextField
              autoCapitalize="none"
              autoCorrect={false}
              error={googleUsernameSubmitted ? googleUsernameError : undefined}
              feedback={usernameFeedback(t, googleUsernameAvailability)}
              label={t("account_username_label")}
              onBlur={() => setGoogleUsernameSubmitted(true)}
              onChangeText={(value) => setGoogleUsername(value.toLowerCase())}
              value={googleUsername}
            />
            <TermsAcceptance
              checked={googleTermsAccepted}
              onChange={setGoogleTermsAccepted}
              onOpenPrivacy={() => setSocialLegalDocument("privacy")}
              onOpenTerms={() => setSocialLegalDocument("terms")}
            />
            <Button
              disabled={
                !googleTermsAccepted ||
                googleUsernameAvailability === "checking" ||
                googleUsernameAvailability === "unavailable"
              }
              label={t("account_google_create_account")}
              loading={isGoogleSubmitting}
              onPress={createGoogleAccount}
            />
            <SocialLegalDocumentDialog
              document={socialLegalDocument}
              maxHeight={viewportHeight - space[10]}
              onDismiss={() => setSocialLegalDocument(null)}
            />
          </View>
        </ModalDialog>

        <View style={styles.register}>
          <Text color="secondary">{t("account_register_prompt")}</Text>
          <Button
            label={t("account_register")}
            onPress={() => router.push("/account/register")}
            secondarySurfaceColor={colors.surface.canvas}
            variant="secondary"
          />
        </View>
      </KeyboardAwareScrollView>
    </Screen>
  );
}

export default function AccountTabScreen() {
  return <AccountScreen />;
}

const styles = StyleSheet.create({
  authenticatedContent: { gap: space[6], marginHorizontal: space[5] },
  content: { gap: space[5] },
  form: { gap: space[4] },
  forgotPassword: { alignSelf: "flex-end", marginBottom: space[2] },
  forgotPasswordText: { textDecorationLine: "underline" },
  googleButton: {
    alignItems: "center",
    alignSelf: "center",
    borderRadius: radius.pill,
    borderWidth: 1,
    height: control.minHeight + space[1],
    justifyContent: "center",
    width: control.minHeight + space[1],
  },
  googleButtonDisabled: { opacity: 0.55 },
  googleLogo: { height: 22, width: 22 },
  legalDocument: { flexShrink: 1, gap: space[3] },
  legalDocumentContent: { gap: space[4] },
  legalDocumentHeader: { alignItems: "center", flexDirection: "row", gap: space[2] },
  legalDocumentSection: { gap: space[1] },
  legalDocumentScroll: { flexShrink: 1 },
  legalDocumentTitle: { flex: 1, flexShrink: 1 },
  register: { gap: space[3], marginHorizontal: space[5] },
  navigationRow: {
    alignItems: "center",
    borderBottomWidth: 1,
    borderTopWidth: 1,
    flexDirection: "row",
    justifyContent: "space-between",
    minHeight: control.minHeight + space[5],
  },
});

type SocialLegalDocumentDialogProps = {
  document: SocialLegalDocument | null;
  maxHeight: number;
  onDismiss: () => void;
};

function SocialLegalDocumentDialog({
  document,
  maxHeight,
  onDismiss,
}: SocialLegalDocumentDialogProps) {
  const t = getTranslator();
  const isTerms = document === "terms";
  const sections = isTerms ? termsSections : privacySections;

  return (
    <ModalDialog
      dismissAccessibilityLabel={t("common_close")}
      dialogStyle={{ maxHeight }}
      onDismiss={onDismiss}
      visible={document !== null}
    >
      <View style={styles.legalDocument}>
        <View style={styles.legalDocumentHeader}>
          <Text style={styles.legalDocumentTitle} variant="title">
            {t(isTerms ? "terms_of_use_title" : "privacy_policy_title")}
          </Text>
          <DialogCloseButton accessibilityLabel={t("common_close")} onPress={onDismiss} />
        </View>
        <ScrollView
          contentContainerStyle={styles.legalDocumentContent}
          showsVerticalScrollIndicator={false}
          style={styles.legalDocumentScroll}
        >
          <Text color="secondary">
            {t(isTerms ? "terms_of_use_updated" : "privacy_policy_updated")}
          </Text>
          {isTerms ? null : <Text>{t("privacy_policy_intro")}</Text>}
          {sections.map(([title, body]) => (
            <View key={title} style={styles.legalDocumentSection}>
              <Text variant="bodyLarge">{t(title)}</Text>
              <Text color="secondary">{t(body)}</Text>
            </View>
          ))}
        </ScrollView>
      </View>
    </ModalDialog>
  );
}

const termsSections = [
  ["terms_of_use_service_title", "terms_of_use_service_body"],
  ["terms_of_use_account_title", "terms_of_use_account_body"],
  ["terms_of_use_content_title", "terms_of_use_content_body"],
  ["terms_of_use_changes_title", "terms_of_use_changes_body"],
  ["terms_of_use_law_title", "terms_of_use_law_body"],
] as const;

const privacySections = [
  ["privacy_policy_controller_title", "privacy_policy_controller_body"],
  ["privacy_policy_data_title", "privacy_policy_data_body"],
  ["privacy_policy_purposes_title", "privacy_policy_purposes_body"],
  ["privacy_policy_sharing_title", "privacy_policy_sharing_body"],
  ["privacy_policy_retention_title", "privacy_policy_retention_body"],
  ["privacy_policy_public_title", "privacy_policy_public_body"],
  ["privacy_policy_minors_title", "privacy_policy_minors_body"],
  ["privacy_policy_rights_title", "privacy_policy_rights_body"],
  ["privacy_policy_changes_title", "privacy_policy_changes_body"],
] as const;

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
