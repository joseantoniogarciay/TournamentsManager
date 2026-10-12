import { useFocusEffect } from "expo-router";
import { useCallback, useEffect, useRef, useState } from "react";
import { APISessionInvalidatedError } from "@/api/fetch";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import { getTranslator, type TranslationKey } from "@/shared/i18n/locale";

type Page<Item> = { items: Item[]; nextCursor?: string };
export function useCollection<Item, Value extends Page<Item>>(
  identity: string | null,
  fetchPage: (cursor: string | undefined, signal: AbortSignal) => Promise<Value>,
  itemKey: (item: Item) => string,
) {
  const t = getTranslator();
  const { show } = useFeedback();
  const [value, setValue] = useState<{ identity: string; page: Value } | null>(null);
  const [error, setError] = useState<TranslationKey | null>(null);
  const [moreError, setMoreError] = useState<TranslationKey | null>(null);
  const [loading, setLoading] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [loadingMore, setLoadingMore] = useState(false);
  const active = useRef<AbortController | null>(null);
  const generation = useRef(0);
  const current = useRef(value);
  current.current = value;
  useEffect(() => {
    setValue(null);
    setError(null);
    setMoreError(null);
    setLoading(false);
    setRefreshing(false);
    setLoadingMore(false);
    return () => {
      generation.current++;
      active.current?.abort();
      active.current = null;
    };
  }, [identity]);
  const load = useCallback(
    async (mode: "initial" | "refresh" | "more" = "initial") => {
      if (!identity || active.current) return;
      const previous = current.current?.identity === identity ? current.current.page : null;
      if (mode === "more" && !previous?.nextCursor) return;
      const abort = new AbortController();
      active.current = abort;
      const version = ++generation.current;
      if (mode === "initial") setLoading(true);
      if (mode === "refresh") setRefreshing(true);
      if (mode === "more") {
        setLoadingMore(true);
        setMoreError(null);
      }
      try {
        let next = await fetchPage(
          mode === "more" ? previous?.nextCursor : undefined,
          abort.signal,
        );
        if (version !== generation.current) return;
        if (mode === "refresh" && previous) {
          const cursors = new Set<string>();
          while (next.nextCursor && next.items.length < previous.items.length) {
            const cursor = next.nextCursor;
            if (cursors.has(cursor)) throw new Error("Repeated collection cursor");
            cursors.add(cursor);
            const continuation = await fetchPage(cursor, abort.signal);
            if (version !== generation.current) return;
            next = {
              ...continuation,
              items: [
                ...new Map(
                  [...next.items, ...continuation.items].map((item) => [itemKey(item), item]),
                ).values(),
              ],
            };
          }
        }
        const page =
          mode === "more" && previous
            ? {
                ...next,
                items: [
                  ...new Map(
                    [...previous.items, ...next.items].map((item) => [itemKey(item), item]),
                  ).values(),
                ],
              }
            : next;
        setValue({ identity, page: page as Value });
        setError(null);
        setMoreError(null);
      } catch (err) {
        if (
          abort.signal.aborted ||
          version !== generation.current ||
          err instanceof APISessionInvalidatedError
        )
          return;
        const failure = getRequestFailure(err);
        if (mode === "more") setMoreError(failure.messageKey);
        else if (previous) show({ kind: failure.kind, message: t(failure.messageKey) });
        else setError(failure.messageKey);
      } finally {
        if (version === generation.current) {
          active.current = null;
          setLoading(false);
          setRefreshing(false);
          setLoadingMore(false);
        }
      }
    },
    [fetchPage, identity, itemKey, show, t],
  );
  useFocusEffect(
    useCallback(() => {
      void load(current.current?.identity === identity ? "refresh" : "initial");
    }, [identity, load]),
  );
  return {
    page: value?.identity === identity ? value.page : null,
    error,
    moreError,
    loading,
    refreshing,
    loadingMore,
    load,
  };
}
