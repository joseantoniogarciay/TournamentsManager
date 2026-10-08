function publicKey(value: string | undefined) {
  const key = value?.trim();
  return key &&
    !/placeholder|replace[_-]|change-me|[<>]/i.test(key) &&
    /^phc_[A-Za-z0-9_-]+$/.test(key)
    ? key
    : undefined;
}

const sensitiveExceptionPattern =
  /\b(?:bearer\s+[a-z0-9._-]+|(?:password|token|secret|authorization)\s*[:=]|[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,})/i;

export function isReliabilityEventSafe(event: {
  event: string;
  properties?: Record<string, unknown>;
}) {
  return (
    event.event === "$exception" &&
    !sensitiveExceptionPattern.test(JSON.stringify(event.properties?.["$exception_list"] ?? ""))
  );
}

export function telemetryConfig(
  environment: string | undefined,
  productionKey: string | undefined,
) {
  const apiKey = publicKey(productionKey);
  return environment === "production" && apiKey
    ? ({ environment, apiKey, productAnalyticsAllowed: false } as const)
    : null;
}
