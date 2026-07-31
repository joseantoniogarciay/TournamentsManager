import { router } from "expo-router";
import { useState } from "react";
import { Image, Pressable, ScrollView, StyleSheet, View } from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";

import { control, radius, space } from "@tournaments-manager/design-tokens";

import googleLogo from "../../../../assets/google-g.png";

import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { useSession } from "@/shared/session/session-provider";
import { Button, Card, Screen, Text, TextField } from "@/shared/ui";

export default function AccountScreen() {
  const t = getTranslator();
  const { show } = useFeedback();
  const { colors } = usePreferences();
  const { user } = useSession();
  const insets = useSafeAreaInsets();
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

  if (user) {
    return (
      <Screen bottomInset="none" topInset="navigation-bar">
        <View style={{ paddingBottom: insets.bottom + space[12] }}>
          <Card>
            <View style={styles.form}>
              <Text variant="title">{t("account_authenticated_title")}</Text>
              <Text color="secondary">
                {t("account_authenticated_mock").replace("{username}", user.username)}
              </Text>
            </View>
          </Card>
        </View>
      </Screen>
    );
  }

  return (
    <Screen bottomInset="none" topInset="navigation-bar">
      <ScrollView
        contentContainerStyle={[styles.content, { paddingBottom: insets.bottom + space[12] }]}
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
              secureTextEntry
              value={password}
            />
            <Button label={t("account_sign_in")} onPress={signIn} />
          </View>
        </Card>

        <Card>
          <View style={styles.form}>
            <Text variant="title">{t("account_social_title")}</Text>
            <Pressable
              accessibilityLabel={t("account_google_unavailable")}
              accessibilityRole="button"
              accessibilityState={{ disabled: true }}
              disabled
              onPress={() => undefined}
              style={[styles.googleButton, { borderColor: colors.border.default }]}
            >
              <Image source={googleLogo} style={styles.googleLogo} />
            </Pressable>
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
  content: { gap: space[5] },
  form: { gap: space[4] },
  googleButton: {
    alignItems: "center",
    alignSelf: "center",
    borderRadius: radius.pill,
    borderWidth: 1,
    height: control.minHeight + space[1],
    justifyContent: "center",
    opacity: 0.55,
    width: control.minHeight + space[1],
  },
  googleLogo: { height: 22, width: 22 },
  register: { gap: space[3], marginHorizontal: space[5] },
});

function isEmail(value: string) {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value);
}
