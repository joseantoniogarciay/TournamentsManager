import { Redirect } from "expo-router";
import { ScrollView, StyleSheet, View } from "react-native";
import { space } from "@tournaments-manager/design-tokens";
import { useSessions } from "@/features/session-management/use-sessions";
import { getCurrentLanguage, getTranslator } from "@/shared/i18n/locale";
import { useSession } from "@/shared/session/session-provider";
import {
  Button,
  Card,
  LoadingTransition,
  RequestErrorCard,
  Screen,
  Text,
  useConfirmationDialog,
  useTabContentBottomPadding,
} from "@/shared/ui";

export default function SessionsScreen() {
  const t = getTranslator();
  const { user, isRestoring } = useSession();
  const { page, loadError, busy, pending, load, revoke } = useSessions(user?.id);
  const bottomPadding = useTabContentBottomPadding();
  const { confirm } = useConfirmationDialog();
  const confirmRevoke = (id?: string) => {
    if (busy) return;
    confirm({
      title: t(id ? "sessions_revoke_title" : "sessions_revoke_others"),
      description: t(id ? "sessions_revoke_description" : "sessions_revoke_others_description"),
      acceptLabel: t(id ? "sessions_revoke" : "sessions_revoke_others"),
      acceptVariant: "destructive",
      cancelLabel: t("common_cancel"),
      onAccept: () => void revoke(id),
      onCancel: () => undefined,
    });
  };
  const formatDate = (value: string) =>
    new Date(value).toLocaleString(getCurrentLanguage(), {
      dateStyle: "medium",
      timeStyle: "short",
    });
  if (!user && !isRestoring) return <Redirect href="/account" />;
  return (
    <Screen navigationTitle={t("sessions_title")} topInset="navigation-bar" bottomInset="none">
      {loadError ? (
        <RequestErrorCard
          actionLabel={t("common_retry")}
          message={loadError}
          loading={busy}
          onRetry={() => void load()}
        />
      ) : !page ? (
        <LoadingTransition active message={t("common_loading")} />
      ) : (
        <ScrollView contentContainerStyle={[styles.content, { paddingBottom: bottomPadding }]}>
          <Card>
            <View style={styles.section}>
              <Text color="secondary">{t("sessions_description")}</Text>
              <Button
                label={t("sessions_revoke_others")}
                variant="destructive"
                disabled={busy || (!page.nextCursor && !page.items.some((item) => !item.current))}
                loading={pending === "others"}
                onPress={() => confirmRevoke()}
              />
              <Button
                label={t("sessions_refresh")}
                variant="secondary"
                disabled={busy}
                onPress={() => void load()}
              />
            </View>
          </Card>
          {page.items.map((item) => (
            <Card key={item.id}>
              <View style={styles.section}>
                <Text variant="title">
                  {t(`sessions_client_${item.client}`)} · {t(`sessions_platform_${item.platform}`)}
                </Text>
                {item.current ? <Text>{t("sessions_current")}</Text> : null}
                <Text color="secondary">
                  {t("sessions_created").replace("{date}", formatDate(item.createdAt))}
                </Text>
                <Text color="secondary">
                  {t("sessions_last_seen").replace("{date}", formatDate(item.lastSeenAt))}
                </Text>
                {!item.current ? (
                  <Button
                    label={t("sessions_revoke")}
                    variant="destructive"
                    disabled={busy}
                    loading={pending === item.id}
                    onPress={() => confirmRevoke(item.id)}
                  />
                ) : null}
              </View>
            </Card>
          ))}
          {page.items.length === 0 ? (
            <View style={styles.empty}>
              <Text color="secondary">{t("sessions_empty")}</Text>
            </View>
          ) : null}
          {page.nextCursor ? (
            <Card>
              <Button
                label={t("sessions_more")}
                variant="secondary"
                loading={busy && !pending}
                disabled={busy}
                onPress={() => void load(page.nextCursor ?? undefined)}
              />
            </Card>
          ) : null}
        </ScrollView>
      )}
    </Screen>
  );
}
const styles = StyleSheet.create({
  content: { gap: space[5] },
  section: { gap: space[3] },
  empty: { paddingHorizontal: space[5] },
});
