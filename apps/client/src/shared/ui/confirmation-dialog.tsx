import { BlurView } from "expo-blur";
import {
  createContext,
  type PropsWithChildren,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { useFocusEffect } from "expo-router";
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
  acceptVariant?: "primary" | "destructive";
  cancelLabel: string;
  onAccept: () => void;
  onCancel: () => void;
};

type ModalDialogProps = {
  visible: boolean;
  dismissAccessibilityLabel: string;
  onDismiss: () => void;
  children: ReactNode;
};

type ConfirmationDialogContextValue = {
  confirm: (dialog: Omit<Props, "visible">) => void;
  dialog: Omit<Props, "visible"> | null;
  accept: () => void;
  cancel: () => void;
};

const ConfirmationDialogContext = createContext<ConfirmationDialogContextValue | null>(null);

export function ConfirmationDialogProvider({ children }: PropsWithChildren) {
  const [dialog, setDialog] = useState<Omit<Props, "visible"> | null>(null);
  const confirm = useCallback((nextDialog: Omit<Props, "visible">) => setDialog(nextDialog), []);
  const dismiss = useCallback(() => setDialog(null), []);
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
  const confirmContextValue = useMemo(
    () => ({ accept, cancel, confirm, dialog }),
    [accept, cancel, confirm, dialog],
  );

  return (
    <ConfirmationDialogContext.Provider value={confirmContextValue}>
      {children}
    </ConfirmationDialogContext.Provider>
  );
}

export function useConfirmationDialog() {
  const value = useContext(ConfirmationDialogContext);
  if (!value)
    throw new Error("useConfirmationDialog debe usarse dentro de ConfirmationDialogProvider");
  return { confirm: value.confirm };
}

export function ConfirmationDialogHost() {
  const value = useContext(ConfirmationDialogContext);
  const [isFocused, setIsFocused] = useState(false);
  useFocusEffect(
    useCallback(() => {
      setIsFocused(true);
      return () => setIsFocused(false);
    }, []),
  );
  if (!value)
    throw new Error("ConfirmationDialogHost debe usarse dentro de ConfirmationDialogProvider");
  if (!isFocused || !value.dialog) return null;
  return (
    <ConfirmationDialog {...value.dialog} onAccept={value.accept} onCancel={value.cancel} visible />
  );
}

export function ConfirmationDialog({
  visible,
  title,
  description,
  acceptLabel,
  acceptVariant = "primary",
  cancelLabel,
  onAccept,
  onCancel,
}: Props) {
  useEffect(() => {
    if (!visible || Platform.OS === "web") return;

    const subscription = BackHandler.addEventListener("hardwareBackPress", () => {
      onCancel();
      return true;
    });
    return () => subscription.remove();
  }, [onCancel, visible]);

  return (
    <ModalDialog dismissAccessibilityLabel={cancelLabel} onDismiss={onCancel} visible={visible}>
      <View style={styles.copy}>
        <Text variant="title">{title}</Text>
        <Text color="secondary">{description}</Text>
      </View>
      <View style={styles.actions}>
        <Button label={acceptLabel} onPress={onAccept} variant={acceptVariant} />
        <Button label={cancelLabel} onPress={onCancel} variant="secondary" />
      </View>
    </ModalDialog>
  );
}

export function ModalDialog({
  visible,
  dismissAccessibilityLabel,
  onDismiss,
  children,
}: ModalDialogProps) {
  const { colors } = usePreferences();
  const [webScrimVisible, setWebScrimVisible] = useState(false);

  useEffect(() => {
    if (Platform.OS !== "web") return;

    if (!visible) {
      setWebScrimVisible(false);
      return;
    }

    let secondFrame: ReturnType<typeof requestAnimationFrame> | undefined;
    const firstFrame = requestAnimationFrame(() => {
      secondFrame = requestAnimationFrame(() => setWebScrimVisible(true));
    });

    return () => {
      cancelAnimationFrame(firstFrame);
      if (secondFrame !== undefined) cancelAnimationFrame(secondFrame);
    };
  }, [visible]);

  if (!visible) return null;

  return (
    <Modal
      animationType={Platform.OS === "web" ? "none" : "fade"}
      onRequestClose={onDismiss}
      transparent
      visible={visible}
    >
      <View accessibilityViewIsModal style={styles.backdrop}>
        {Platform.OS === "web" ? (
          <View
            pointerEvents="none"
            style={[styles.webScrim, webScrimVisible && styles.webScrimVisible]}
          />
        ) : (
          <BlurView
            blurMethod="dimezisBlurViewSdk31Plus"
            intensity={45}
            pointerEvents="none"
            style={styles.scrim}
            tint="dark"
          />
        )}
        {Platform.OS === "android" ? (
          <View
            pointerEvents="none"
            style={[styles.androidDimmingLayer, { backgroundColor: colors.surface.canvas }]}
          />
        ) : null}
        <Pressable
          accessibilityLabel={dismissAccessibilityLabel}
          accessibilityRole="button"
          onPress={onDismiss}
          style={styles.dismissArea}
        />
        <View
          style={[
            styles.dialog,
            { backgroundColor: colors.surface.default, borderColor: colors.border.default },
          ]}
        >
          {children}
        </View>
      </View>
    </Modal>
  );
}

const styles = StyleSheet.create({
  actions: { gap: space[3] },
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
  webScrim: {
    ...StyleSheet.absoluteFill,
    backdropFilter: "blur(0px)",
    backgroundColor: "rgba(0, 0, 0, 0)",
    transitionDuration: "160ms",
    transitionProperty: "background-color, backdrop-filter",
  },
  webScrimVisible: {
    backdropFilter: "blur(8px)",
    backgroundColor: "rgba(0, 0, 0, 0.35)",
  },
});
