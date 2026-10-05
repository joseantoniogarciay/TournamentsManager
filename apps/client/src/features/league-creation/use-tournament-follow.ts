import { useEffect, useRef, useState } from "react";

import { APISessionInvalidatedError } from "@/api/fetch";
import { useFeedback } from "@/shared/feedback/feedback-provider";
import { getRequestFailure } from "@/shared/feedback/request-failure";
import { getTranslator } from "@/shared/i18n/locale";
import { useSession } from "@/shared/session/session-provider";
import {
  followTournamentRequest,
  unfollowTournamentRequest,
  TournamentUnavailableError,
} from "./api";
import { invalidateTournamentLibrary } from "./tournament-library-revision";

export function useTournamentFollow(
  id: string,
  relationship: string | null | undefined,
  onRelationshipChange: (value: string | null) => void,
) {
  const { user } = useSession();
  const { show } = useFeedback();
  const t = getTranslator();
  const [isSaving, setIsSaving] = useState(false);
  const inFlight = useRef(false);
  const generation = useRef(0);
  useEffect(() => {
    setIsSaving(false);
    return () => {
      generation.current += 1;
      inFlight.current = false;
    };
  }, [id, user?.id]);
  const canFollow = Boolean(user) && (relationship === null || relationship === "follower");
  const toggleFollow = async () => {
    if (!id || !canFollow || inFlight.current) return;
    inFlight.current = true;
    setIsSaving(true);
    const requestGeneration = generation.current;
    const wasFollowed = relationship === "follower";
    try {
      await (wasFollowed ? unfollowTournamentRequest(id) : followTournamentRequest(id));
      // Incluso si la ruta se desmontó, el cambio confirmado invalida la biblioteca.
      invalidateTournamentLibrary();
      if (requestGeneration !== generation.current) return;
      onRelationshipChange(wasFollowed ? null : "follower");
    } catch (error) {
      if (requestGeneration !== generation.current || error instanceof APISessionInvalidatedError)
        return;
      const failure = getRequestFailure(error);
      show({
        kind: failure.kind,
        message: t(
          error instanceof TournamentUnavailableError ? "league_unavailable" : failure.messageKey,
        ),
      });
    } finally {
      if (requestGeneration === generation.current) {
        inFlight.current = false;
        setIsSaving(false);
      }
    }
  };
  return { canFollow, isFollowed: relationship === "follower", isSaving, toggleFollow };
}
