import { router, useLocalSearchParams } from "expo-router";
import { useEffect, useState } from "react";
import { Platform, StyleSheet, View } from "react-native";

import {
  confirmRegistration,
  RegistrationVerificationError,
  type RegistrationVerificationFailure,
} from "@/features/registration/api";
import { usePendingVerification } from "@/features/registration/pending-verification";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import { getTranslator } from "@/shared/i18n/locale";
import { useSession } from "@/shared/session/session-provider";
import { Button, Screen, Text } from "@/shared/ui";

const minimumConfirmationTransitionDuration = 2_000;

export default function LinkConfirmationScreen() {
  const t = getTranslator();
  const { show } = useFeedback();
  const { token, setToken } = usePendingVerification();
  const { beginSessionReplacement, cancelSessionReplacement, completeSessionReplacement } =
    useSession();
  const params = useLocalSearchParams<{ sent?: string | string[]; token?: string | string[] }>();
  const [isConfirming, setIsConfirming] = useState(false);
  const [confirmationFailure, setConfirmationFailure] = useState<
    RegistrationVerificationFailure | "unexpected" | null
  >(null);

  useEffect(() => {
    const incomingToken = typeof params.token === "string" ? params.token : null;
    if (!incomingToken || token) return;
    setToken(incomingToken);
    router.replace("/link/confirm");
  }, [params.token, setToken, token]);

  useEffect(() => {
    if (!token || isConfirming || confirmationFailure) return;
    setIsConfirming(true);
    beginSessionReplacement();
    const confirmationStartedAt = Date.now();
    void confirmRegistration(token, Platform.OS === "web" ? "cookie" : "bearer")
      .then(async (session) => {
        const remainingDuration = Math.max(
          0,
          minimumConfirmationTransitionDuration - (Date.now() - confirmationStartedAt),
        );
        if (remainingDuration > 0) {
          await new Promise<void>((resolve) => setTimeout(resolve, remainingDuration));
        }
        setToken(null);
        show({ kind: "success", message: t("link_confirmation_success") });
        completeSessionReplacement(session.user);
      })
      .catch((error: unknown) => {
        cancelSessionReplacement();
        if (error instanceof RegistrationVerificationError) {
          setConfirmationFailure(error.failure);
          show({
            kind: "generic-error",
            message: t(getVerificationFailureMessageKey(error.failure)),
          });
          return;
        }
        setConfirmationFailure("unexpected");
        const failure = getRequestFailure(error);
        show({ kind: failure.kind, message: t(failure.messageKey) });
      })
      .finally(() => setIsConfirming(false));
  }, [
    beginSessionReplacement,
    cancelSessionReplacement,
    completeSessionReplacement,
    confirmationFailure,
    isConfirming,
    setToken,
    show,
    t,
    token,
  ]);

  const close = () => {
    if (router.canGoBack()) {
      router.back();
      return;
    }
    router.replace("/");
  };

  if ((token || isConfirming) && !confirmationFailure) return <Screen />;

  return (
    <Screen>
      <View style={styles.content}>
        <Text color="secondary">
          {confirmationFailure && confirmationFailure !== "unexpected"
            ? t(getVerificationFailureMessageKey(confirmationFailure))
            : params.sent === "1"
              ? t("account_registration_email_sent")
              : t("link_confirmation_missing")}
        </Text>
        <Button label={t("common_close")} variant="secondary" onPress={close} />
      </View>
    </Screen>
  );
}

const styles = StyleSheet.create({
  content: { flex: 1, justifyContent: "center", gap: 16 },
});

function getVerificationFailureMessageKey(failure: RegistrationVerificationFailure) {
  return failure === "already-used" ? "link_confirmation_used" : "link_confirmation_expired";
}
