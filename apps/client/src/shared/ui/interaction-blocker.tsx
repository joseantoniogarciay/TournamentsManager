import { StyleSheet, View } from "react-native";

type InteractionBlockerProps = { accessibilityLabel: string };

/** Capa transparente para estados modales que deben impedir interacción en la ruta. */
export function InteractionBlocker({ accessibilityLabel }: InteractionBlockerProps) {
  return (
    <View
      accessible
      accessibilityLabel={accessibilityLabel}
      accessibilityRole="progressbar"
      accessibilityState={{ busy: true }}
      accessibilityViewIsModal
      pointerEvents="auto"
      style={styles.blocker}
    />
  );
}

const styles = StyleSheet.create({
  blocker: {
    bottom: 0,
    left: 0,
    position: "absolute",
    right: 0,
    top: 0,
    zIndex: 1,
  },
});
