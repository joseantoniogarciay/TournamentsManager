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

export default ({ config }: ConfigContext): ExpoConfig => ({
  ...config,
  name: environmentConfig[appEnvironment].name,
  slug: "fast-tourney",
  scheme: environmentConfig[appEnvironment].scheme,
  version: "1.0.0",
  orientation: "portrait",
  icon: "./assets/fast-tourney-icon.png",
  ios: { bundleIdentifier: environmentConfig[appEnvironment].bundleIdentifier },
  userInterfaceStyle: "automatic",
  plugins: ["expo-router"],
  experiments: { typedRoutes: true },
  extra: { appEnvironment },
});
