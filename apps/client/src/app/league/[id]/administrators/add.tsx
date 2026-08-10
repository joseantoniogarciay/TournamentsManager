import { router, Stack, useLocalSearchParams } from "expo-router";
import { SymbolView } from "expo-symbols";
import { useEffect, useState } from "react";
import { ActivityIndicator, Platform, Pressable, ScrollView, StyleSheet } from "react-native";

import { control, radius, space } from "@tournaments-manager/design-tokens";

import {
  assignLeagueAdministratorRequest,
  LeagueAdministratorConflictError,
  listLeagueAdministratorUsernames,
  searchPublicUsernames,
  UserSearchRateLimitedError,
} from "@/features/league-creation/api";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { useSession } from "@/shared/session/session-provider";
import { Screen, Text, TextField } from "@/shared/ui";
import { WebIcon } from "@/shared/ui/web-icon";

const minimumQueryLength = 3;
const debounceMilliseconds = 400;

export default function AddLeagueAdministratorScreen() {
  const t = getTranslator();
  const { id } = useLocalSearchParams<{ id: string }>();
  const { colors } = usePreferences();
  const { show } = useFeedback();
  const { user } = useSession();
  const [query, setQuery] = useState("");
  const [administrators, setAdministrators] = useState<string[]>();
  const [results, setResults] = useState<string[]>([]);
  const [searching, setSearching] = useState(false);
  const [assigning, setAssigning] = useState<string>();

  useEffect(() => {
    if (!id) return;
    void listLeagueAdministratorUsernames(id)
      .then(setAdministrators)
      .catch((error) => {
        const failure = getRequestFailure(error);
        show({ kind: failure.kind, message: t(failure.messageKey) });
      });
  }, [id, show, t]);

  useEffect(() => {
    if (administrators === undefined || query.length < minimumQueryLength) {
      setResults([]);
      setSearching(false);
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
  }, [administrators, query, show, t]);

  const close = () => {
    if (router.canDismiss()) router.dismiss();
    else router.replace(id ? `/league/${id}/administrators` : "/");
  };
  const assign = async (username: string) => {
    if (!id || assigning) return;
    setAssigning(username);
    try {
      await assignLeagueAdministratorRequest(id, username);
      close();
    } catch (error) {
      if (error instanceof LeagueAdministratorConflictError) {
        show({ kind: "generic-error", message: t("league_administrator_self_assignment") });
      } else {
        const failure = getRequestFailure(error);
        show({ kind: failure.kind, message: t(failure.messageKey) });
      }
    } finally {
      setAssigning(undefined);
    }
  };
  const closeButton = (
    <Pressable
      accessibilityLabel={t("common_close")}
      accessibilityRole="button"
      onPress={close}
      style={[
        styles.closeButton,
        { backgroundColor: colors.surface.default, borderColor: colors.border.default },
      ]}
    >
      {Platform.OS === "web" ? (
        <WebIcon color={colors.text.primary} name="close" size={control.iconSize} />
      ) : (
        <SymbolView name="xmark" size={control.iconSize} tintColor={colors.text.primary} />
      )}
    </Pressable>
  );
  const visibleResults = results.filter(
    (username) => username !== user?.username && !administrators?.includes(username),
  );
  const helper =
    administrators === undefined
      ? undefined
      : query.length < minimumQueryLength
        ? t("league_administrator_search_minimum").replace("{count}", String(minimumQueryLength))
        : !searching && visibleResults.length === 0
          ? t("league_administrator_search_empty")
          : undefined;

  return (
    <>
      <Stack.Screen
        options={{
          headerBackVisible: false,
          headerShadowVisible: false,
          headerStyle: { backgroundColor: colors.surface.canvas },
          headerTintColor: colors.text.primary,
          title: t("league_add_administrator"),
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
          <TextField
            accessibilityLabel={t("league_administrator_username")}
            autoCapitalize="none"
            autoFocus
            onChangeText={(value) => setQuery(value.toLowerCase())}
            placeholder={t("league_administrator_username_placeholder")}
            value={query}
          />
          {searching || administrators === undefined ? (
            <ActivityIndicator color={colors.text.primary} />
          ) : null}
          {helper ? <Text color="secondary">{helper}</Text> : null}
          {visibleResults.map((username) => (
            <Pressable
              key={username}
              accessibilityRole="button"
              disabled={assigning !== undefined}
              onPress={() => void assign(username)}
              style={[styles.row, { borderColor: colors.border.default }]}
            >
              <Text variant="bodyLarge">{username}</Text>
              {assigning === username ? <ActivityIndicator color={colors.text.primary} /> : null}
            </Pressable>
          ))}
        </ScrollView>
      </Screen>
    </>
  );
}

const styles = StyleSheet.create({
  closeButton: {
    alignItems: "center",
    borderRadius: radius.pill,
    borderWidth: 1,
    height: control.minHeight,
    justifyContent: "center",
    width: control.minHeight,
  },
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
