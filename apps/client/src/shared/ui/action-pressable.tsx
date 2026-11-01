import { Platform, Pressable, type PressableProps } from "react-native";

import { getTranslator } from "@/shared/i18n/locale";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { usePreferences } from "@/shared/preferences/preferences-provider";

export type ActionAvailability =
  { disabled?: false; disabledReason?: string } | { disabled: boolean; disabledReason: string };

type Props = Omit<PressableProps, "disabled"> & ActionAvailability;

// Keep the target reachable: availability guards the business action, while a
// press on an unavailable control only explains how to recover (ADR-0177).
export function ActionPressable({
  disabled = false,
  disabledReason,
  accessibilityHint,
  accessibilityState,
  onPress,
  style,
  ...props
}: Props) {
  const { show } = useFeedback();
  const { colors } = usePreferences();
  const t = getTranslator();
  const blocked = disabled || accessibilityState?.busy === true;
  const reason = accessibilityState?.busy ? t("common_action_in_progress") : disabledReason;

  const control = (
    <Pressable
      {...props}
      accessibilityState={{ ...accessibilityState, disabled: blocked }}
      accessibilityHint={blocked ? reason : accessibilityHint}
      onLongPress={blocked ? undefined : props.onLongPress}
      onPress={(event) => {
        if (blocked) {
          show({ kind: "generic-error", message: reason ?? t("common_action_in_progress") });
          return;
        }
        onPress?.(event);
      }}
      style={(state) => [
        typeof style === "function" ? style(state) : style,
        blocked && { backgroundColor: colors.surface.subtle, opacity: 1 },
      ]}
    />
  );

  // aria-disabled applies to focusable descendants. Keep it on a boxless
  // ancestor: RN Web maps it on a button to HTML disabled, suppressing the
  // activation that explains the reason. The inner native button retains its
  // keyboard behavior, layout and focus target.
  return Platform.OS === "web" ? (
    <span aria-disabled={blocked} style={{ display: "contents" }}>
      {control}
    </span>
  ) : (
    control
  );
}
