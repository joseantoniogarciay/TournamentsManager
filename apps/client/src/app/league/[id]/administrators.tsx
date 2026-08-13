import { router, Stack, useFocusEffect, useLocalSearchParams } from "expo-router";
import { SymbolView } from "expo-symbols";
import { useCallback, useState } from "react";
import { ActivityIndicator, Platform, Pressable, ScrollView, StyleSheet, View } from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";

import { control, radius, space } from "@tournaments-manager/design-tokens";

import {
  listLeagueAdministratorUsernames,
  removeLeagueAdministratorRequest,
} from "@/features/league-creation/api";
import { useLeague, useLeagueStore } from "@/features/league-creation/league-store";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { Card, NavigationHeaderButton, Screen, Text, useConfirmationDialog } from "@/shared/ui";
import { WebIcon } from "@/shared/ui/web-icon";

export default function LeagueAdministratorsScreen() {
  const t = getTranslator();
  const { id } = useLocalSearchParams<{ id: string }>();
  const { colors } = usePreferences();
  const insets = useSafeAreaInsets();
  const { show } = useFeedback();
  const { confirm } = useConfirmationDialog();
  const league = useLeague(id);
  const { loadLeague } = useLeagueStore();
  const [administrators, setAdministrators] = useState<string[]>();
  const [loadFailed, setLoadFailed] = useState(false);
  const [removingUsername, setRemovingUsername] = useState<string>();

  const load = useCallback(() => {
    if (!id) return;
    setAdministrators(undefined);
    setLoadFailed(false);
    void loadLeague(id).catch((error) => {
      const failure = getRequestFailure(error);
      show({ kind: failure.kind, message: t(failure.messageKey) });
    });
    void listLeagueAdministratorUsernames(id)
      .then(setAdministrators)
      .catch((error) => {
        setAdministrators([]);
        setLoadFailed(true);
        const failure = getRequestFailure(error);
        show({ kind: failure.kind, message: t(failure.messageKey) });
      });
  }, [id, loadLeague, show, t]);
  useFocusEffect(load);

  const close = () => {
    if (router.canDismiss()) {
      router.dismiss();
      return;
    }
    router.replace(id ? `/league/${id}` : "/");
  };
  const remove = async (username: string) => {
    if (!id || removingUsername) return;
    setRemovingUsername(username);
    try {
      await removeLeagueAdministratorRequest(id, username);
      setAdministrators((current) =>
        current?.filter((administrator) => administrator !== username),
      );
    } catch (error) {
      const failure = getRequestFailure(error);
      show({ kind: failure.kind, message: t(failure.messageKey) });
    } finally {
      setRemovingUsername(undefined);
    }
  };
  const confirmRemove = (username: string) =>
    confirm({
      title: t("league_remove_administrator_title"),
      description: t("league_remove_administrator_description").replace("{name}", username),
      acceptLabel: t("league_remove_administrator"),
      cancelLabel: t("common_cancel"),
      onAccept: () => void remove(username),
      onCancel: () => undefined,
    });
  const canAddAdministrator =
    league !== undefined && league.state !== "completed" && league.state !== "cancelled";
  const navigationButton = (
    onPress: () => void,
    label: string,
    icon: "close" | "add",
    side: "left" | "right",
  ) => (
    <NavigationHeaderButton
      accessibilityLabel={label}
      icon={icon}
      nativeIcon={
        icon === "add"
          ? { android: "add", ios: "plus", web: "add" }
          : { android: "close", ios: "xmark", web: "close" }
      }
      onPress={onPress}
      side={side}
    />
  );

  return (
    <>
      <Stack.Screen
        options={{
          headerBackVisible: false,
          headerShadowVisible: false,
          headerStyle: { backgroundColor: colors.surface.canvas },
          headerTintColor: colors.text.primary,
          headerTitleAlign: "center",
          title: t("league_administrators"),
          ...(Platform.OS !== "ios"
            ? {
                headerLeft: () => navigationButton(close, t("common_back"), "close", "left"),
                headerRight: () =>
                  canAddAdministrator
                    ? navigationButton(
                        () => router.push(`/league/${id}/administrators/add`),
                        t("league_add_administrator"),
                        "add",
                        "right",
                      )
                    : null,
              }
            : {}),
        }}
      />
      {Platform.OS === "ios" ? (
        <>
          <Stack.Toolbar placement="left">
            <Stack.Toolbar.Button
              accessibilityLabel={t("common_back")}
              icon="xmark"
              onPress={close}
            />
          </Stack.Toolbar>
          {canAddAdministrator ? (
            <Stack.Toolbar placement="right">
              <Stack.Toolbar.Button
                accessibilityLabel={t("league_add_administrator")}
                icon="plus"
                onPress={() => router.push(`/league/${id}/administrators/add`)}
              />
            </Stack.Toolbar>
          ) : null}
        </>
      ) : null}
      <Screen bottomInset="none" topInset="navigation-bar">
        {administrators === undefined ? (
          <View style={styles.loader}>
            <ActivityIndicator color={colors.text.primary} />
          </View>
        ) : (
          <ScrollView
            contentContainerStyle={[styles.content, { paddingBottom: insets.bottom + space[5] }]}
            showsVerticalScrollIndicator={false}
          >
            {!loadFailed && administrators.length === 0 ? (
              <View style={styles.emptyState}>
                <Text color="secondary">{t("league_administrators_empty")}</Text>
              </View>
            ) : null}
            {administrators.map((username) => (
              <Card density="compact" key={username}>
                <View style={styles.row}>
                  <Text style={styles.username} variant="bodyLarge">
                    {username}
                  </Text>
                  <Pressable
                    accessibilityLabel={t("league_remove_administrator")}
                    accessibilityRole="button"
                    accessibilityState={{ busy: removingUsername === username }}
                    disabled={removingUsername !== undefined}
                    onPress={() => confirmRemove(username)}
                    style={styles.removeButton}
                  >
                    {removingUsername === username ? (
                      <ActivityIndicator color={colors.text.primary} />
                    ) : Platform.OS === "web" ? (
                      <WebIcon color={colors.text.primary} name="close" size={control.iconSize} />
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
          </ScrollView>
        )}
      </Screen>
    </>
  );
}

const styles = StyleSheet.create({
  content: { flexGrow: 1, gap: space[5], paddingBottom: space[5] },
  emptyState: {
    alignItems: "center",
    flex: 1,
    justifyContent: "center",
    paddingHorizontal: space[5],
  },
  loader: { alignItems: "center", flex: 1, justifyContent: "center" },
  navigationButton: {
    alignItems: "center",
    borderRadius: radius.pill,
    borderWidth: 1,
    height: control.minHeight,
    justifyContent: "center",
    width: control.minHeight,
  },
  removeButton: {
    alignItems: "center",
    height: control.minHeight,
    justifyContent: "center",
    width: control.minHeight,
  },
  row: { alignItems: "center", flexDirection: "row", gap: space[3] },
  username: { flex: 1 },
});
