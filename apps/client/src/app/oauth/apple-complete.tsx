import { router } from "expo-router";
import { getTranslator } from "@/shared/i18n/locale";
import { Button, Card, Screen, Text } from "@/shared/ui";

/** The root layout completes the web popup; direct/cold visits never open a session. */
export default function AppleAuthenticationReturn() {
  const t = getTranslator();
  return (
    <Screen>
      <Card>
        <Text>{t("account_social_return_description")}</Text>
        <Button label={t("account_title")} onPress={() => router.replace("/account")} />
      </Card>
    </Screen>
  );
}
