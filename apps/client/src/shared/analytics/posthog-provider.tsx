import Constants from "expo-constants";
import { type PropsWithChildren, useEffect, useMemo } from "react";
import { PostHog, PostHogProvider } from "posthog-react-native";

import { isReliabilityEventSafe, telemetryConfig } from "./telemetry-config";

const configuration = telemetryConfig(
  Constants.expoConfig?.extra?.appEnvironment,
  process.env.EXPO_PUBLIC_POSTHOG_PRODUCTION_API_KEY,
);
const posthogEUHost = "https://eu.i.posthog.com";

/**
 * Separates essential reliability capture from optional product analytics.
 * The single EU project is reserved for production reliability.
 * Beta/local do not initialize PostHog or activate product analytics.
 */
export function ClientTelemetryProvider({ children }: PropsWithChildren) {
  if (!configuration) return children;

  return (
    <EnabledClientTelemetryProvider configuration={configuration}>
      {children}
    </EnabledClientTelemetryProvider>
  );
}

function EnabledClientTelemetryProvider({
  configuration,
  children,
}: PropsWithChildren<{ configuration: NonNullable<ReturnType<typeof telemetryConfig>> }>) {
  const client = useMemo(
    () =>
      new PostHog(configuration.apiKey, {
        defaultOptIn: true,
        host: posthogEUHost,
        captureAppLifecycleEvents: false,
        disableGeoip: true,
        disableRemoteFeatureFlags: true,
        disableSurveys: true,
        capturePushNotificationOpened: false,
        capturePushNotificationSubscriptions: false,
        enableSessionReplay: false,
        errorTracking: {
          autocapture: {
            console: false,
            nativeCrashes: true,
            uncaughtExceptions: true,
            unhandledRejections: true,
          },
          exceptionSteps: { enabled: false },
        },
        before_send: (event) => {
          if (!event || !isReliabilityEventSafe(event)) return null;
          event.properties = { ...event.properties, environment: configuration.environment };
          return event;
        },
      }),
    [configuration],
  );

  useEffect(
    () => () => {
      void client.shutdown();
    },
    [client],
  );

  return (
    <PostHogProvider autocapture={false} client={client}>
      {children}
    </PostHogProvider>
  );
}
