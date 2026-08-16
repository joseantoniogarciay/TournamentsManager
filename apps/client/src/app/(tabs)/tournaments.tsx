import { router, type Href, useFocusEffect } from "expo-router";
import { useCallback, useRef, useState } from "react";
import { Pressable, RefreshControl, ScrollView, StyleSheet, View } from "react-native";

import { color, control, radius, space, typography } from "@tournaments-manager/design-tokens";

import { APISessionInvalidatedError } from "@/api/fetch";
import { getTranslator } from "@/shared/i18n/locale";
import { listRelatedLeagues } from "@/features/league-creation/api";
import { LeagueCard } from "@/features/league-creation/components/league-card";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { useSession } from "@/shared/session/session-provider";
import { Card, LoadingTransition, Screen, Text, useTabContentBottomPadding } from "@/shared/ui";

type LeagueRelationship = "administered" | "followed";
const floatingActionButtonSize = control.minHeight + 12;

export default function TournamentsScreen() {
  const t = getTranslator();
  const { isRestoring, revision, user } = useSession();
  const { show } = useFeedback();
  const tabContentBottomPadding = useTabContentBottomPadding();
  const [administered, setAdministered] = useState<Awaited<ReturnType<typeof listRelatedLeagues>>>(
    [],
  );
  const [followed, setFollowed] = useState<Awaited<ReturnType<typeof listRelatedLeagues>>>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [hasLoadedLeagues, setHasLoadedLeagues] = useState(false);
  const [selectedRelationship, setSelectedRelationship] =
    useState<LeagueRelationship>("administered");
  const loadedAccountID = useRef<string | null>(null);
  const isInitialLoad = !isRestoring && Boolean(user) && !hasLoadedLeagues;
  const showFloatingAction = !isRestoring && (!user || !isInitialLoad);

  const loadLeagues = useCallback(
    async (isManualRefresh = false) => {
      if (!user) return;
      if (isManualRefresh) setIsRefreshing(true);
      else setIsLoading(true);
      try {
        const [nextAdministered, nextFollowed] = await Promise.all([
          listRelatedLeagues("administered"),
          listRelatedLeagues("followed"),
        ]);
        setAdministered(nextAdministered);
        setFollowed(nextFollowed);
        setSelectedRelationship(
          nextAdministered.length > 0 || nextFollowed.length === 0 ? "administered" : "followed",
        );
      } catch (error) {
        if (error instanceof APISessionInvalidatedError) return;
        const failure = getRequestFailure(error);
        show({ kind: failure.kind, message: t(failure.messageKey) });
      } finally {
        setHasLoadedLeagues(true);
        if (isManualRefresh) setIsRefreshing(false);
        else setIsLoading(false);
      }
    },
    [show, t, user],
  );

  useFocusEffect(
    useCallback(() => {
      if (!user) {
        loadedAccountID.current = null;
        setAdministered([]);
        setFollowed([]);
        setIsLoading(false);
        setIsRefreshing(false);
        setHasLoadedLeagues(false);
        return;
      }
      if (loadedAccountID.current === user.id) return;
      loadedAccountID.current = user.id;
      void loadLeagues();
    }, [loadLeagues, user]),
  );

  return (
    <Screen bottomInset="none">
      {!isRestoring && user && hasLoadedLeagues && !isLoading ? (
        <LeagueLibrary
          administered={administered}
          bottomPadding={tabContentBottomPadding + floatingActionButtonSize + space[5]}
          followed={followed}
          isRefreshing={isRefreshing}
          onRefresh={() => void loadLeagues(true)}
          onSelectRelationship={setSelectedRelationship}
          selectedRelationship={selectedRelationship}
        />
      ) : (
        <ScrollView
          contentContainerStyle={[
            styles.content,
            { paddingBottom: tabContentBottomPadding + floatingActionButtonSize + space[5] },
          ]}
          key={revision}
          showsVerticalScrollIndicator={false}
          style={styles.scroll}
        >
          {!isRestoring && !user ? (
            <Card>
              <View style={styles.copy}>
                <Text variant="title">{t("tournaments_title")}</Text>
                <Text color="secondary">{t("tournaments_description")}</Text>
              </View>
            </Card>
          ) : null}
        </ScrollView>
      )}
      {isInitialLoad ? <LoadingTransition active message={t("common_loading")} /> : null}
      {showFloatingAction ? (
        <View style={[styles.floatingAction, { bottom: tabContentBottomPadding - space[4] }]}>
          <CreateTournamentButton />
        </View>
      ) : null}
    </Screen>
  );
}

