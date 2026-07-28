import { router } from "expo-router";
import { StyleSheet, View } from "react-native";

import { getTranslator } from "@/shared/i18n/locale";
import { Button, Screen, Text } from "@/shared/ui";

export default function LinkConfirmationScreen() {
  const t = getTranslator();

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
        <Text variant="title">{t("link_confirmation_title")}</Text>
        <Text color="secondary">{t("link_confirmation_description")}</Text>
        <Button label={t("common_close")} variant="secondary" onPress={close} />
      </View>
    </Screen>
  );
}

const styles = StyleSheet.create({
  content: { flex: 1, justifyContent: "center", gap: 16 },
});
