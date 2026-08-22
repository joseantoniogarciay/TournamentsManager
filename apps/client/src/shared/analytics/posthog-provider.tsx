import { usePathname } from "expo-router";
import { type PropsWithChildren, useEffect, useMemo, useState } from "react";
import { PostHog, PostHogProvider } from "posthog-react-native";

import { usePreferences } from "@/shared/preferences/preferences-provider";

import {
  activateProductAnalytics,
  captureProductAnalyticsEvent,
  deactivateProductAnalytics,
} from "./posthog-client";

const posthogAPIKey = process.env.EXPO_PUBLIC_POSTHOG_API_KEY;
const posthogEUHost = "https://eu.i.posthog.com";
const isPostHogBetaEnvironment = process.env.APP_ENV === "development";
const sensitiveExceptionPattern =
  /\b(?:bearer\s+[a-z0-9._-]+|(?:password|token|secret|authorization)\s*[:=]|[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,})/i;

/**
 * Separates essential reliability capture from optional product analytics.
 * The PostHog SDK starts only in the public beta, while the facade that emits
 * product events remains unavailable until the person explicitly enables it.
 */
export function ClientTelemetryProvider({ children }: PropsWithChildren) {
  if (!isPostHogBetaEnvironment || !posthogAPIKey) return children;

  return (
    <EnabledClientTelemetryProvider apiKey={posthogAPIKey}>
      {children}
    </EnabledClientTelemetryProvider>
  );
}

function EnabledClientTelemetryProvider({
  apiKey,
  children,
}: PropsWithChildren<{ apiKey: string }>) {
  const { productAnalyticsEnabled } = usePreferences();
  const [isProductAnalyticsReady, setIsProductAnalyticsReady] = useState(false);
  const client = useMemo(
    () =>
      new PostHog(apiKey, {
        defaultOptIn: true,
        host: posthogEUHost,
        captureAppLifecycleEvents: false,
        disableGeoip: true,
        disableRemoteFeatureFlags: true,
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
          if (!event) return null;
          const exceptions = JSON.stringify(event.properties?.["$exception_list"] ?? "");
          return sensitiveExceptionPattern.test(exceptions) ? null : event;
        },
      }),
    [],
  );

  useEffect(() => {
    if (!productAnalyticsEnabled) {
      deactivateProductAnalytics(client);
      setIsProductAnalyticsReady(false);
      return;
    }

    activateProductAnalytics(client);
    setIsProductAnalyticsReady(true);
    return () => {
      deactivateProductAnalytics(client);
      setIsProductAnalyticsReady(false);
    };
  }, [client, productAnalyticsEnabled]);

  useEffect(
    () => () => {
      void client.shutdown();
    },
    [client],
  );

  return (
    <PostHogProvider autocapture={false} client={client}>
      {isProductAnalyticsReady ? <ProductAnalyticsRouteObserver /> : null}
      {children}
    </PostHogProvider>
  );
}

function ProductAnalyticsRouteObserver() {
  const pathname = usePathname();

  useEffect(() => {
    captureProductAnalyticsEvent("screen_viewed", { screen: screenName(pathname) });
  }, [pathname]);

  return null;
}

function screenName(pathname: string) {
  if (pathname === "/") return "home";
  if (pathname === "/tournaments") return "tournaments";
  if (pathname === "/account") return "account";
  if (pathname === "/account/access") return "account_access";
  if (pathname === "/account/register") return "account_register";
  if (pathname === "/account/forgot-password") return "account_forgot_password";
  if (pathname === "/account/password") return "account_password";
  if (pathname === "/account/google-link") return "account_google_link";
  if (pathname === "/account/notifications") return "account_notifications";
  if (pathname === "/account/settings") return "account_settings";
  if (pathname === "/account-authentication") return "account_authentication";
  if (pathname === "/create-tournament") return "create_tournament";
  if (pathname === "/privacy-policy") return "privacy_policy";
  if (pathname === "/link/confirm") return "registration_confirmation";
  if (pathname === "/link/password-reset") return "password_reset_confirmation";
  if (pathname.endsWith("/administrators/add")) return "league_administrators_add";
  if (pathname.endsWith("/administrators")) return "league_administrators";
  if (pathname.endsWith("/teams")) return "league_teams";
  if (pathname.endsWith("/standings")) return "league_standings";
  if (pathname.endsWith("/transfer")) return "league_transfer";
  if (pathname.startsWith("/league/")) return "league_detail";
  return "unknown";
}
