import { router, useLocalSearchParams } from "expo-router";
import { useEffect, useRef, useState } from "react";
import { Platform, StyleSheet, View } from "react-native";

import { space } from "@tournaments-manager/design-tokens";

import {
  confirmRegistration,
  RegistrationVerificationError,
  type RegistrationVerificationFailure,
} from "@/features/registration/api";
import { usePendingVerification } from "@/features/registration/pending-verification";
import { getTranslator } from "@/shared/i18n/locale";
import { useSession } from "@/shared/session/session-provider";
import { Button, Card, Screen, Text } from "@/shared/ui";

const minimumConfirmationTransitionDuration = 2_000;

type ConfirmationAttempt = {
  controller: AbortController;
  token: string;
};

export default function LinkConfirmationScreen() {
  const t = getTranslator();
  const { token, setToken } = usePendingVerification();
  const { beginSessionReplacement, cancelSessionReplacement, completeSessionReplacement } =
    useSession();
  const params = useLocalSearchParams<{ sent?: string | string[]; token?: string | string[] }>();
  const [isConfirming, setIsConfirming] = useState(false);
  const [confirmationFailure, setConfirmationFailure] = useState<
    RegistrationVerificationFailure | "unexpected" | null
  >(null);
  const activeAttempt = useRef<ConfirmationAttempt | null>(null);

  useEffect(() => {
    const incomingToken = typeof params.token === "string" ? params.token : null;
    if (!incomingToken || incomingToken === token) return;
    activeAttempt.current?.controller.abort();
    setToken(incomingToken);
    setConfirmationFailure(null);
    router.replace("/link/confirm");
  }, [params.token, setToken, token]);

  useEffect(() => {
    const hasTokenInURL = typeof params.token === "string";
    if (hasTokenInURL || !token || confirmationFailure || activeAttempt.current?.token === token) {
      return;
    }

    activeAttempt.current?.controller.abort();
    const controller = new AbortController();
    const attempt = { controller, token };
    activeAttempt.current = attempt;
    setIsConfirming(true);
    beginSessionReplacement();
    const confirmationStartedAt = Date.now();
    void confirmRegistration(token, Platform.OS === "web" ? "cookie" : "bearer", controller.signal)
      .then(async (session) => {
        if (controller.signal.aborted || activeAttempt.current !== attempt) return;
        const remainingDuration = Math.max(
          0,
          minimumConfirmationTransitionDuration - (Date.now() - confirmationStartedAt),
        );
        if (remainingDuration > 0) {
          await new Promise<void>((resolve) => setTimeout(resolve, remainingDuration));
        }
        if (controller.signal.aborted || activeAttempt.current !== attempt) return;
        setToken(null);
        completeSessionReplacement(session.user);
      })
      .catch((error: unknown) => {
        if (controller.signal.aborted || activeAttempt.current !== attempt) return;
        cancelSessionReplacement();
        if (error instanceof RegistrationVerificationError) {
          setConfirmationFailure(error.failure);
          return;
        }
        setConfirmationFailure("unexpected");
      })
      .finally(() => {
        if (activeAttempt.current !== attempt) return;
        activeAttempt.current = null;
        setIsConfirming(false);
      });
  }, [
    beginSessionReplacement,
    cancelSessionReplacement,
    completeSessionReplacement,
    confirmationFailure,
    params.token,
    setToken,
    token,
  ]);

  const returnHome = () => router.replace("/");

  if ((token || params.token || isConfirming) && !confirmationFailure) {
    return <VerificationState message={t("link_confirmation_loading")} />;
  }

  return (
    <Screen>
      <View style={styles.content}>
        <Card>
          <View style={styles.cardContent}>
            <Text color="secondary">
              {confirmationFailure === "unexpected"
                ? t("common_request_error")
                : confirmationFailure
                  ? t(getVerificationFailureMessageKey(confirmationFailure))
                  : params.sent === "1"
                    ? t("account_registration_email_sent")
                    : t("link_confirmation_missing")}
            </Text>
            <Button
              label={t("link_confirmation_action")}
              variant="secondary"
              onPress={returnHome}
            />
          </View>
        </Card>
      </View>
    </Screen>
  );
}

function VerificationState({ message }: { message: string }) {
  return (
    <Screen>
      <View style={styles.content}>
        <Card>
          <Text color="secondary">{message}</Text>
        </Card>
      </View>
    </Screen>
  );
}

const styles = StyleSheet.create({
  content: { flex: 1, justifyContent: "center" },
  cardContent: { gap: space[4] },
});

function getVerificationFailureMessageKey(failure: RegistrationVerificationFailure) {
  return failure === "already-used" ? "link_confirmation_used" : "link_confirmation_expired";
}
