import { StyleSheet, View } from "react-native";
import { StatusBar } from "expo-status-bar";

import { Button, Screen, Text } from "@/shared/ui";

export default function HomeScreen() {
  return (
    <Screen>
      <StatusBar style="dark" />
      <View style={styles.content}>
        <Text variant="display">TournamentsManager</Text>
        <Text color="secondary">La home se definirá sobre las primitivas Pulse.</Text>
        <Button label="Crear liga" onPress={() => undefined} />
      </View>
    </Screen>
  );
}

const styles = StyleSheet.create({
  content: { flex: 1, justifyContent: "center", gap: 16 },
});
