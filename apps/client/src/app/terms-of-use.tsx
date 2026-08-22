import { router, Stack } from "expo-router";
import { ScrollView, StyleSheet, View } from "react-native";

import { space } from "@tournaments-manager/design-tokens";

import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { Card, NavigationHeaderButton, Screen, Text, usesLiquidGlassNavigation } from "@/shared/ui";

export default function TermsOfUseScreen() {
  const t = getTranslator();
  const { colors } = usePreferences();
  const close = () => (router.canDismiss() ? router.dismiss() : router.replace("/"));
  const sections = [
    ["terms_of_use_service_title", "terms_of_use_service_body"],
    ["terms_of_use_account_title", "terms_of_use_account_body"],
    ["terms_of_use_content_title", "terms_of_use_content_body"],
    ["terms_of_use_changes_title", "terms_of_use_changes_body"],
    ["terms_of_use_law_title", "terms_of_use_law_body"],
  ] as const;

  return (
    <>
      <Stack.Screen
        options={{
          headerBackVisible: false,
          ...(!usesLiquidGlassNavigation
            ? {
                headerLeft: () => (
                  <NavigationHeaderButton
                    accessibilityLabel={t("common_close")}
                    icon="close"
                    nativeIcon={{ android: "close", ios: "xmark", web: "close" }}
                    onPress={close}
                  />
                ),
              }
            : {}),
          headerShadowVisible: false,
          headerStyle: { backgroundColor: colors.surface.canvas },
          headerTintColor: colors.text.primary,
          headerTitle: t("terms_of_use_title"),
          headerTitleAlign: "center",
        }}
      >
        {usesLiquidGlassNavigation ? (
          <Stack.Toolbar placement="left">
            <Stack.Toolbar.Button accessibilityLabel={t("common_close")} icon="xmark" onPress={close} />
          </Stack.Toolbar>
        ) : null}
      </Stack.Screen>
      <Screen bottomInset="none" topInset="navigation-bar">
        <ScrollView contentContainerStyle={styles.content} showsVerticalScrollIndicator={false}>
          <Card>
            <View style={styles.section}>
              <Text variant="display">{t("terms_of_use_title")}</Text>
              <Text color="secondary">{t("terms_of_use_updated")}</Text>
            </View>
          </Card>
          {sections.map(([title, body]) => (
            <Card key={title}>
              <View style={styles.section}>
                <Text variant="title">{t(title)}</Text>
                <Text color="secondary">{t(body)}</Text>
              </View>
            </Card>
          ))}
        </ScrollView>
      </Screen>
    </>
  );
}

const styles = StyleSheet.create({
  content: { gap: space[5], paddingBottom: space[8] },
  section: { gap: space[2] },
});
