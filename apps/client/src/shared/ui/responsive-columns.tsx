import { type PropsWithChildren } from "react";
import { View } from "react-native";

import { space } from "@tournaments-manager/design-tokens";

export type ResponsiveColumnsProps = PropsWithChildren<{
  gap?: number;
  fullWidthLast?: boolean;
  variant?: "equal" | "sidebar";
}>;

export function ResponsiveColumns({ children, gap = space[5] }: ResponsiveColumnsProps) {
  return <View style={{ gap }}>{children}</View>;
}
