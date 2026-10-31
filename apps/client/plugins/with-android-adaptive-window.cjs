"use strict";

const { AndroidConfig, withAndroidManifest } = require("expo/config-plugins");

function configureAdaptiveWindow(manifest) {
  const activity = AndroidConfig.Manifest.getMainActivityOrThrow(manifest);
  // All four rotations when enabled by the person; preserve their rotation lock.
  activity.$["android:screenOrientation"] = "fullUser";
  activity.$["android:resizeableActivity"] = "true";
  return manifest;
}

module.exports = (config) =>
  withAndroidManifest(config, (mod) => {
    mod.modResults = configureAdaptiveWindow(mod.modResults);
    return mod;
  });
module.exports.configureAdaptiveWindow = configureAdaptiveWindow;