function LeagueLibrary({
  administered,
  bottomPadding,
  followed,
  isRefreshing,
  onRefresh,
  selectedRelationship,
  onSelectRelationship,
}: {
  administered: Awaited<ReturnType<typeof listRelatedLeagues>>;
  bottomPadding: number;
  followed: Awaited<ReturnType<typeof listRelatedLeagues>>;
  isRefreshing: boolean;
  onRefresh: () => void;
  selectedRelationship: LeagueRelationship;
  onSelectRelationship: (relationship: LeagueRelationship) => void;
}) {
  const t = getTranslator();
  const { colors } = usePreferences();
  const leagues = selectedRelationship === "administered" ? administered : followed;
  const empty =
    selectedRelationship === "administered"
      ? t("tournaments_administered_empty")
      : t("tournaments_followed_empty");

  return (
    <View style={styles.library}>
      <View
        accessibilityRole="tablist"
        style={[styles.segmentedBar, { backgroundColor: colors.surface.subtle }]}
      >
        <Segment
          count={administered.length}
          label={t("tournaments_administered")}
          selected={selectedRelationship === "administered"}
          onPress={() => onSelectRelationship("administered")}
        />
        <Segment
          count={followed.length}
          label={t("tournaments_followed")}
          selected={selectedRelationship === "followed"}
          onPress={() => onSelectRelationship("followed")}
        />
      </View>
      <ScrollView
        contentContainerStyle={[styles.libraryContent, { paddingBottom: bottomPadding }]}
        refreshControl={
          <RefreshControl
            onRefresh={onRefresh}
            refreshing={isRefreshing}
            colors={[colors.indicator.default]}
            tintColor={colors.indicator.default}
          />
        }
        showsVerticalScrollIndicator={false}
        style={styles.scroll}
      >
        {leagues.length === 0 ? (
          <View style={styles.empty}>
            <Text color="secondary">{empty}</Text>
          </View>
        ) : (
          leagues.map((league) => <LeagueCard key={league.id} league={league} />)
        )}
      </ScrollView>
    </View>
  );
}

function Segment({
  count,
  label,
  selected,
  onPress,
}: {
  count: number;
  label: string;
  selected: boolean;
  onPress: () => void;
}) {
  return (
    <Pressable
      accessibilityRole="tab"
      accessibilityState={{ selected }}
      onPress={onPress}
      style={[styles.segment, selected ? styles.segmentSelected : undefined]}
    >
      <Text color={selected ? "onBrand" : "primary"} style={styles.segmentLabel}>
        {label}
      </Text>
      <Text color={selected ? "onBrand" : "secondary"}>{count}</Text>
    </Pressable>
  );
}

function CreateTournamentButton() {
  const t = getTranslator();

  return (
    <Pressable
      accessibilityLabel={t("home_create_tournament")}
      accessibilityRole="button"
      onPress={() => router.push("/create-tournament" as Href)}
    >
      <View style={styles.floatingActionButton}>
        <Text color="onBrand" variant="title">
          {t("common_add")}
        </Text>
      </View>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  scroll: { flex: 1, minHeight: 0 },
  content: { flexGrow: 1 },
  copy: { gap: space[2] },
  empty: { alignItems: "center", flex: 1, justifyContent: "center", paddingHorizontal: space[5] },
  floatingAction: { position: "absolute", right: space[5] },
  floatingActionButton: {
    alignItems: "center",
    backgroundColor: color.brand.primary,
    borderRadius: radius.pill,
    height: floatingActionButtonSize,
    justifyContent: "center",
    width: floatingActionButtonSize,
  },
  library: { flex: 1, minHeight: 0, overflow: "hidden" },
  libraryContent: { flexGrow: 1, gap: space[5] },
  segment: {
    alignItems: "center",
    borderRadius: radius.control - 2,
    flex: 1,
    flexDirection: "row",
    gap: space[2],
    justifyContent: "center",
    minHeight: control.minHeight,
  },
  segmentLabel: { fontFamily: typography.family.bold },
  segmentedBar: {
    borderRadius: radius.control,
    flexDirection: "row",
    marginBottom: space[5],
    marginHorizontal: space[5],
    padding: space[1],
  },
  segmentSelected: { backgroundColor: color.brand.primary },
});
