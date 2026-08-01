import { Platform, ScrollView, type ScrollViewProps } from "react-native";

/** Mantiene el campo activo desplazable por encima del teclado en iOS. */
export function KeyboardAwareScrollView({
  keyboardShouldPersistTaps = "handled",
  ...props
}: ScrollViewProps) {
  return (
    <ScrollView
      automaticallyAdjustKeyboardInsets={Platform.OS === "ios"}
      keyboardShouldPersistTaps={keyboardShouldPersistTaps}
      {...props}
    />
  );
}
