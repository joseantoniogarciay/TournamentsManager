import { router } from "expo-router";
import { useCallback, useEffect, useState } from "react";
import { StyleSheet, View } from "react-native";

import { space } from "@tournaments-manager/design-tokens";

import {
  getAccountAccessMethods,
  reauthenticateWithGoogle,
  reauthenticateWithPassword,
  setAccountPassword,
} from "@/features/account-access/api";
import { useGoogleIdentityProof } from "@/features/federated-google/use-google-identity-proof";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import { getTranslator } from "@/shared/i18n/locale";
import {
  Button,
  Card,
  KeyboardAwareScrollView,
  Screen,
  Text,
  TextField,
  useTabContentBottomPadding,
} from "@/shared/ui";

export default function AccountPasswordScreen() {
  const t = getTranslator();
  const { show } = useFeedback();
  const [currentPassword, setCurrentPassword] = useState("");
  const [password, setPassword] = useState("");
  const [hasPassword, setHasPassword] = useState<boolean | null>(null);
  const [googleTicket, setGoogleTicket] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const tabContentBottomPadding = useTabContentBottomPadding();
  useEffect(() => {
    void getAccountAccessMethods().then((access) => setHasPassword(access.methods.password));
  }, []);
  const onGoogleProof = useCallback(async (challenge: { id: string }, idToken: string) => {
    setGoogleTicket(await reauthenticateWithGoogle(challenge.id, idToken, "set-local-password"));
  }, []);
  const proof = useGoogleIdentityProof(onGoogleProof);
  useEffect(() => {
    if (hasPassword === false) proof.prepare();
  }, [hasPassword, proof.prepare]);
  const valid =
    password.length >= 8 && (hasPassword ? currentPassword.length >= 8 : Boolean(googleTicket));

  const submit = async () => {
    if (!valid || submitting) return;
    setSubmitting(true);
    try {
      const ticket = hasPassword
        ? await reauthenticateWithPassword(currentPassword, "set-local-password")
        : googleTicket;
      if (!ticket) return;
      await setAccountPassword(ticket, password);
      router.replace("/account/access" as never);
    } catch (error) {
      const failure = getRequestFailure(error);
      show({ kind: failure.kind, message: t(failure.messageKey) });
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Screen bottomInset="none" topInset="navigation-bar">
      <KeyboardAwareScrollView
        contentContainerStyle={[styles.content, { paddingBottom: tabContentBottomPadding }]}
      >
        <Card>
          <View style={styles.form}>
            {hasPassword ? (
              <Text color="secondary">{t("account_password_change_description")}</Text>
            ) : null}
            {hasPassword === false && !googleTicket ? (
              <Text color="secondary">{t("account_password_add_google_description")}</Text>
            ) : null}
            {hasPassword ? (
              <TextField
                autoComplete="current-password"
                label={t("account_password_current_label")}
                onChangeText={setCurrentPassword}
                secureTextEntry
                value={currentPassword}
              />
            ) : null}
            {hasPassword === false && !googleTicket ? (
              <Button
                disabled={!proof.isConfigured}
                label={t("account_password_google_reauthenticate")}
                loading={proof.isLoading}
                onPress={() => void proof.start()}
                variant="secondary"
              />
            ) : null}
            <TextField
              autoComplete="new-password"
              label={t("account_password_new_label")}
              onChangeText={setPassword}
              secureTextEntry
              value={password}
            />
            <Button
              disabled={!valid}
              label={t("account_password_save")}
              loading={submitting}
              onPress={() => void submit()}
            />
          </View>
        </Card>
      </KeyboardAwareScrollView>
    </Screen>
  );
}

const styles = StyleSheet.create({ content: { gap: space[5] }, form: { gap: space[4] } });
