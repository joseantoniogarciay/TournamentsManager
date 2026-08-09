import { router, type Href, useFocusEffect } from "expo-router";
import { useCallback, useRef, useState } from "react";
import {
  ActivityIndicator,
  Pressable,
  RefreshControl,
  ScrollView,
  StyleSheet,
  View,
} from "react-native";

import { color, control, radius, space } from "@tournaments-manager/design-tokens";

import { APISessionInvalidatedError } from "@/api/fetch";
import { getTranslator } from "@/shared/i18n/locale";
import { listRelatedLeagues } from "@/features/league-creation/api";
import { LeagueCreatorChip } from "@/features/league-creation/components/league-creator-chip";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getLeagueStateLabel } from "@/shared/i18n/league-state";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { useSession } from "@/shared/session/session-provider";
import { Card, Screen, Text, useTabContentBottomPadding } from "@/shared/ui";

type LeagueRelationship = "administered" | "followed";
const floatingActionButtonSize = control.minHeight + 12;

export default function TournamentsScreen() {
  const t = getTranslator();
  const { isRestoring, revision, user } = useSession();
  const { show } = useFeedback();
  const { colors } = usePreferences();
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
      <ScrollView
        contentContainerStyle={[
          styles.content,
          { paddingBottom: tabContentBottomPadding + floatingActionButtonSize + space[5] },
        ]}
        key={revision}
        refreshControl={
          user ? (
            <RefreshControl
              onRefresh={() => void loadLeagues(true)}
              refreshing={isRefreshing}
              tintColor={colors.border.focus}
            />
          ) : undefined
        }
        showsVerticalScrollIndicator={false}
      >
        {!isRestoring && !user ? (
          <Card>
            <View style={styles.copy}>
              <Text variant="title">{t("tournaments_title")}</Text>
              <Text color="secondary">{t("tournaments_description")}</Text>
            </View>
          </Card>
        ) : null}

        {!isRestoring && user && hasLoadedLeagues && !isLoading ? (
          <LeagueLibrary
            administered={administered}
            followed={followed}
            selectedRelationship={selectedRelationship}
            onSelectRelationship={setSelectedRelationship}
          />
        ) : null}
      </ScrollView>
      {isInitialLoad ? (
        <View
          accessibilityLabel={t("common_loading")}
          accessibilityRole="progressbar"
          style={styles.loader}
        >
          <ActivityIndicator color={colors.text.primary} size="large" />
        </View>
      ) : null}
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
  followed,
  selectedRelationship,
  onSelectRelationship,
}: {
  administered: Awaited<ReturnType<typeof listRelatedLeagues>>;
  followed: Awaited<ReturnType<typeof listRelatedLeagues>>;
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
      {leagues.length === 0 ? (
        <View style={styles.empty}>
          <Text color="secondary">{empty}</Text>
        </View>
      ) : (
        leagues.map((league) => (
          <Card key={league.id}>
            <Pressable
              accessibilityLabel={t("tournaments_open_league").replace("{name}", league.name)}
              accessibilityRole="button"
              onPress={() => router.push(`/league/${league.id}` as never)}
              style={styles.row}
            >
              <View style={styles.copy}>
                <Text>{league.name}</Text>
                <View style={styles.leagueState}>
                  {league.relationship === "organizer" ? <LeagueCreatorChip /> : null}
                  <Text color="secondary">{getLeagueStateLabel(t, league.state)}</Text>
                </View>
              </View>
              <Text color="secondary">›</Text>
            </Pressable>
          </Card>
        ))
      )}
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
      <Text color={selected ? "onBrand" : "primary"}>{label}</Text>
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
  content: { flexGrow: 1 },
  copy: { gap: space[2] },
  leagueState: { alignItems: "center", flexDirection: "row", flexWrap: "wrap", gap: space[2] },
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
  library: { flex: 1, gap: space[5] },
  loader: {
    alignItems: "center",
    bottom: 0,
    justifyContent: "center",
    left: 0,
    position: "absolute",
    right: 0,
    top: 0,
  },
  row: {
    alignItems: "center",
    flexDirection: "row",
    justifyContent: "space-between",
    minHeight: 44,
  },
  segment: {
    alignItems: "center",
    borderRadius: radius.control - 2,
    flex: 1,
    flexDirection: "row",
    gap: space[2],
    justifyContent: "center",
    minHeight: control.minHeight,
  },
  segmentedBar: {
    borderRadius: radius.control,
    flexDirection: "row",
    marginHorizontal: space[5],
    padding: space[1],
  },
  segmentSelected: { backgroundColor: color.brand.primary },
});
