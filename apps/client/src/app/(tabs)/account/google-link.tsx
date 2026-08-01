import { router } from "expo-router";
import { useCallback, useEffect, useState } from "react";
import { StyleSheet, View } from "react-native";

import { space } from "@tournaments-manager/design-tokens";

import {
  GoogleLinkError,
  getAccountAccessMethods,
  linkGoogle,
  reauthenticateWithGoogle,
  reauthenticateWithPassword,
} from "@/features/account-access/api";
import { useGoogleIdentityProof } from "@/features/federated-google/use-google-identity-proof";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import { getTranslator } from "@/shared/i18n/locale";
import { Button, Card, Screen, Text, TextField } from "@/shared/ui";

type Stage = "reauthenticate" | "prepare-google" | "connecting";

export default function GoogleLinkScreen() {
  const t = getTranslator();
  const { show } = useFeedback();
  const [hasPassword, setHasPassword] = useState<boolean | null>(null);
  const [password, setPassword] = useState("");
  const [ticket, setTicket] = useState<string | null>(null);
  const [stage, setStage] = useState<Stage>("reauthenticate");

  useEffect(() => {
    void getAccountAccessMethods()
      .then((access) => setHasPassword(access.methods.password))
      .catch(() => {
        router.back();
      });
  }, []);

  const onProof = useCallback(
    async (challenge: { id: string }, idToken: string) => {
      if (!ticket) {
        const nextTicket = await reauthenticateWithGoogle(challenge.id, idToken);
        setTicket(nextTicket);
        setStage("prepare-google");
        return;
      }
      setStage("connecting");
      await linkGoogle(ticket, challenge.id, idToken);
      router.back();
    },
    [ticket],
  );
  const proof = useGoogleIdentityProof(onProof);

  useEffect(() => {
    if (!proof.error) return;
    const message =
      proof.error instanceof GoogleLinkError
        ? t(
            proof.error.reason === "conflict"
              ? "account_google_link_conflict"
              : "account_google_link_expired",
          )
        : t(getRequestFailure(proof.error).messageKey);
    show({ kind: "generic-error", message });
    router.back();
  }, [proof.error, show, t]);

  const confirmPassword = async () => {
    if (password.length < 8) return;
    try {
      setTicket(await reauthenticateWithPassword(password));
      setStage("prepare-google");
    } catch (error) {
      show({
        kind: "generic-error",
        message: t(
          error instanceof GoogleLinkError
            ? "account_google_link_expired"
            : getRequestFailure(error).messageKey,
        ),
      });
      router.back();
    }
  };

  const startGoogle = () => void proof.start();
  const googleOnly = hasPassword === false;
  return (
    <Screen topInset="navigation-bar">
      <Card>
        <View style={styles.form}>
          <Text variant="title">{t("account_google_link_title")}</Text>
          {stage === "connecting" ? (
            <Text color="secondary">{t("account_google_link_connecting")}</Text>
          ) : null}
          {stage === "reauthenticate" && googleOnly ? (
            <>
              <Text color="secondary">{t("account_google_link_google_reauth_description")}</Text>
              <Button
                disabled={!proof.isConfigured || proof.isLoading}
                label={t("account_google_link_reauthenticate")}
                loading={proof.isLoading}
                onPress={startGoogle}
              />
            </>
          ) : null}
          {stage === "reauthenticate" && hasPassword ? (
            <>
              <Text color="secondary">{t("account_google_link_password_description")}</Text>
              <TextField
                autoComplete="current-password"
                label={t("account_password_current_label")}
                onChangeText={setPassword}
                secureTextEntry
                value={password}
              />
              <Button
                disabled={password.length < 8}
                label={t("account_google_link_reauthenticate")}
                onPress={() => void confirmPassword()}
              />
            </>
          ) : null}
          {stage === "prepare-google" ? (
            <>
              <Text color="secondary">{t("account_google_link_ready_description")}</Text>
              <Button
                disabled={!proof.isConfigured || proof.isLoading}
                label={t("account_google_link_continue")}
                loading={proof.isLoading}
                onPress={startGoogle}
              />
            </>
          ) : null}
        </View>
      </Card>
    </Screen>
  );
}
const styles = StyleSheet.create({ form: { gap: space[4] } });
