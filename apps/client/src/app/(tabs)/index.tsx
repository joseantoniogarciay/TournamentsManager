import { useEffect } from "react";
import { ScrollView, StyleSheet, View } from "react-native";
import { StatusBar } from "expo-status-bar";
import { useSafeAreaInsets } from "react-native-safe-area-context";

import { radius, space } from "@tournaments-manager/design-tokens";

import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { useSession } from "@/shared/session/session-provider";
import { consumeDeferredInitialDeepLink } from "@/shared/navigation/deep-link-gate";
import { Button, Card, Screen, Text } from "@/shared/ui";
import { router, type Href } from "expo-router";

const safeAreaProbeHeight = 500;

export default function HomeScreen() {
  const { show } = useFeedback();
  const { resolvedTheme } = usePreferences();
  const { revision } = useSession();
  const insets = useSafeAreaInsets();
  const t = getTranslator();

  useEffect(() => {
    const frame = requestAnimationFrame(() => {
      const deferredDeepLink = consumeDeferredInitialDeepLink();
      if (deferredDeepLink) router.replace(deferredDeepLink as Href);
    });
    return () => cancelAnimationFrame(frame);
  }, []);

  return (
    <Screen bottomInset="none">
      <StatusBar style={resolvedTheme === "dark" ? "light" : "dark"} />
      <ScrollView
        key={revision}
        contentContainerStyle={[styles.content, { paddingBottom: insets.bottom + space[12] }]}
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

        <Card style={{ height: safeAreaProbeHeight }} />
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
  const { colors } = usePreferences();

  return (
    <View style={styles.step}>
      <View style={[styles.stepNumber, { borderColor: colors.text.primary }]}>
        <Text>{number}</Text>
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
  content: { gap: space[5] },
  hero: { gap: space[4] },
  section: { gap: space[2] },
  steps: { gap: space[5] },
  step: { flexDirection: "row", gap: space[3] },
  stepNumber: {
    alignItems: "center",
    borderWidth: 1,
    borderRadius: radius.pill,
    height: 28,
    justifyContent: "center",
    width: 28,
  },
  stepContent: { flex: 1, gap: space[1] },
});
