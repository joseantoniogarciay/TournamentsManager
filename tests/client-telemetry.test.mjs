import assert from "node:assert/strict";
import test from "node:test";
import { mkdtemp, mkdir, readFile, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import {
  isReliabilityEventSafe,
  telemetryConfig,
} from "../apps/client/src/shared/analytics/telemetry-config.ts";
import { prepareWebTelemetry } from "../infra/home/prepare-web-telemetry.mjs";

const production = "phc_ProductionTestKey";

test("production permits safe exceptions and rejects product events or sensitive messages", () => {
  assert.equal(
    isReliabilityEventSafe({
      event: "$exception",
      properties: { $exception_list: [{ value: "Unexpected UI state" }] },
    }),
    true,
  );
  for (const event of ["screen_viewed", "$screen", "account_signed_in", "$identify"])
    assert.equal(isReliabilityEventSafe({ event }), false);
  for (const value of [
    "token=secret",
    "password:secret",
    "Bearer sample.token",
    "person@example.test",
  ])
    assert.equal(
      isReliabilityEventSafe({ event: "$exception", properties: { $exception_list: [{ value }] } }),
      false,
    );
});

test("the only project is reserved for production reliability", () => {
  assert.deepEqual(telemetryConfig("production", production), {
    apiKey: production,
    environment: "production",
    productAnalyticsAllowed: false,
  });
});

test("beta and local never start PostHog even with a valid key", () => {
  for (const environment of ["development", "local", "staging", undefined])
    assert.equal(telemetryConfig(environment, production), null);
});

test("missing keys and placeholders cannot start production telemetry", () => {
  for (const key of [
    undefined,
    "",
    "REPLACE_WITH_KEY",
    "phx_personal",
    "phc_<KEY>",
    "phc_REPLACE_WITH_KEY",
  ])
    assert.equal(telemetryConfig("production", key), null);
});

async function webFixture(context) {
  const root = await mkdtemp(join(tmpdir(), "tm-posthog-"));
  context.after(() => rm(root, { recursive: true, force: true }));
  const web = join(root, "_expo/static/js/web");
  await mkdir(web, { recursive: true });
  await writeFile(
    join(web, "entry.js"),
    "console.log('bundle');\n//# sourceMappingURL=entry.js.map",
  );
  await writeFile(join(web, "entry.js.map"), "{}");
  return { root, web };
}

test("web maps are private even when telemetry is not configured", async (context) => {
  const { root, web } = await webFixture(context);
  await prepareWebTelemetry("production", root, "a".repeat(40), {
    env: {},
    run: () => assert.fail("disabled telemetry must not upload"),
  });
  await assert.rejects(readFile(join(web, "entry.js.map")), { code: "ENOENT" });
  assert.doesNotMatch(await readFile(join(web, "entry.js"), "utf8"), /sourceMappingURL/);
});

test("web upload uses explicit project credentials and stops publication on failure", async (context) => {
  const { root, web } = await webFixture(context);
  const file = join(root, "posthog-production.env");
  await writeFile(
    file,
    "POSTHOG_CLI_HOST=https://eu.posthog.com\nPOSTHOG_CLI_PROJECT_ID=255144\nPOSTHOG_CLI_API_KEY=phx_TestOnlyKey\n",
  );
  const env = {
    EXPO_PUBLIC_POSTHOG_PRODUCTION_API_KEY: production,
    FASTTOURNEY_PROD_POSTHOG_SYMBOLS_CONFIG: file,
    POSTHOG_CLI_API_KEY: "wrong-inherited-project",
    POSTHOG_CLI_PROJECT_ID: "999",
  };
  const run = (command, args, options) => {
    assert.equal(command, "posthog-cli");
    assert.equal(args[args.indexOf("--dotenv-file") + 1], file);
    assert.equal(args[args.indexOf("--host") + 1], "https://eu.posthog.com");
    assert.equal(args[args.indexOf("--release-name") + 1], "fasttourney-web-production");
    assert.equal(options.env.POSTHOG_CLI_API_KEY, undefined);
    assert.equal(options.env.POSTHOG_CLI_PROJECT_ID, undefined);
    return { status: 1 };
  };
  await assert.rejects(prepareWebTelemetry("production", root, "a".repeat(40), { env, run }));
  assert.equal(await readFile(join(web, "entry.js.map"), "utf8"), "{}");
  await prepareWebTelemetry("production", root, "a".repeat(40), {
    env,
    run: (...args) => {
      run(...args);
      return { status: 0 };
    },
  });
  await assert.rejects(readFile(join(web, "entry.js.map")), { code: "ENOENT" });
});

test("incomplete CLI credentials cannot fall back to a logged-in project", async (context) => {
  const { root } = await webFixture(context);
  const file = join(root, "placeholder.env");
  await writeFile(file, "POSTHOG_CLI_PROJECT_ID=REPLACE_WITH_ID\nPOSTHOG_CLI_API_KEY=\n");
  await assert.rejects(
    prepareWebTelemetry("production", root, "a".repeat(40), {
      env: {
        EXPO_PUBLIC_POSTHOG_PRODUCTION_API_KEY: production,
        FASTTOURNEY_PROD_POSTHOG_SYMBOLS_CONFIG: file,
      },
      run: () => assert.fail("incomplete credentials must not launch CLI"),
    }),
  );
});

test("a different or duplicated project cannot receive production symbols", async (context) => {
  const { root } = await webFixture(context);
  const file = join(root, "wrong-project.env");
  for (const project of ["123", "255144\nPOSTHOG_CLI_PROJECT_ID=123"]) {
    await writeFile(
      file,
      `POSTHOG_CLI_HOST=https://eu.posthog.com\nPOSTHOG_CLI_PROJECT_ID=${project}\nPOSTHOG_CLI_API_KEY=phx_TestOnlyKey\n`,
    );
    await assert.rejects(
      prepareWebTelemetry("production", root, "a".repeat(40), {
        env: {
          EXPO_PUBLIC_POSTHOG_PRODUCTION_API_KEY: production,
          FASTTOURNEY_PROD_POSTHOG_SYMBOLS_CONFIG: file,
        },
        run: () => assert.fail("wrong project must not launch CLI"),
      }),
    );
  }
});
