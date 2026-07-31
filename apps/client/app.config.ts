import type { ConfigContext, ExpoConfig } from "expo/config";

type AppEnvironment = "development" | "production";

const appEnvironment: AppEnvironment =
  process.env.APP_ENV === "production" ? "production" : "development";

const environmentConfig = {
  development: {
    bundleIdentifier: "com.fasttourney.app.dev",
    name: "Fast Tourney Dev",
    scheme: "fasttourney-dev",
  },
  production: {
    bundleIdentifier: "com.fasttourney.app",
    name: "Fast Tourney",
    scheme: "fasttourney",
  },
} as const;

const appLinkDomain = getAppLinkDomain(process.env.EXPO_PUBLIC_APP_LINK_URL);

function getAppLinkDomain(appLinkURL: string | undefined) {
  if (!appLinkURL) return undefined;
  const parsedURL = new URL(appLinkURL);
  if (parsedURL.protocol !== "https:" || parsedURL.pathname !== "/" || parsedURL.search) {
    throw new Error("EXPO_PUBLIC_APP_LINK_URL debe ser un origen HTTPS sin path ni query");
  }
  return parsedURL.hostname;
}

export default ({ config }: ConfigContext): ExpoConfig => ({
  ...config,
  name: environmentConfig[appEnvironment].name,
  slug: "fast-tourney",
  scheme: environmentConfig[appEnvironment].scheme,
  version: "1.0.0",
  orientation: "portrait",
  icon: "./assets/fast-tourney-icon.png",
  ios: {
    bundleIdentifier: environmentConfig[appEnvironment].bundleIdentifier,
    associatedDomains: appLinkDomain ? [`applinks:${appLinkDomain}`] : undefined,
  },
  android: {
    package: environmentConfig[appEnvironment].bundleIdentifier,
    intentFilters: appLinkDomain
      ? [
          {
            action: "VIEW",
            autoVerify: true,
            category: ["BROWSABLE", "DEFAULT"],
            data: [{ scheme: "https", host: appLinkDomain, pathPrefix: "/link/" }],
          },
        ]
      : undefined,
  },
  userInterfaceStyle: "automatic",
  plugins: [
    "expo-router",
    "expo-secure-store",
    [
      "expo-splash-screen",
      {
        backgroundColor: "#F8FAFC",
        image: "./assets/fast-tourney-icon.png",
        imageWidth: 160,
        dark: {
          backgroundColor: "#101828",
          image: "./assets/fast-tourney-icon.png",
        },
      },
    ],
  ],
  experiments: { typedRoutes: true },
  extra: { appEnvironment },
});
