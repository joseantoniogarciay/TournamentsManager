import { router, useLocalSearchParams } from "expo-router";
import { Platform, StyleSheet, View } from "react-native";
import { useEffect, useState } from "react";
import { space } from "@tournaments-manager/design-tokens";
import { confirmRecovery, inspectRecovery } from "@/features/password-recovery/api";
import {
  getPendingPasswordResetToken,
  setPendingPasswordResetToken,
} from "@/features/password-recovery/pending-link";
import { getTranslator } from "@/shared/i18n/locale";
import { useSession } from "@/shared/session/session-provider";
import { Button, Card, KeyboardAwareScrollView, Screen, Text, TextField } from "@/shared/ui";

export default function PasswordResetScreen() {
  const t = getTranslator();
  const params = useLocalSearchParams<{ token?: string }>();
  const incoming = typeof params.token === "string" ? params.token : "";
  const [token] = useState(() => incoming || getPendingPasswordResetToken() || "");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [visible, setVisible] = useState(false);
  const [loading, setLoading] = useState(true);
  const [failed, setFailed] = useState(false);
  const { beginSessionReplacement, cancelSessionReplacement, completeSessionReplacement } =
    useSession();
  useEffect(() => {
    if (incoming) {
      setPendingPasswordResetToken(incoming);
      router.replace("/link/password-reset" as never);
      return;
    }
    if (!token) {
      setFailed(true);
      setLoading(false);
      return;
    }
    void inspectRecovery(token)
      .then(setEmail)
      .catch(() => setFailed(true))
      .finally(() => setLoading(false));
  }, [incoming, token]);
  const submit = async () => {
    try {
      beginSessionReplacement();
      const session = await confirmRecovery(
        token,
        password,
        Platform.OS === "web" ? "cookie" : "bearer",
      );
      completeSessionReplacement(session.user);
    } catch {
      cancelSessionReplacement();
      setFailed(true);
    }
  };
  if (loading) return <Screen />;
  if (failed)
    return (
      <Screen>
        <Text>{t("password_recovery_link_invalid")}</Text>
      </Screen>
    );
  return (
    <Screen>
      <KeyboardAwareScrollView>
        <Card>
          <View style={styles.form}>
            <Text variant="title">{t("password_recovery_new_title")}</Text>
            <TextField
              editable={false}
              keyboardType="email-address"
              label={t("account_email_label")}
              value={email}
            />
            <TextField
              autoComplete="new-password"
              label={t("password_recovery_new_password")}
              onChangeText={setPassword}
              passwordVisibility={{
                isVisible: visible,
                label: t(visible ? "password_hide" : "password_show"),
                onPress: () => setVisible(!visible),
              }}
              secureTextEntry={!visible}
              value={password}
            />
            <Text color="secondary">{strength(t, password)}</Text>
            <Button
              disabled={password.length < 8}
              label={t("password_recovery_save")}
              onPress={() => void submit()}
            />
          </View>
        </Card>
      </KeyboardAwareScrollView>
    </Screen>
  );
}
const styles = StyleSheet.create({ form: { gap: space[4] } });
function strength(t: ReturnType<typeof getTranslator>, value: string) {
  return value.length < 8
    ? t("password_strength_weak")
    : value.length < 15
      ? t("password_strength_ok")
      : t("password_strength_strong");
}
