/* global require, module, __dirname, process */
// eslint-disable-next-line @typescript-eslint/no-require-imports -- Metro loads this CommonJS configuration synchronously.
const { getPostHogExpoConfig } = require("posthog-react-native/metro");

module.exports = getPostHogExpoConfig(__dirname, {
  enabled: process.env.APP_ENV === "production",
});
