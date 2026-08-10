import { router, Stack, useLocalSearchParams } from "expo-router";
import { SymbolView } from "expo-symbols";
import { useEffect, useState } from "react";
import { ActivityIndicator, Platform, Pressable, ScrollView, StyleSheet } from "react-native";

import { control, space } from "@tournaments-manager/design-tokens";

import {
  assignLeagueAdministratorRequest,
  searchPublicUsernames,
  UserSearchRateLimitedError,
} from "@/features/league-creation/api";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import { getTranslator } from "@/shared/i18n/locale";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { Screen, Text, TextField } from "@/shared/ui";
import { WebIcon } from "@/shared/ui/web-icon";

const minimumQueryLength = 3;
const debounceMilliseconds = 400;

export default function AddLeagueAdministratorScreen() {
  const t = getTranslator();
  const { id } = useLocalSearchParams<{ id: string }>();
  const { colors } = usePreferences();
  const { show } = useFeedback();
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<string[]>([]);
  const [searching, setSearching] = useState(false);
  const [assigning, setAssigning] = useState<string>();

  useEffect(() => {
    if (query.length < minimumQueryLength) {
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
  }, [query, show, t]);

  const close = () => {
    if (router.canDismiss()) router.dismiss();
    else router.replace(id ? `/league/${id}` : "/");
  };
  const assign = async (username: string) => {
    if (!id || assigning) return;
    setAssigning(username);
    try {
      await assignLeagueAdministratorRequest(id, username);
      show({ kind: "success", message: t("league_administrator_added") });
      close();
    } catch (error) {
      const failure = getRequestFailure(error);
      show({ kind: failure.kind, message: t(failure.messageKey) });
    } finally {
      setAssigning(undefined);
    }
  };
  const closeButton = (
    <Pressable
      accessibilityLabel={t("common_close")}
      accessibilityRole="button"
      onPress={close}
      style={styles.closeButton}
    >
      {Platform.OS === "web" ? (
        <WebIcon color={colors.text.primary} name="close" size={control.iconSize} />
      ) : (
        <SymbolView name="xmark" size={control.iconSize} tintColor={colors.text.primary} />
      )}
    </Pressable>
  );
  const helper =
    query.length < minimumQueryLength
      ? t("league_administrator_search_minimum").replace("{count}", String(minimumQueryLength))
      : !searching && results.length === 0
        ? t("league_administrator_search_empty")
        : undefined;

  return (
    <>
      <Stack.Screen
        options={{
          headerBackVisible: false,
          headerLeft: () => closeButton,
          headerShadowVisible: false,
          headerStyle: { backgroundColor: colors.surface.canvas },
          headerTintColor: colors.text.primary,
          title: t("league_add_administrator"),
        }}
      />
      <Screen topInset="navigation-bar">
        <ScrollView
          contentContainerStyle={styles.content}
          keyboardDismissMode="on-drag"
          keyboardShouldPersistTaps="handled"
        >
          <TextField
            autoCapitalize="none"
            autoFocus
            label={t("league_administrator_username")}
            onChangeText={(value) => setQuery(value.toLowerCase())}
            placeholder={t("league_administrator_username")}
            value={query}
          />
          {searching ? <ActivityIndicator color={colors.text.primary} /> : null}
          {helper ? <Text color="secondary">{helper}</Text> : null}
          {results.map((username) => (
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
    height: control.minHeight,
    justifyContent: "center",
    width: control.minHeight,
  },
  content: { gap: space[3], paddingBottom: space[5] },
  row: {
    alignItems: "center",
    borderBottomWidth: 1,
    flexDirection: "row",
    justifyContent: "space-between",
    minHeight: control.minHeight,
    paddingHorizontal: space[2],
  },
});
