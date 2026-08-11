import { router, Stack } from "expo-router";
import { useEffect, useState } from "react";
import { Pressable, ScrollView, StyleSheet, View } from "react-native";
import { space } from "@tournaments-manager/design-tokens";
import type { Notification } from "@/api/generated/models";
import {
  deleteAllNotifications,
  deleteNotification,
  listNotifications,
  markAllNotificationsRead,
} from "@/features/notifications/api";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { Button, Card, Screen, Text, useConfirmationDialog } from "@/shared/ui";
import { useNotifications } from "@/features/notifications/notification-provider";

export default function NotificationsScreen() {
  const t = getTranslator();
  const { colors } = usePreferences();
  const { show } = useFeedback();
  const { confirm } = useConfirmationDialog();
  const { refresh } = useNotifications();
  const [items, setItems] = useState<Notification[]>();
  useEffect(() => {
    void Promise.all([markAllNotificationsRead(), listNotifications()])
      .then(([, next]) => {
        setItems(next);
        return refresh(true);
      })
      .catch((error) => {
        setItems([]);
        const failure = getRequestFailure(error);
        show({ kind: failure.kind, message: t(failure.messageKey) });
      });
  }, [refresh, show, t]);
  const remove = async (id: string) => {
    try {
      await deleteNotification(id);
      setItems((current) => current?.filter((item) => item.id !== id));
    } catch (error) {
      const failure = getRequestFailure(error);
      show({ kind: failure.kind, message: t(failure.messageKey) });
    }
  };
  const removeAll = () =>
    confirm({
      title: t("notifications_delete_all_title"),
      description: t("notifications_delete_all_description"),
      acceptLabel: t("notifications_delete_all"),
      acceptVariant: "destructive",
      cancelLabel: t("common_cancel"),
      onCancel: () => undefined,
      onAccept: async () => {
        try {
          await deleteAllNotifications();
          setItems([]);
        } catch (error) {
          const failure = getRequestFailure(error);
          show({ kind: failure.kind, message: t(failure.messageKey) });
        }
      },
    });
  return (
    <Screen topInset="navigation-bar">
      <Stack.Screen options={{ title: t("notifications_title") }} />
      <ScrollView contentContainerStyle={styles.content}>
        {items === undefined ? (
          <Text>{t("common_loading")}</Text>
        ) : items.length === 0 ? (
          <Text color="secondary" style={styles.empty}>
            {t("notifications_empty")}
          </Text>
        ) : (
          <>
            {items.map((item) => (
              <Card density="compact" key={item.id}>
                <View style={styles.row}>
                  <Pressable
                    accessibilityRole="button"
                    accessibilityLabel={t("notifications_open_league").replace(
                      "{league}",
                      item.leagueName,
                    )}
                    onPress={() => router.push(`/league/${item.leagueId}`)}
                    style={styles.message}
                  >
                    <Text variant="body">
                      {t("notifications_administrator_assigned").replace(
                        "{league}",
                        item.leagueName,
                      )}
                    </Text>
                    <Text color="secondary">{new Date(item.createdAt).toLocaleString()}</Text>
                  </Pressable>
                  <Pressable
                    accessibilityRole="button"
                    accessibilityLabel={t("notifications_delete")}
                    onPress={() => void remove(item.id)}
                    style={[styles.delete, { borderColor: colors.border.default }]}
                  >
                    <Text>×</Text>
                  </Pressable>
                </View>
              </Card>
            ))}
            <Button label={t("notifications_delete_all")} onPress={removeAll} variant="secondary" />
          </>
        )}
      </ScrollView>
    </Screen>
  );
}
const styles = StyleSheet.create({
  content: { gap: space[5], paddingBottom: space[8] },
  empty: { paddingHorizontal: space[5] },
  row: { alignItems: "center", flexDirection: "row", gap: space[3] },
  message: { flex: 1, gap: space[1] },
  delete: { alignItems: "center", borderWidth: 1, height: 44, justifyContent: "center", width: 44 },
});
