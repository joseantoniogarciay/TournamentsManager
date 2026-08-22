import type { PostHog } from "posthog-react-native";

type ProductAnalyticsProperties = Parameters<PostHog["capture"]>[1];

let activeClient: PostHog | null = null;

export function activateProductAnalytics(client: PostHog) {
  activeClient = client;
}

export function deactivateProductAnalytics(client: PostHog) {
  if (activeClient === client) activeClient = null;
}

export function isProductAnalyticsActive() {
  return activeClient !== null;
}

export function captureProductAnalyticsEvent(
  event: string,
  properties?: ProductAnalyticsProperties,
) {
  activeClient?.capture(event, properties);
}
