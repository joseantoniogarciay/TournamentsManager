import { router } from "expo-router";
import { useCallback, useEffect, useRef, useState } from "react";
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
import { Button, ModalDialog, Screen, Text, TextField } from "@/shared/ui";

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
  const [reauthenticating, setReauthenticating] = useState(false);
  const reauthentication = useRef({ generation: 0, busy: false });
  const handledProofError = useRef<unknown>(null);

  useEffect(() => {
    if (!visible) return;
    const operation = reauthentication.current;
    const generation = ++operation.generation;
    operation.busy = false;
    setReauthenticating(false);
    setHasPassword(null);
    setPassword("");
    setTicket(null);
    setStage("reauthenticate");
    void getAccountAccessMethods()
      .then((access) => {
        if (operation.generation === generation) setHasPassword(access.methods.password);
      })
      .catch(() => {
        if (operation.generation === generation) onDismiss();
      });
    return () => {
      operation.generation++;
      operation.busy = false;
    };
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
    if (!proof.error || proof.error === handledProofError.current) return;
    handledProofError.current = proof.error;
    if (!visible) return;
    const message =
      proof.error instanceof GoogleLinkError
        ? t(
            proof.error.reason === "conflict"
              ? "account_google_link_conflict"
              : proof.error.reason === "wrong-account"
                ? "account_google_reauthentication_wrong_account"
                : "account_google_link_expired",
          )
        : t(getRequestFailure(proof.error).messageKey);
    show({ kind: "generic-error", message });
    onDismiss();
  }, [onDismiss, proof.error, show, t, visible]);

  useEffect(() => {
    if (visible && stage === "prepare-google" && proof.isConfigured) proof.prepare();
  }, [proof.isConfigured, proof.prepare, stage, visible]);

  const confirmPassword = async () => {
    const operation = reauthentication.current;
    if (password.length < 8 || operation.busy) return;
    const generation = operation.generation;
    operation.busy = true;
    setReauthenticating(true);
    try {
      const nextTicket = await reauthenticateWithPassword(password, "link-google");
      if (operation.generation !== generation) return;
      setTicket(nextTicket);
      setStage("prepare-google");
    } catch (error) {
      if (operation.generation !== generation) return;
      show({
        kind: "generic-error",
        message: t(
          error instanceof GoogleLinkError
            ? "account_google_link_expired"
            : getRequestFailure(error).messageKey,
        ),
      });
      onDismiss();
    } finally {
      if (operation.generation === generation) {
        operation.busy = false;
        setReauthenticating(false);
      }
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
              disabled={password.length < 8 || reauthenticating}
              label={t("account_google_link_reauthenticate")}
              loading={reauthenticating}
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
  const t = getTranslator();
  const dismiss = useCallback(() => router.back(), []);
  return (
    <Screen
      navigationTitle={t("account_google_link_title")}
      bottomInset="none"
      topInset="navigation-bar"
    >
      <GoogleLinkDialog onDismiss={dismiss} onLinked={dismiss} visible />
    </Screen>
  );
}
const styles = StyleSheet.create({ form: { gap: space[4] } });
