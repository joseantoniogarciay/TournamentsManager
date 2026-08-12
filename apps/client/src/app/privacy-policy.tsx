import { router, Stack } from "expo-router";
import { SymbolView } from "expo-symbols";
import { Platform, Pressable, ScrollView, StyleSheet, View } from "react-native";

import { control, radius, space } from "@tournaments-manager/design-tokens";

import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { Card, Screen, Text } from "@/shared/ui";
import { WebIcon } from "@/shared/ui/web-icon";

export default function PrivacyPolicyScreen() {
  const t = getTranslator();
  const { colors } = usePreferences();
  const close = () => {
    if (router.canDismiss()) {
      router.dismiss();
      return;
    }
    router.replace("/");
  };

  const sections = [
    ["privacy_policy_controller_title", "privacy_policy_controller_body"],
    ["privacy_policy_data_title", "privacy_policy_data_body"],
    ["privacy_policy_purposes_title", "privacy_policy_purposes_body"],
    ["privacy_policy_sharing_title", "privacy_policy_sharing_body"],
    ["privacy_policy_retention_title", "privacy_policy_retention_body"],
    ["privacy_policy_public_title", "privacy_policy_public_body"],
    ["privacy_policy_minors_title", "privacy_policy_minors_body"],
    ["privacy_policy_rights_title", "privacy_policy_rights_body"],
    ["privacy_policy_changes_title", "privacy_policy_changes_body"],
  ] as const;

  return (
    <>
      <Stack.Screen
        options={{
          headerBackVisible: false,
          headerLeft: () => (
            <Pressable
              accessibilityLabel={t("common_close")}
              accessibilityRole="button"
              onPress={close}
              style={[
                styles.navigationButton,
                { backgroundColor: colors.surface.default, borderColor: colors.border.default },
              ]}
            >
              {Platform.OS === "web" ? (
                <WebIcon color={colors.text.primary} name="close" size={control.iconSize} />
              ) : (
                <SymbolView
                  name={{ android: "close", ios: "xmark", web: "close" }}
                  size={control.iconSize}
                  tintColor={colors.text.primary}
                />
              )}
            </Pressable>
          ),
          headerShadowVisible: false,
          headerStyle: { backgroundColor: colors.surface.canvas },
          headerTintColor: colors.text.primary,
          headerTitle: t("privacy_policy_title"),
          headerTitleAlign: "center",
        }}
      />
      <Screen bottomInset="none" topInset="navigation-bar">
        <ScrollView contentContainerStyle={styles.content} showsVerticalScrollIndicator={false}>
          <Card>
            <View style={styles.intro}>
              <Text variant="display">{t("privacy_policy_title")}</Text>
              <Text color="secondary">{t("privacy_policy_updated")}</Text>
              <Text>{t("privacy_policy_intro")}</Text>
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
  intro: { gap: space[3] },
  navigationButton: {
    alignItems: "center",
    borderRadius: radius.pill,
    borderWidth: 1,
    height: control.minHeight,
    justifyContent: "center",
    width: control.minHeight,
  },
  section: { gap: space[2] },
});
