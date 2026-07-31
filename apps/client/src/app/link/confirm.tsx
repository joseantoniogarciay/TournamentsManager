import { router, useLocalSearchParams } from "expo-router";
import { useEffect, useState } from "react";
import { Platform, StyleSheet, View } from "react-native";

import { confirmRegistration } from "@/features/registration/api";
import { usePendingVerification } from "@/features/registration/pending-verification";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import { getTranslator } from "@/shared/i18n/locale";
import { useSession } from "@/shared/session/session-provider";
import { Button, Screen, Text } from "@/shared/ui";

export default function LinkConfirmationScreen() {
  const t = getTranslator();
  const { show } = useFeedback();
  const { token, setToken } = usePendingVerification();
  const { replaceSession } = useSession();
  const params = useLocalSearchParams<{ sent?: string | string[]; token?: string | string[] }>();
  const [isConfirming, setIsConfirming] = useState(false);

  useEffect(() => {
    const incomingToken = typeof params.token === "string" ? params.token : null;
    if (!incomingToken || token) return;
    setToken(incomingToken);
    router.replace("/link/confirm");
  }, [params.token, setToken, token]);

  useEffect(() => {
    if (!token || isConfirming) return;
    setIsConfirming(true);
    void confirmRegistration(token, Platform.OS === "web" ? "cookie" : "bearer")
      .then((session) => {
        setToken(null);
        replaceSession(session.user);
        router.dismissAll();
        router.replace("/");
      })
      .catch((error: unknown) => {
        const failure = getRequestFailure(error);
        show({ kind: failure.kind, message: t(failure.messageKey) });
      })
      .finally(() => setIsConfirming(false));
  }, [isConfirming, replaceSession, setToken, show, t, token]);

  const close = () => {
    if (router.canGoBack()) {
      router.back();
      return;
    }
    router.replace("/");
  };

  return (
    <Screen>
      <View style={styles.content}>
        <Text color="secondary">
          {token
            ? t("link_confirmation_description")
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
