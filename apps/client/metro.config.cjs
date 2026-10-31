"use strict";

const { getPostHogExpoConfig } = require("posthog-react-native/metro");

module.exports = getPostHogExpoConfig(__dirname, {
  enabled: process.env.APP_ENV === "production",
});
