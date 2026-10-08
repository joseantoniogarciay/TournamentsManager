import assert from "node:assert/strict";
import { mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";
import { URL } from "node:url";
import { prepareAssociations, validateAssociations } from "../infra/app-links/prepare.mjs";

function fixture(environment) {
  const directory = new URL(`../infra/app-links/${environment}/`, import.meta.url);
  // Synthetic values only test structure; they are never deployment inputs.
  const apple = JSON.parse(
    readFileSync(new URL("apple-app-site-association.template", directory), "utf8").replaceAll(
      "APPLE_TEAM_ID",
      "TEST123456",
    ),
  );
  const android = JSON.parse(
    readFileSync(new URL("assetlinks.json.template", directory), "utf8").replace(
      /ANDROID_(DEVELOPMENT|RELEASE)_CERT_SHA256/,
      Array(32).fill("AB").join(":"),
    ),
  );
  return { apple, android };
}

for (const environment of ["development", "production"]) {
  test(`${environment}: covers routes and rejects foreign or local app identities`, () => {
    const { apple, android } = fixture(environment);
    validateAssociations(environment, apple, android);
    const foreign = fixture(environment === "development" ? "production" : "development");
    assert.throws(() => validateAssociations(environment, foreign.apple, android));
    assert.throws(() => validateAssociations(environment, apple, foreign.android));
    android[0].target.package_name = "com.fasttourney.app.local";
    assert.throws(() => validateAssociations(environment, apple, android));
  });
}

test("rejects partial, conditional, and extra Apple associations", () => {
  for (const mutate of [
    (apple) => apple.applinks.details[0].components.pop(),
    (apple) => {
      apple.applinks.details[0].appID = "TEST123456.com.fasttourney.app.local";
    },
    (apple) => {
      apple.activitycontinuation = { apps: ["TEST123456.com.fasttourney.app.local"] };
    },
    (apple) => {
      apple.applinks.details[0].components[0].exclude = true;
    },
    (apple) => apple.applinks.details[0].appIDs.push("TEST123456.com.fasttourney.app.local"),
    (apple) => apple.webcredentials.apps.push("TEST123456.com.fasttourney.app.dev"),
  ]) {
    const { apple, android } = fixture("production");
    mutate(apple);
    assert.throws(() => validateAssociations("production", apple, android));
  }
});

test("rejects every malformed fingerprint and additional Android statements", () => {
  for (const value of [
    "ANDROID_RELEASE_CERT_SHA256",
    ":".repeat(95),
    "AB".repeat(32),
    "AB:".repeat(32),
    Array(32).fill("ab").join(":"),
  ]) {
    const { apple, android } = fixture("production");
    android[0].target.sha256_cert_fingerprints.push(value);
    assert.throws(() => validateAssociations("production", apple, android));
  }
  const { apple, android } = fixture("production");
  android.push(JSON.parse(JSON.stringify(android[0])));
  assert.throws(() => validateAssociations("production", apple, android));
});

test("copies only a complete validated pair and supports a required mobile gate", () => {
  const source = mkdtempSync(join(tmpdir(), "app-links-source-"));
  const destination = mkdtempSync(join(tmpdir(), "app-links-release-"));
  try {
    assert.equal(prepareAssociations("development", source, destination), false);
    assert.throws(() => prepareAssociations("development", source, destination, true));
    const { apple, android } = fixture("development");
    writeFileSync(join(source, "apple-app-site-association"), JSON.stringify(apple));
    assert.throws(() => prepareAssociations("development", source, destination));
    writeFileSync(join(source, "assetlinks.json"), "<html>SPA</html>");
    assert.throws(() => prepareAssociations("development", source, destination));
    writeFileSync(join(source, "assetlinks.json"), JSON.stringify(android));
    assert.equal(prepareAssociations("development", source, destination, true), true);
    for (const name of ["apple-app-site-association", "assetlinks.json"]) {
      assert.equal(
        readFileSync(join(destination, ".well-known", name), "utf8"),
        readFileSync(join(source, name), "utf8"),
      );
    }
    assert.throws(() => prepareAssociations("local", source, destination));
  } finally {
    rmSync(source, { recursive: true, force: true });
    rmSync(destination, { recursive: true, force: true });
  }
});
