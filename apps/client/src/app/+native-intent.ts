import { isNativeSocialAuthReturn } from "@/features/social-authentication/config";
import { deferInitialDeepLink, toInternalPath } from "@/shared/navigation/deep-link-gate";

/**
 * Expo Router lo invoca antes de montar React cuando la app nativa nace desde
 * un enlace. Las entregas a una app ya viva conservan el comportamiento normal.
 */
export function redirectSystemPath({ path, initial }: { path: string | null; initial: boolean }) {
  if (!path) return path;

  // AuthSession owns warm OAuth returns; the router must not unmount the
  // initiating modal and lose its proof/draft while the browser is closing.
  if (isNativeSocialAuthReturn(path)) {
    return initial ? "/account" : null;
  }

  const internalPath = toInternalPath(path);
  if (!initial) return internalPath;

  return deferInitialDeepLink(internalPath) ? "/" : internalPath;
}
