import { StyleSheet, View } from "react-native";

import { space } from "@tournaments-manager/design-tokens";

import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getTranslator } from "@/shared/i18n/locale";
import { Button, Card, Text } from "@/shared/ui";

import { getTipPaymentLinks } from "../payment-links";

export function TipCard() {
  const t = getTranslator();
  const { show } = useFeedback();
  const paymentLinks = getTipPaymentLinks();

  if (!paymentLinks) return null;

  const openPaymentLink = (url: string) => {
    const checkout = window.open(url, "_blank", "noopener,noreferrer");
    if (!checkout) show({ kind: "generic-error", message: t("common_request_error") });
  };

  return (
    <Card>
      <View style={styles.content}>
        <Text variant="title">{t("tips_title")}</Text>
        <Text color="secondary">{t("tips_description")}</Text>
        <View style={styles.actions}>
          {paymentLinks.map(({ amount, url }) => (
            <Button
              key={amount}
              label={t(`tips_amount_${amount}_eur`)}
              onPress={() => openPaymentLink(url)}
              variant="secondary"
            />
          ))}
        </View>
        <Text color="secondary" style={styles.disclaimer}>
          {t("tips_disclaimer")}
        </Text>
      </View>
    </Card>
  );
}

const styles = StyleSheet.create({
  actions: { gap: space[3] },
  content: { gap: space[3] },
  disclaimer: { marginTop: space[1] },
});
