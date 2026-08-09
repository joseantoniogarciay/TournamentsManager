import { BlurView } from "expo-blur";
import {
  createContext,
  type PropsWithChildren,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { BackHandler, Modal, Platform, Pressable, StyleSheet, View } from "react-native";

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

type ConfirmationDialogContextValue = {
  confirm: (dialog: Omit<Props, "visible">) => void;
};

const ConfirmationDialogContext = createContext<ConfirmationDialogContextValue | null>(null);

export function ConfirmationDialogProvider({ children }: PropsWithChildren) {
  const [dialog, setDialog] = useState<Omit<Props, "visible"> | null>(null);
  const confirm = useCallback((nextDialog: Omit<Props, "visible">) => setDialog(nextDialog), []);
  const dismiss = useCallback(() => setDialog(null), []);
  const confirmContextValue = useMemo(() => ({ confirm }), [confirm]);

  const cancel = useCallback(() => {
    const onCancel = dialog?.onCancel;
    dismiss();
    onCancel?.();
  }, [dialog, dismiss]);

  const accept = useCallback(() => {
    const onAccept = dialog?.onAccept;
    dismiss();
    onAccept?.();
  }, [dialog, dismiss]);

  return (
    <ConfirmationDialogContext.Provider value={confirmContextValue}>
      <View style={styles.host}>
        <View
          accessibilityElementsHidden={dialog !== null}
          importantForAccessibility={dialog ? "no-hide-descendants" : "auto"}
          style={styles.content}
        >
          {children}
        </View>
        {dialog ? (
          <ConfirmationDialog {...dialog} onAccept={accept} onCancel={cancel} visible />
        ) : null}
      </View>
    </ConfirmationDialogContext.Provider>
  );
}

export function useConfirmationDialog() {
  const value = useContext(ConfirmationDialogContext);
  if (!value)
    throw new Error("useConfirmationDialog debe usarse dentro de ConfirmationDialogProvider");
  return value;
}

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

  useEffect(() => {
    if (!visible || Platform.OS === "web") return;

    const subscription = BackHandler.addEventListener("hardwareBackPress", () => {
      onCancel();
      return true;
    });
    return () => subscription.remove();
  }, [onCancel, visible]);

  if (!visible) return null;

  return (
    <Modal animationType="fade" onRequestClose={onCancel} transparent visible={visible}>
      <View accessibilityViewIsModal style={styles.backdrop}>
        <BlurView
          blurMethod="dimezisBlurViewSdk31Plus"
          intensity={45}
          pointerEvents="none"
          style={styles.scrim}
          tint="dark"
        />
        {Platform.OS === "android" ? (
          <View
            pointerEvents="none"
            style={[styles.androidDimmingLayer, { backgroundColor: colors.surface.canvas }]}
          />
        ) : null}
        <Pressable
          accessibilityLabel={cancelLabel}
          accessibilityRole="button"
          onPress={onCancel}
          style={styles.dismissArea}
        />
        <View
          style={[
            styles.dialog,
            { backgroundColor: colors.surface.default, borderColor: colors.border.default },
          ]}
        >
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
  content: { flex: 1 },
  host: { flex: 1 },
  backdrop: {
    ...StyleSheet.absoluteFill,
    alignItems: "center",
    elevation: 1,
    justifyContent: "center",
    padding: space[5],
    zIndex: 1,
  },
  copy: { gap: space[2] },
  dialog: {
    borderWidth: 1,
    borderRadius: radius.card,
    gap: space[5],
    maxWidth: 440,
    padding: space[5],
    width: "100%",
  },
  dismissArea: StyleSheet.absoluteFill,
  scrim: {
    ...StyleSheet.absoluteFill,
  },
  androidDimmingLayer: {
    ...StyleSheet.absoluteFill,
    opacity: 0.16,
  },
});
