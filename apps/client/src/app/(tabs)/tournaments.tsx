import { router, type Href } from "expo-router";
import { Pressable, RefreshControl, ScrollView, StyleSheet, View } from "react-native";

import { color, control, radius, space, typography } from "@tournaments-manager/design-tokens";

import { getTranslator } from "@/shared/i18n/locale";
import type { AccountTournamentPage } from "@/api/generated/models";
import {
  useTournamentLibrary,
  type TournamentRelationship,
} from "@/features/league-creation/use-tournament-library";
import { TournamentCard } from "@/features/league-creation/components/league-card";
import { usePreferences } from "@/shared/preferences/preferences-provider";
import { useSession } from "@/shared/session/session-provider";
import {
  Button,
  Card,
  LoadingTransition,
  RequestErrorCard,
  Screen,
  Text,
  useTabContentBottomPadding,
} from "@/shared/ui";

const floatingActionButtonSize = control.minHeight + 12;

export default function TournamentsScreen() {
  const t = getTranslator();
  const { isRestoring, revision, user } = useSession();
  const tabContentBottomPadding = useTabContentBottomPadding();
  const {
    administered,
    followed,
    isLoading,
    isRefreshing,
    isLoadingMore,
    loadError,
    hasLoadedTournaments,
    selectedRelationship,
    setSelectedRelationship,
    loadTournaments,
    loadMore,
  } = useTournamentLibrary();
  const isInitialLoad = !isRestoring && Boolean(user) && !hasLoadedTournaments;
  const showFloatingAction = !isRestoring && (!user || !isInitialLoad);

  return (
    <Screen bottomInset="none">
      {!isRestoring && user && hasLoadedTournaments && loadError && !isLoading ? (
        <ScrollView
          contentContainerStyle={[styles.content, { paddingBottom: tabContentBottomPadding }]}
        >
          <RequestErrorCard
            message={t(loadError)}
            actionLabel={t("common_retry")}
            onRetry={() => void loadTournaments()}
          />
        </ScrollView>
      ) : !isRestoring && user && hasLoadedTournaments && !isLoading ? (
        <TournamentLibrary
          administered={administered}
          bottomPadding={tabContentBottomPadding + floatingActionButtonSize + space[5]}
          followed={followed}
          isRefreshing={isRefreshing}
          isLoadingMore={isLoadingMore}
          onLoadMore={() => void loadMore()}
          onRefresh={() => void loadTournaments(true)}
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
      {isInitialLoad || isLoading ? (
        <LoadingTransition active message={t("common_loading")} />
      ) : null}
      {showFloatingAction ? (
        <View style={[styles.floatingAction, { bottom: tabContentBottomPadding - space[4] }]}>
          <CreateTournamentButton />
        </View>
      ) : null}
    </Screen>
  );
}

function TournamentLibrary({
  administered,
  bottomPadding,
  followed,
  isRefreshing,
  isLoadingMore,
  onLoadMore,
  onRefresh,
  selectedRelationship,
  onSelectRelationship,
}: {
  administered: AccountTournamentPage;
  bottomPadding: number;
  followed: AccountTournamentPage;
  isRefreshing: boolean;
  isLoadingMore: boolean;
  onLoadMore: () => void;
  onRefresh: () => void;
  selectedRelationship: TournamentRelationship;
  onSelectRelationship: (relationship: TournamentRelationship) => void;
}) {
  const t = getTranslator();
  const { colors } = usePreferences();
  const page = selectedRelationship === "administered" ? administered : followed;
  const leagues = page.items;
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
          count={t("tournaments_loaded_count")
            .replace("{count}", String(administered.items.length))
            .replace("{more}", administered.nextCursor ? "+" : "")}
          label={t("tournaments_administered")}
          selected={selectedRelationship === "administered"}
          onPress={() => onSelectRelationship("administered")}
        />
        <Segment
          count={t("tournaments_loaded_count")
            .replace("{count}", String(followed.items.length))
            .replace("{more}", followed.nextCursor ? "+" : "")}
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
          leagues.map((league) => <TournamentCard key={league.id} league={league} />)
        )}
        {page.nextCursor ? (
          <View style={styles.pagination}>
            <Button
              label={t("tournaments_load_more")}
              loading={isLoadingMore}
              disabled={isRefreshing}
              onPress={onLoadMore}
              variant="secondary"
            />
          </View>
        ) : null}
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
  count: string;
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
  pagination: { marginHorizontal: space[5] },
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
