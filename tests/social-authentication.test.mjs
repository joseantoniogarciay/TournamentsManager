import assert from "node:assert/strict";
import test from "node:test";
import { customTabsBrowserOptions } from "../apps/client/src/features/social-authentication/browser-options.ts";
import { appleReturnStatus } from "../apps/client/src/features/federated-apple/return.ts";
import {
  configuredAppleServiceID,
  configuredGoogleClientID,
  isNativeSocialAuthReturn,
} from "../apps/client/src/features/social-authentication/config.ts";

test("placeholders cannot enable Apple or Google", () => {
  for (const value of [
    undefined,
    "",
    "PLACEHOLDER",
    "REPLACE_WITH_SERVICE_ID",
    "<APPLE_SERVICE_ID>",
    "replace-with-production-google-web-client-id",
  ]) {
    assert.equal(configuredAppleServiceID(value), undefined);
    assert.equal(configuredGoogleClientID(value), undefined);
  }
  assert.equal(configuredAppleServiceID("com.fasttourney.web.dev"), "com.fasttourney.web.dev");
  assert.equal(
    configuredGoogleClientID("123-test.apps.googleusercontent.com"),
    "123-test.apps.googleusercontent.com",
  );
});

test("Android prioritizes Custom Tabs and permits the browser fallback", () => {
  const browsers = {
    browserPackages: ["full.browser", "tabs.browser"],
    servicePackages: ["tabs.browser"],
    preferredBrowserPackage: "full.browser",
    defaultBrowserPackage: "full.browser",
  };
  assert.deepEqual(customTabsBrowserOptions(browsers), { browserPackage: "tabs.browser" });
  assert.deepEqual(customTabsBrowserOptions({ ...browsers, servicePackages: [] }), {});
  assert.deepEqual(customTabsBrowserOptions({ ...browsers, browserPackages: [] }), {});
});

test("Apple return checks custom scheme, host, path and originating challenge", () => {
  const returnURL = "fasttourney-dev://oauth/apple-complete";
  assert.equal(
    appleReturnStatus(`${returnURL}?challengeId=attempt&status=ready`, returnURL, "attempt"),
    "ready",
  );
  assert.equal(
    appleReturnStatus(`${returnURL}?challengeId=attempt&status=cancelled`, returnURL, "attempt"),
    "cancelled",
  );
  for (const url of [
    "evil://oauth/apple-complete?challengeId=attempt&status=ready",
    "fasttourney-dev://evil/apple-complete?challengeId=attempt&status=ready",
    "fasttourney-dev://oauth/other?challengeId=attempt&status=ready",
    `${returnURL}?challengeId=other&status=ready`,
    `${returnURL}?challengeId=attempt&status=ready&status=cancelled`,
    `${returnURL}?challengeId=attempt&challengeId=other&status=ready`,
    `${returnURL}?challengeId=attempt&status=unknown`,
    `${returnURL}?challengeId=attempt&status=ready#token`,
  ])
    assert.throws(() => appleReturnStatus(url, returnURL, "attempt"));
});

test("web Apple return must remain on the initiating origin", () => {
  const returnURL = "https://dev.fasttourney.com/oauth/apple-complete";
  assert.equal(
    appleReturnStatus(`${returnURL}?challengeId=attempt&status=ready`, returnURL, "attempt"),
    "ready",
  );
  assert.throws(() =>
    appleReturnStatus(
      "https://fasttourney.com/oauth/apple-complete?challengeId=attempt&status=ready",
      returnURL,
      "attempt",
    ),
  );
});

test("warm OAuth returns do not replace the initiating route", () => {
  assert.equal(
    isNativeSocialAuthReturn("fasttourney-dev://oauth/apple-complete?challengeId=test"),
    true,
  );
  assert.equal(
    isNativeSocialAuthReturn("com.googleusercontent.apps.123-test:/oauthredirect?code=sample"),
    true,
  );
  assert.equal(isNativeSocialAuthReturn("https://evil.test/path://oauth/apple-complete"), false);
  assert.equal(isNativeSocialAuthReturn("fasttourney-dev://tournament/123"), false);
});

test("malformed successful responses never establish a social session", async () => {
  const { isSocialSession } =
    await import("../apps/client/src/features/social-authentication/session.ts");
  const valid = {
    delivery: "cookie",
    user: { id: "account", username: "person" },
    expiresAt: new Date(Date.now() + 60000).toISOString(),
    refreshExpiresAt: new Date(Date.now() + 120000).toISOString(),
  };
  assert.equal(isSocialSession(valid, "cookie"), true);
  assert.equal(isSocialSession(valid, "bearer"), false);
  for (const value of [
    null,
    {},
    { ...valid, user: {} },
    { ...valid, expiresAt: "invalid" },
    { ...valid, delivery: "other" },
  ])
    assert.equal(isSocialSession(value, "cookie"), false);
  assert.equal(
    isSocialSession(
      { ...valid, delivery: "bearer", sessionToken: "access", refreshToken: "refresh" },
      "bearer",
    ),
    true,
  );
});
