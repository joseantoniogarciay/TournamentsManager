import { Modal, StyleSheet, View } from "react-native";

import { radius, space } from "@tournaments-manager/design-tokens";

import { usePreferences } from "@/shared/preferences/preferences-provider";

import { Button } from "./button";
import { Text } from "./text";

type Props = {
  visible: boolean;
  title: string;
  description: string;
  acceptLabel: string;
  cancelLabel: string;
  onAccept: () => void;
  onCancel: () => void;
};

export function ConfirmationDialog({
  visible,
  title,
  description,
  acceptLabel,
  cancelLabel,
  onAccept,
  onCancel,
}: Props) {
  const { colors } = usePreferences();
  return (
    <Modal animationType="fade" onRequestClose={onCancel} transparent visible={visible}>
      <View style={[styles.backdrop, { backgroundColor: colors.surface.subtle }]}>
        <View style={[styles.dialog, { backgroundColor: colors.surface.default }]}>
          <View style={styles.copy}>
            <Text variant="title">{title}</Text>
            <Text color="secondary">{description}</Text>
          </View>
          <View style={styles.actions}>
            <Button label={acceptLabel} onPress={onAccept} />
            <Button label={cancelLabel} onPress={onCancel} variant="secondary" />
          </View>
        </View>
      </View>
    </Modal>
  );
}

const styles = StyleSheet.create({
  actions: { gap: space[3] },
  backdrop: { alignItems: "center", flex: 1, justifyContent: "center", padding: space[5] },
  copy: { gap: space[2] },
  dialog: {
    borderRadius: radius.card,
    gap: space[5],
    maxWidth: 440,
    padding: space[5],
    width: "100%",
  },
});
