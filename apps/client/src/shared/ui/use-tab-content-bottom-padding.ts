import { Platform } from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";

import { space } from "@tournaments-manager/design-tokens";

/** Reserva en el contenido desplazable el espacio de una tab superpuesta. */
export function useTabContentBottomPadding() {
  const { bottom } = useSafeAreaInsets();
  return (Platform.OS === "web" ? space[10] : bottom) + space[12];
}
