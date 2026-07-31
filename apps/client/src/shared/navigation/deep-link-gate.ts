let pendingInitialDeepLink: string | null = null;

/**
 * Retiene en memoria el enlace que abrió una app nativa terminada. El router
 * inicia en Inicio y lo consume después de que esa raíz se haya pintado.
 */
export function deferInitialDeepLink(path: string) {
  const internalPath = toInternalPath(path);
  if (internalPath === "/") return false;

  pendingInitialDeepLink = internalPath;
  return true;
}

export function consumeDeferredInitialDeepLink() {
  const path = pendingInitialDeepLink;
  pendingInitialDeepLink = null;
  return path;
}

function toInternalPath(path: string) {
  if (path.startsWith("/")) return path;

  try {
    const url = new URL(path);
    const prefix = url.protocol === "http:" || url.protocol === "https:" ? "" : `/${url.host}`;
    return `${prefix}${url.pathname}${url.search}${url.hash}`;
  } catch {
    return "/";
  }
}
