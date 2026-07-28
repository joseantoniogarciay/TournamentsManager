import { router } from "expo-router";
import { StyleSheet, View } from "react-native";

import { Button, Screen, Text } from "@/shared/ui";

export default function LinkConfirmationScreen() {
  const close = () => {
    if (router.canGoBack()) {
      router.back();
      return;
    }
    router.replace("/");
  };

  return (
    <Screen>
      <View style={styles.content}>
        <Text variant="title">Confirmación de enlace</Text>
        <Text color="secondary">
          Esta ruta establecerá la sesión cuando el flujo de identidad esté implementado.
        </Text>
        <Button label="Cerrar" variant="secondary" onPress={close} />
      </View>
    </Screen>
  );
}

const styles = StyleSheet.create({
  content: { flex: 1, justifyContent: "center", gap: 16 },
});
