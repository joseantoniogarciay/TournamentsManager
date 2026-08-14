import { router, Stack } from "expo-router";
import { SymbolView } from "expo-symbols";
import { useCallback, useEffect, useState } from "react";
import { ActivityIndicator, Platform, Pressable, ScrollView, StyleSheet, View } from "react-native";
import { control, space } from "@tournaments-manager/design-tokens";
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
import {
  NavigationHeaderButton,
  Card,
  LoadingTransition,
  RequestErrorCard,
  Screen,
  Text,
  useConfirmationDialog,
} from "@/shared/ui";
import { WebIcon } from "@/shared/ui/web-icon";
import { useNotifications } from "@/features/notifications/notification-provider";

export default function NotificationsScreen() {
  const t = getTranslator();
  const { colors } = usePreferences();
  const { show } = useFeedback();
  const { confirm } = useConfirmationDialog();
  const { refresh } = useNotifications();
  const [items, setItems] = useState<Notification[]>();
  const [loadErrorMessage, setLoadErrorMessage] = useState<string>();
  const [isLoading, setIsLoading] = useState(false);
  const [removingNotificationID, setRemovingNotificationID] = useState<string>();
  const [removingAll, setRemovingAll] = useState(false);
  const load = useCallback(() => {
    setIsLoading(true);
    setLoadErrorMessage(undefined);
    void Promise.all([markAllNotificationsRead(), listNotifications()])
      .then(([, next]) => {
        setItems(next);
        return refresh(true);
      })
      .catch((error) => {
        setLoadErrorMessage(t(getRequestFailure(error).messageKey));
      })
      .finally(() => setIsLoading(false));
  }, [refresh, t]);
  useEffect(() => {
    load();
  }, [load]);
  const remove = async (id: string) => {
    if (removingNotificationID !== undefined || removingAll) return;
    setRemovingNotificationID(id);
    try {
      await deleteNotification(id);
      setItems((current) => current?.filter((item) => item.id !== id));
    } catch (error) {
      const failure = getRequestFailure(error);
      show({ kind: failure.kind, message: t(failure.messageKey) });
    } finally {
      setRemovingNotificationID(undefined);
    }
  };
  const confirmRemove = (id: string) =>
    confirm({
      title: t("notifications_delete_title"),
      description: t("notifications_delete_description"),
      acceptLabel: t("notifications_delete"),
      acceptVariant: "destructive",
      cancelLabel: t("common_cancel"),
      onAccept: () => void remove(id),
      onCancel: () => undefined,
    });
  const removeAll = () =>
    confirm({
      title: t("notifications_delete_all_title"),
      description: t("notifications_delete_all_description"),
      acceptLabel: t("notifications_delete_all"),
      acceptVariant: "destructive",
      cancelLabel: t("common_cancel"),
      onCancel: () => undefined,
      onAccept: async () => {
        if (removingAll || removingNotificationID !== undefined) return;
        setRemovingAll(true);
        try {
          await deleteAllNotifications();
          setItems([]);
        } catch (error) {
          const failure = getRequestFailure(error);
          show({ kind: failure.kind, message: t(failure.messageKey) });
        } finally {
          setRemovingAll(false);
        }
      },
    });
  const showDeleteAll = (items?.length ?? 0) > 0;
  return (
    <>
      <Stack.Screen
        options={{
          title: t("notifications_title"),
          ...(Platform.OS !== "ios" && showDeleteAll
            ? {
                headerRight: () => (
                  <NavigationHeaderButton
                    accessibilityLabel={t("notifications_delete_all")}
                    icon="trash"
                    nativeIcon={{ android: "delete", ios: "trash", web: "delete" }}
                    onPress={removeAll}
                    side="right"
                  />
                ),
              }
            : {}),
        }}
      />
      {Platform.OS === "ios" && showDeleteAll ? (
        <Stack.Toolbar placement="right">
          <Stack.Toolbar.Button
            accessibilityLabel={t("notifications_delete_all")}
            icon="trash"
            onPress={removeAll}
          />
        </Stack.Toolbar>
      ) : null}
      <Screen topInset="navigation-bar">
        {loadErrorMessage ? (
          <RequestErrorCard
            actionLabel={t("common_retry")}
            loading={isLoading}
            message={loadErrorMessage}
            onRetry={load}
          />
        ) : items === undefined ? (
          <LoadingTransition active message={t("common_loading")} />
        ) : (
          <ScrollView contentContainerStyle={styles.content}>
            {items.length === 0 ? (
              <View style={styles.empty}>
                <Text color="secondary">{t("notifications_empty")}</Text>
              </View>
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
                        <Text color="secondary">
                          {formatNotificationDate(item.createdAt, t("notifications_invalid_date"))}
                        </Text>
                      </Pressable>
                      <Pressable
                        accessibilityRole="button"
                        accessibilityLabel={t("notifications_delete")}
                        accessibilityState={{ busy: removingNotificationID === item.id }}
                        disabled={removingNotificationID !== undefined || removingAll}
                        onPress={() => confirmRemove(item.id)}
                        style={styles.delete}
                      >
                        {removingNotificationID === item.id ? (
                          <ActivityIndicator color={colors.text.primary} />
                        ) : Platform.OS === "web" ? (
                          <WebIcon
                            color={colors.text.primary}
                            name="close"
                            size={control.iconSize}
                          />
                        ) : (
                          <SymbolView
                            name={{ android: "close", ios: "xmark", web: "close" }}
                            size={control.iconSize}
                            tintColor={colors.text.primary}
                          />
                        )}
                      </Pressable>
                    </View>
                  </Card>
                ))}
              </>
            )}
          </ScrollView>
        )}
      </Screen>
    </>
  );
}

function formatNotificationDate(value: string, fallback: string) {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? fallback : date.toLocaleString();
}

const styles = StyleSheet.create({
  content: { flexGrow: 1, gap: space[5], paddingBottom: space[8] },
  empty: { alignItems: "center", flex: 1, justifyContent: "center", paddingHorizontal: space[5] },
  row: { alignItems: "center", flexDirection: "row", gap: space[3] },
  message: { flex: 1, gap: space[1] },
  delete: {
    alignItems: "center",
    height: control.minHeight,
    justifyContent: "center",
    width: control.minHeight,
  },
});
