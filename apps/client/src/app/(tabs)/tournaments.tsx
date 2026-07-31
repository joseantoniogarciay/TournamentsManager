import { StyleSheet, View } from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";

import { space } from "@tournaments-manager/design-tokens";

import { getTranslator } from "@/shared/i18n/locale";
import { useSession } from "@/shared/session/session-provider";
import { Card, Screen, Text } from "@/shared/ui";

export default function TournamentsScreen() {
  const t = getTranslator();
  const { revision } = useSession();
  const insets = useSafeAreaInsets();

  return (
    <Screen bottomInset="none">
      <View key={revision} style={[styles.content, { paddingBottom: insets.bottom + space[12] }]}>
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
  content: { flex: 1, justifyContent: "center" },
  copy: { gap: space[2] },
});
