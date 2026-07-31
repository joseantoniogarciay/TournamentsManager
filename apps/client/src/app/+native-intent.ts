import { deferInitialDeepLink } from "@/shared/navigation/deep-link-gate";

/**
 * Expo Router lo invoca antes de montar React cuando la app nativa nace desde
 * un enlace. Las entregas a una app ya viva conservan el comportamiento normal.
 */
export function redirectSystemPath({ path, initial }: { path: string | null; initial: boolean }) {
  if (!path || !initial) return path;

  return deferInitialDeepLink(path) ? "/" : path;
}
