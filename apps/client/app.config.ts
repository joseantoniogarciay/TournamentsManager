import type { ConfigContext, ExpoConfig } from "expo/config";

type AppEnvironment = "development" | "production";

const appEnvironment: AppEnvironment =
  process.env.APP_ENV === "production" ? "production" : "development";

const environmentConfig = {
  development: {
    bundleIdentifier: "com.fasttourney.app.dev",
    name: "Fast Tourney Dev",
    scheme: "fasttourney-dev",
    appLinkURL: "https://dev.fasttourney.com",
  },
  production: {
    bundleIdentifier: "com.fasttourney.app",
    name: "Fast Tourney",
    scheme: "fasttourney",
    appLinkURL: "https://fasttourney.com",
  },
} as const;

const appLinkDomain = getAppLinkDomain(
  process.env.EXPO_PUBLIC_APP_LINK_URL || environmentConfig[appEnvironment].appLinkURL,
);
const googleRedirectSchemes = getGoogleRedirectSchemes();

function getAppLinkDomain(appLinkURL: string) {
  const parsedURL = new URL(appLinkURL);
  if (parsedURL.protocol !== "https:" || parsedURL.pathname !== "/" || parsedURL.search) {
    throw new Error("EXPO_PUBLIC_APP_LINK_URL debe ser un origen HTTPS sin path ni query");
  }
  return parsedURL.hostname;
}

function getGoogleRedirectSchemes() {
  const clientIDs = [
    process.env.EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID,
    process.env.EXPO_PUBLIC_GOOGLE_ANDROID_CLIENT_ID,
  ];
  return [
    ...new Set(
      clientIDs
        .filter((clientID): clientID is string => Boolean(clientID))
        .map(
          (clientID) =>
            `com.googleusercontent.apps.${clientID.replace(/\.apps\.googleusercontent\.com$/, "")}`,
        ),
    ),
  ];
}

export default ({ config }: ConfigContext): ExpoConfig => ({
  ...config,
  name: environmentConfig[appEnvironment].name,
  slug: "fast-tourney",
  scheme: [environmentConfig[appEnvironment].scheme, ...googleRedirectSchemes],
  version: "1.0.0",
  orientation: "portrait",
  web: {
    description: "Crea, organiza y sigue ligas de fútbol con FastTourney.",
    lang: "es",
  },
  icon: "./assets/fast-tourney-icon.png",
  ios: {
    bundleIdentifier: environmentConfig[appEnvironment].bundleIdentifier,
    associatedDomains: [`applinks:${appLinkDomain}`, `webcredentials:${appLinkDomain}`],
  },
  android: {
    package: environmentConfig[appEnvironment].bundleIdentifier,
    intentFilters: [
      {
        action: "VIEW",
        autoVerify: true,
        category: ["BROWSABLE", "DEFAULT"],
        data: [{ scheme: "https", host: appLinkDomain, pathPrefix: "/link/" }],
      },
    ],
  },
  userInterfaceStyle: "automatic",
  plugins: [
    "expo-router",
    "expo-secure-store",
    "expo-web-browser",
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
