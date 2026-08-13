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
import { Button, ModalDialog, Text, TextField } from "@/shared/ui";

type Stage = "reauthenticate" | "prepare-google" | "connecting";

export function GoogleLinkDialog({
  onDismiss,
  onLinked,
  visible,
}: {
  onDismiss: () => void;
  onLinked: () => void;
  visible: boolean;
}) {
  const t = getTranslator();
  const { show } = useFeedback();
  const [hasPassword, setHasPassword] = useState<boolean | null>(null);
  const [password, setPassword] = useState("");
  const [ticket, setTicket] = useState<string | null>(null);
  const [stage, setStage] = useState<Stage>("reauthenticate");

  useEffect(() => {
    if (!visible) return;
    setPassword("");
    setTicket(null);
    setStage("reauthenticate");
    void getAccountAccessMethods()
      .then((access) => setHasPassword(access.methods.password))
      .catch(() => {
        onDismiss();
      });
  }, [onDismiss, visible]);

  const onProof = useCallback(
    async (challenge: { id: string }, idToken: string) => {
      if (!ticket) {
        const nextTicket = await reauthenticateWithGoogle(challenge.id, idToken, "link-google");
        setTicket(nextTicket);
        setStage("prepare-google");
        return;
      }
      setStage("connecting");
      await linkGoogle(ticket, challenge.id, idToken);
      onLinked();
    },
    [onLinked, ticket],
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
    onDismiss();
  }, [onDismiss, proof.error, show, t]);

  useEffect(() => {
    if (visible && stage === "prepare-google") proof.prepare();
  }, [proof.prepare, stage, visible]);

  const confirmPassword = async () => {
    if (password.length < 8) return;
    try {
      setTicket(await reauthenticateWithPassword(password, "link-google"));
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
      onDismiss();
    }
  };

  const startGoogle = () => void proof.start();
  const googleOnly = hasPassword === false;
  return (
    <ModalDialog
      dismissAccessibilityLabel={t("common_cancel")}
      onDismiss={onDismiss}
      visible={visible}
    >
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
    </ModalDialog>
  );
}

export default function GoogleLinkScreen() {
  return (
    <GoogleLinkDialog onDismiss={() => router.back()} onLinked={() => router.back()} visible />
  );
}
const styles = StyleSheet.create({ form: { gap: space[4] } });
