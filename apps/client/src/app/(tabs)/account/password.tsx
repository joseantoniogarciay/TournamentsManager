import { router } from "expo-router";
import { useState } from "react";
import { StyleSheet, View } from "react-native";

import { space } from "@tournaments-manager/design-tokens";

import { setAccountPassword } from "@/features/account-access/api";
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
  const [submitting, setSubmitting] = useState(false);
  const tabContentBottomPadding = useTabContentBottomPadding();
  const valid = currentPassword.length >= 8 && password.length >= 8;

  const submit = async () => {
    if (!valid || submitting) return;
    setSubmitting(true);
    try {
      await setAccountPassword(currentPassword, password);
      router.back();
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
            <Text color="secondary">{t("account_password_change_description")}</Text>
            <TextField
              autoComplete="current-password"
              label={t("account_password_current_label")}
              onChangeText={setCurrentPassword}
              secureTextEntry
              value={currentPassword}
            />
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
