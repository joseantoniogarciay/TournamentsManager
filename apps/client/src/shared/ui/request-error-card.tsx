import { StyleSheet, View } from "react-native";

import { space } from "@tournaments-manager/design-tokens";

import { Button } from "./button";
import { Card } from "./card";
import { Text } from "./text";

type RequestErrorCardProps = {
  actionLabel: string;
  loading?: boolean;
  message: string;
  onRetry: () => void;
};

/** Estado terminal reutilizable para una carga que no puede mostrar contenido. */
export function RequestErrorCard({
  actionLabel,
  loading = false,
  message,
  onRetry,
}: RequestErrorCardProps) {
  return (
    <Card>
      <View style={styles.content}>
        <Text color="secondary">{message}</Text>
        <Button label={actionLabel} loading={loading} onPress={onRetry} variant="secondary" />
      </View>
    </Card>
  );
}

const styles = StyleSheet.create({
  content: { gap: space[4] },
});
