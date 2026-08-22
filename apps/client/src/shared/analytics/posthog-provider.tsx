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

/**
 * Keeps product telemetry outside the application domain and creates its SDK
 * only after the person explicitly enables optional analytics.
 */
export function ProductAnalyticsProvider({ children }: PropsWithChildren) {
  const { productAnalyticsEnabled } = usePreferences();

  if (!isPostHogBetaEnvironment || !productAnalyticsEnabled || !posthogAPIKey) return children;

  return (
    <EnabledProductAnalyticsProvider apiKey={posthogAPIKey}>
      {children}
    </EnabledProductAnalyticsProvider>
  );
}

function EnabledProductAnalyticsProvider({
  apiKey,
  children,
}: PropsWithChildren<{ apiKey: string }>) {
  const [isReady, setIsReady] = useState(false);
  const client = useMemo(
    () =>
      new PostHog(apiKey, {
        defaultOptIn: false,
        host: posthogEUHost,
        captureAppLifecycleEvents: false,
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
      }),
    [],
  );

  useEffect(() => {
    let isMounted = true;
    void client.optIn().then(() => {
      if (!isMounted) return;
      activateProductAnalytics(client);
      setIsReady(true);
    });
    return () => {
      isMounted = false;
      deactivateProductAnalytics(client);
      void client.optOut().finally(() => client.shutdown());
    };
  }, [client]);

  return (
    <PostHogProvider autocapture={false} client={client}>
      {isReady ? <ProductAnalyticsRouteObserver /> : null}
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
