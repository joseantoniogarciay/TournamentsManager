import { ScrollView, StyleSheet, View } from "react-native";

import { space } from "@tournaments-manager/design-tokens";

import { getTranslator } from "@/shared/i18n/locale";
import { useSession } from "@/shared/session/session-provider";
import { Card, Screen, Text, useTabContentBottomPadding } from "@/shared/ui";

export default function TournamentsScreen() {
  const t = getTranslator();
  const { revision } = useSession();
  const tabContentBottomPadding = useTabContentBottomPadding();

  return (
    <Screen bottomInset="none">
      <ScrollView
        contentContainerStyle={[styles.content, { paddingBottom: tabContentBottomPadding }]}
        key={revision}
        showsVerticalScrollIndicator={false}
      >
        <Card>
          <View style={styles.copy}>
            <Text variant="title">{t("tournaments_title")}</Text>
            <Text color="secondary">{t("tournaments_description")}</Text>
          </View>
        </Card>
      </ScrollView>
    </Screen>
  );
}

const styles = StyleSheet.create({
  content: { flexGrow: 1, justifyContent: "center" },
  copy: { gap: space[2] },
});
