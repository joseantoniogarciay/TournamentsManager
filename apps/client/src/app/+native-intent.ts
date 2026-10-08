import { isNativeSocialAuthReturn } from "@/features/social-authentication/config";
import {
  clearPendingTeamInvitation,
  invitationTokenFromURL,
  rememberPendingTeamInvitation,
} from "@/features/league-creation/team-invitation";
import { deferInitialDeepLink, toInternalPath } from "@/shared/navigation/deep-link-gate";

let invitationRevision = 0;

/** Captura el fragmento antes de navegar, tanto al arrancar como con la app abierta. */

export async function redirectSystemPath({
  path,
  initial,
}: {
  path: string | null;
  initial: boolean;
}) {
  if (!path) return path;

  // AuthSession owns warm OAuth returns; the router must not unmount the
  // initiating modal and lose its proof/draft while the browser is closing.
  if (isNativeSocialAuthReturn(path)) {
    return initial ? "/account" : null;
  }

  let internalPath = toInternalPath(path);
  const routePath = internalPath.split(/[?#]/, 1)[0];
  if (
    routePath === "/join-team" ||
    routePath === "/--/join-team" ||
    /^\/[^/]+\/--\/join-team$/.test(routePath)
  ) {
    const incoming = invitationTokenFromURL(internalPath);
    if (incoming.present) {
      try {
        if (incoming.token) await rememberPendingTeamInvitation(incoming.token);
        else await clearPendingTeamInvitation();
        invitationRevision += 1;
        internalPath = `/join-team?invitationRevision=${invitationRevision}`;
      } catch {
        return "/";
      }
    }
  }
  if (!initial) return internalPath;

  return deferInitialDeepLink(internalPath) ? "/" : internalPath;
}
