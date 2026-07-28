import { StyleSheet, View } from "react-native";

import { space } from "@tournaments-manager/design-tokens";

import { getTranslator } from "@/shared/i18n/locale";
import { Card, Screen, Text } from "@/shared/ui";

export default function TournamentsScreen() {
  const t = getTranslator();

  return (
    <Screen>
      <View style={styles.content}>
        <Card>
          <View style={styles.copy}>
            <Text variant="title">{t("tournaments_title")}</Text>
            <Text color="secondary">{t("tournaments_description")}</Text>
          </View>
        </Card>
      </View>
    </Screen>
  );
}

const styles = StyleSheet.create({
  content: { flex: 1, justifyContent: "center", paddingBottom: space[12] },
  copy: { gap: space[2] },
});
