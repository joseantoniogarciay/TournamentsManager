/** Example values must never enable a provider or create a native URL scheme. */
export function configuredGoogleClientID(value: string | undefined) {
  return value &&
    !/placeholder|replace[_-]|change-me|[<>]/i.test(value) &&
    /^\d+-[A-Za-z0-9_-]+\.apps\.googleusercontent\.com$/.test(value)
    ? value
    : undefined;
}

export function configuredAppleServiceID(value: string | undefined) {
  if (!value || /placeholder|replace[_-]|change-me|[<>]/i.test(value)) return undefined;
  return /^[A-Za-z0-9][A-Za-z0-9.-]+$/.test(value) ? value : undefined;
}

/** OAuth return events belong to AuthSession, not a new router navigation. */
export function isNativeSocialAuthReturn(path: string) {
  try {
    const url = new URL(path);
    return (
      ((url.protocol === "fasttourney:" || url.protocol === "fasttourney-dev:") &&
        url.host === "oauth" &&
        url.pathname === "/apple-complete") ||
      (/^com\.googleusercontent\.apps\.\d+-[A-Za-z0-9_-]+:$/.test(url.protocol) &&
        url.pathname === "/oauthredirect")
    );
  } catch {
    return false;
  }
}
