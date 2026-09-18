import { useMemo, useState } from "react";
import { StyleSheet, View } from "react-native";

import { space } from "@tournaments-manager/design-tokens";

import { SuggestionRateLimitedError, submitSuggestion } from "@/features/suggestions/api";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getTranslator } from "@/shared/i18n/locale";
import { Button, Card, Text, TextField } from "@/shared/ui";

const minimumLength = 8;
const maximumLength = 1000;

export function SuggestionCard() {
  const t = getTranslator();
  const { show } = useFeedback();
  const [body, setBody] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const normalizedLength = useMemo(() => Array.from(body.trim()).length, [body]);
  const canSubmit = normalizedLength >= minimumLength && normalizedLength <= maximumLength;

  const onSubmit = async () => {
    if (!canSubmit || isSubmitting) return;
    setIsSubmitting(true);
    try {
      await submitSuggestion(body.trim());
      setBody("");
      show({ kind: "success", message: t("home_suggestion_thanks") });
    } catch (error) {
      if (error instanceof SuggestionRateLimitedError) {
        show({ kind: "generic-error", message: t("home_suggestion_rate_limited") });
        return;
      }
      const failure = getRequestFailure(error);
      show({ kind: failure.kind, message: t(failure.messageKey) });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Card>
      <View style={styles.content}>
        <Text color="secondary">{t("home_suggestion_description")}</Text>
        <TextField
          accessibilityHint={t("home_suggestion_requirement")}
          accessibilityLabel={t("home_suggestion_accessibility_label")}
          editable={!isSubmitting}
          multiline
          numberOfLines={4}
          onChangeText={(value) => setBody(Array.from(value).slice(0, maximumLength).join(""))}
          placeholder={t("home_suggestion_placeholder")}
          value={body}
        />
        <Button
          disabled={!canSubmit}
          label={t("home_suggestion_send")}
          loading={isSubmitting}
          onPress={() => void onSubmit()}
        />
      </View>
    </Card>
  );
}

const styles = StyleSheet.create({
  content: { gap: space[4] },
});
