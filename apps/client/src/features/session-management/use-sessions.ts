import { useFocusEffect } from "expo-router";
import { useCallback, useRef, useState } from "react";
import type { ActiveSessionPage } from "@/api/generated/models";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import { getTranslator } from "@/shared/i18n/locale";
import { listSessions, revokeOthers, revokeSession } from "./api";

export function useSessions(accountID: string | undefined) {
  const t = getTranslator();
  const { show } = useFeedback();
  const [page, setPage] = useState<ActiveSessionPage>();
  const [loadError, setLoadError] = useState<string>();
  const [busy, setBusy] = useState(false);
  const [pending, setPending] = useState<string>();
  const controller = useRef<AbortController | null>(null);
  const inFlight = useRef(false);
  const pageRef = useRef<ActiveSessionPage | undefined>(undefined);
  const report = useCallback(
    (error: unknown) => {
      const failure = getRequestFailure(error);
      show({ kind: failure.kind, message: t(failure.messageKey) });
    },
    [show, t],
  );

  const load = useCallback(
    async (cursor?: string) => {
      if (!accountID || inFlight.current) return;
      inFlight.current = true;
      const request = new AbortController();
      controller.current = request;
      setBusy(true);
      setLoadError(undefined);
      try {
        const next = await listSessions(cursor, request.signal);
        if (request.signal.aborted) return;
        setPage((previous) => {
          const updated =
            cursor && previous
              ? {
                  ...next,
                  items: [
                    ...previous.items,
                    ...next.items.filter(
                      (item) => !previous.items.some((old) => old.id === item.id),
                    ),
                  ],
                }
              : next;
          pageRef.current = updated;
          return updated;
        });
      } catch (error) {
        if (request.signal.aborted) return;
        if (pageRef.current) report(error);
        else setLoadError(t(getRequestFailure(error).messageKey));
      } finally {
        if (controller.current === request) {
          inFlight.current = false;
          setBusy(false);
        }
      }
    },
    [accountID, report, t],
  );

  useFocusEffect(
    useCallback(() => {
      pageRef.current = undefined;
      setPage(undefined);
      setLoadError(undefined);
      setPending(undefined);
      setBusy(false);
      void load();
      return () => {
        controller.current?.abort();
        controller.current = null;
        inFlight.current = false;
      };
    }, [load]),
  );

  const revoke = async (id?: string) => {
    if (!accountID || inFlight.current) return;
    inFlight.current = true;
    const request = new AbortController();
    controller.current = request;
    setBusy(true);
    setPending(id ?? "others");
    try {
      if (id) await revokeSession(id, request.signal);
      else await revokeOthers(request.signal);
      if (request.signal.aborted) return;
      if (id) {
        setPage((previous) => {
          const next = previous && {
            ...previous,
            items: previous.items.filter((item) => item.id !== id),
          };
          pageRef.current = next;
          return next;
        });
      } else {
        pageRef.current = undefined;
        setPage(undefined);
        const next = await listSessions(undefined, request.signal);
        if (!request.signal.aborted) {
          pageRef.current = next;
          setPage(next);
        }
      }
    } catch (error) {
      if (request.signal.aborted) return;
      if (pageRef.current) report(error);
      else setLoadError(t(getRequestFailure(error).messageKey));
    } finally {
      if (controller.current === request) {
        inFlight.current = false;
        setBusy(false);
        setPending(undefined);
      }
    }
  };
  return { page, loadError, busy, pending, load, revoke };
}
