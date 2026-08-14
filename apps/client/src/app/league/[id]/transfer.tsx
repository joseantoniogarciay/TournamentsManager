import { router, Stack, useLocalSearchParams } from "expo-router";
import { useEffect, useState } from "react";
import { ActivityIndicator, Platform, Pressable, ScrollView, StyleSheet } from "react-native";

import { control, space } from "@tournaments-manager/design-tokens";

import {
  searchPublicUsernames,
  transferLeagueOwnershipRequest,
  UserSearchRateLimitedError,
} from "@/features/league-creation/api";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { useSession } from "@/shared/session/session-provider";
import { NavigationHeaderButton, Screen, Text, TextField } from "@/shared/ui";

const minimumQueryLength = 3;
const debounceMilliseconds = 400;

export default function TransferLeagueScreen() {
  const t = getTranslator();
  const { id } = useLocalSearchParams<{ id: string }>();
  const { colors } = usePreferences();
  const { show, showAfterNavigation } = useFeedback();
  const { user } = useSession();
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<string[]>([]);
  const [searching, setSearching] = useState(false);
  const [transferring, setTransferring] = useState<string>();

  useEffect(() => {
    if (query.length < minimumQueryLength) {
      setResults([]);
      return;
    }
    const controller = new AbortController();
    setSearching(true);
    const timer = setTimeout(() => {
      void searchPublicUsernames(query, controller.signal)
        .then(setResults)
        .catch((error) => {
          if (error instanceof Error && error.name === "AbortError") return;
          setResults([]);
          show({
            kind: "generic-error",
            message:
              error instanceof UserSearchRateLimitedError
                ? t("league_administrator_search_rate_limited")
                : t(getRequestFailure(error).messageKey),
          });
        })
        .finally(() => setSearching(false));
    }, debounceMilliseconds);
    return () => {
      clearTimeout(timer);
      controller.abort();
    };
  }, [query, show, t]);

  const close = () => {
    if (router.canDismiss()) router.dismiss();
    else router.replace(id ? `/league/${id}` : "/");
  };
  const transfer = async (username: string) => {
    if (!id || transferring) return;
    setTransferring(username);
    try {
      await transferLeagueOwnershipRequest(id, username);
      showAfterNavigation({ kind: "success", message: t("league_transfer_success") });
      close();
    } catch (error) {
      const failure = getRequestFailure(error);
      show({ kind: failure.kind, message: t(failure.messageKey) });
    } finally {
      setTransferring(undefined);
    }
  };
  const visibleResults = results.filter((username) => username !== user?.username);
  const helper =
    query.length < minimumQueryLength
      ? t("league_administrator_search_minimum").replace("{count}", String(minimumQueryLength))
      : !searching && visibleResults.length === 0
        ? t("league_administrator_search_empty")
        : undefined;
  const closeButton = (
    <NavigationHeaderButton
      accessibilityLabel={t("common_close")}
      icon="close"
      nativeIcon={{ android: "close", ios: "xmark", web: "close" }}
      onPress={close}
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
          title: t("league_transfer"),
        }}
      >
        {Platform.OS === "ios" ? (
          <Stack.Toolbar placement="left">
            <Stack.Toolbar.Button
              accessibilityLabel={t("common_close")}
              icon="xmark"
              onPress={close}
            />
          </Stack.Toolbar>
        ) : null}
      </Stack.Screen>
      {Platform.OS !== "ios" ? <Stack.Screen options={{ headerLeft: () => closeButton }} /> : null}
      <Screen topInset="navigation-bar">
        <ScrollView
          contentContainerStyle={styles.content}
          keyboardDismissMode="on-drag"
          keyboardShouldPersistTaps="handled"
        >
          <Text color="secondary">{t("league_transfer_description")}</Text>
          <TextField
            accessibilityLabel={t("league_administrator_username")}
            autoCapitalize="none"
            autoFocus
            onChangeText={(value) => setQuery(value.toLowerCase())}
            placeholder={t("league_administrator_username_placeholder")}
            value={query}
          />
          {searching ? <ActivityIndicator color={colors.text.primary} /> : null}
          {helper ? <Text color="secondary">{helper}</Text> : null}
          {visibleResults.map((username) => (
            <Pressable
              key={username}
              accessibilityRole="button"
              disabled={transferring !== undefined}
              onPress={() => void transfer(username)}
              style={[styles.row, { borderColor: colors.border.default }]}
            >
              <Text variant="bodyLarge">{username}</Text>
              {transferring === username ? <ActivityIndicator color={colors.text.primary} /> : null}
            </Pressable>
          ))}
        </ScrollView>
      </Screen>
    </>
  );
}

const styles = StyleSheet.create({
  content: { gap: space[3], paddingBottom: space[5], paddingHorizontal: space[5] },
  row: {
    alignItems: "center",
    borderBottomWidth: 1,
    flexDirection: "row",
    justifyContent: "space-between",
    minHeight: control.minHeight,
    paddingHorizontal: space[2],
  },
});
