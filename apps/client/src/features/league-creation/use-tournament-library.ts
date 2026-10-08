import { useFocusEffect } from "expo-router";
import { useCallback, useEffect, useRef, useState } from "react";

import { APISessionInvalidatedError } from "@/api/fetch";
import type { AccountTournamentPage } from "@/api/generated/models";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getTranslator, type TranslationKey } from "@/shared/i18n/locale";
import { useSession } from "@/shared/session/session-provider";
import { listRelatedTournaments } from "./api";
import { getTournamentLibraryRevision } from "./tournament-library-revision";

export type TournamentRelationship = "administered" | "followed";

export function useTournamentLibrary() {
  const t = getTranslator();
  const { user } = useSession();
  const { show } = useFeedback();
  const [administered, setAdministered] = useState<AccountTournamentPage>({ items: [] });
  const [followed, setFollowed] = useState<AccountTournamentPage>({ items: [] });
  const [isLoading, setIsLoading] = useState(false);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [loadError, setLoadError] = useState<TranslationKey | null>(null);
  const [hasLoadedTournaments, setHasLoadedTournaments] = useState(false);
  const [selectedRelationship, setSelectedRelationship] =
    useState<TournamentRelationship>("administered");
  const loadedAccountID = useRef<string | null>(null);
  const loadedRevision = useRef(-1);
  const generation = useRef(0);
  const refreshInFlight = useRef(false);
  const appendInFlight = useRef(false);

  useEffect(
    () => () => {
      generation.current += 1;
      loadedAccountID.current = null;
      refreshInFlight.current = false;
      appendInFlight.current = false;
    },
    [user?.id],
  );

  const loadTournaments = useCallback(
    async (isManualRefresh = false) => {
      if (!user || refreshInFlight.current) return;
      refreshInFlight.current = true;
      const requestGeneration = ++generation.current;
      appendInFlight.current = false;
      setIsLoadingMore(false);
      if (isManualRefresh) setIsRefreshing(true);
      else setIsLoading(true);
      try {
        const [nextAdministered, nextFollowed] = await Promise.all([
          listRelatedTournaments("administered"),
          listRelatedTournaments("followed"),
        ]);
        if (requestGeneration !== generation.current) return;
        setLoadError(null);
        setAdministered(nextAdministered);
        setFollowed(nextFollowed);
        if (!isManualRefresh)
          setSelectedRelationship(
            nextAdministered.items.length > 0 || nextFollowed.items.length === 0
              ? "administered"
              : "followed",
          );
      } catch (error) {
        if (requestGeneration !== generation.current || error instanceof APISessionInvalidatedError)
          return;
        const failure = getRequestFailure(error);
        if (isManualRefresh) show({ kind: failure.kind, message: t(failure.messageKey) });
        else setLoadError(failure.messageKey);
      } finally {
        if (requestGeneration === generation.current) {
          refreshInFlight.current = false;
          setHasLoadedTournaments(true);
          setIsRefreshing(false);
          setIsLoading(false);
        }
      }
    },
    [show, t, user],
  );

  const loadMore = useCallback(async () => {
    const page = selectedRelationship === "administered" ? administered : followed;
    if (!user || !page.nextCursor || appendInFlight.current || refreshInFlight.current) return;
    appendInFlight.current = true;
    setIsLoadingMore(true);
    const requestGeneration = generation.current;
    const setPage = selectedRelationship === "administered" ? setAdministered : setFollowed;
    try {
      const next = await listRelatedTournaments(selectedRelationship, page.nextCursor);
      if (requestGeneration !== generation.current) return;
      setPage({ ...next, items: [...page.items, ...next.items] });
    } catch (error) {
      if (requestGeneration !== generation.current || error instanceof APISessionInvalidatedError)
        return;
      const failure = getRequestFailure(error);
      show({ kind: failure.kind, message: t(failure.messageKey) });
    } finally {
      if (requestGeneration === generation.current) {
        appendInFlight.current = false;
        setIsLoadingMore(false);
      }
    }
  }, [administered, followed, selectedRelationship, show, t, user]);

  useFocusEffect(
    useCallback(() => {
      if (!user) {
        loadedAccountID.current = null;
        setAdministered({ items: [] });
        setFollowed({ items: [] });
        setIsLoading(false);
        setIsRefreshing(false);
        setIsLoadingMore(false);
        setLoadError(null);
        setHasLoadedTournaments(false);
        return;
      }
      const currentRevision = getTournamentLibraryRevision();
      const sameAccount = loadedAccountID.current === user.id;
      if (sameAccount && loadedRevision.current === currentRevision) return;
      loadedAccountID.current = user.id;
      loadedRevision.current = currentRevision;
      if (!sameAccount) setHasLoadedTournaments(false);
      void loadTournaments(sameAccount);
    }, [loadTournaments, user]),
  );

  return {
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
  };
}
