import { ScrollView, StyleSheet, View } from "react-native";
import { StatusBar } from "expo-status-bar";

import { color, radius, space } from "@tournaments-manager/design-tokens";

import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { Button, Card, Screen, Text } from "@/shared/ui";

export default function HomeScreen() {
  const { show } = useFeedback();
  const { resolvedTheme } = usePreferences();
  const t = getTranslator();

  return (
    <Screen>
      <StatusBar style={resolvedTheme === "dark" ? "light" : "dark"} />
      <ScrollView
        contentContainerStyle={styles.content}
        showsVerticalScrollIndicator={false}
        style={styles.scroll}
      >
        <Card>
          <View style={styles.hero}>
            <Text variant="display">{t("home_hero")}</Text>
            <Text color="secondary" variant="bodyLarge">
              {t("home_introduction")}
            </Text>
            <Button
              label={t("home_create_tournament")}
              onPress={() =>
                show({ kind: "generic-error", message: t("home_create_tournament_unavailable") })
              }
            />
          </View>
        </Card>

        <Card>
          <View style={styles.section}>
            <Text variant="title">{t("home_section_title")}</Text>
            <Text color="secondary">{t("home_section_description")}</Text>
          </View>
        </Card>

        <Card>
          <View style={styles.steps}>
            <Step
              number="1"
              title={t("home_step_1_title")}
              description={t("home_step_1_description")}
            />
            <Step
              number="2"
              title={t("home_step_2_title")}
              description={t("home_step_2_description")}
            />
            <Step
              number="3"
              title={t("home_step_3_title")}
              description={t("home_step_3_description")}
            />
          </View>
        </Card>

        <Card>
          <Text variant="caption" color="secondary">
            {t("home_footer")}
          </Text>
        </Card>
      </ScrollView>
    </Screen>
  );
}

function Step({
  number,
  title,
  description,
}: {
  number: string;
  title: string;
  description: string;
}) {
  return (
    <View style={styles.step}>
      <View style={styles.stepNumber}>
        <Text color="inverse">{number}</Text>
      </View>
      <View style={styles.stepContent}>
        <Text variant="bodyLarge">{title}</Text>
        <Text color="secondary">{description}</Text>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  scroll: { flex: 1 },
  content: { gap: space[5], paddingBottom: space[12] },
  hero: { gap: space[4] },
  section: { gap: space[2] },
  steps: { gap: space[5] },
  step: { flexDirection: "row", gap: space[3] },
  stepNumber: {
    alignItems: "center",
    backgroundColor: color.brand.primary,
    borderRadius: radius.pill,
    height: 28,
    justifyContent: "center",
    width: 28,
  },
  stepContent: { flex: 1, gap: space[1] },
});
