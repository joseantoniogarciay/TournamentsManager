import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import test from "node:test";
import { URL } from "node:url";
import { runInThisContext } from "node:vm";
import ts from "typescript";

const source = readFileSync(
  new URL("../apps/client/src/features/store-review/policy.ts", import.meta.url),
  "utf8",
);
const module = { exports: {} };
runInThisContext(
  `(function(module,exports){${
    ts.transpileModule(source, {
      compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 },
    }).outputText
  }\n})`,
)(module, module.exports);
const { createReviewController, reviewIntervalMs } = module.exports;
const safe = () => true;

function fixture(overrides = {}) {
  let stored = null;
  let now = 1_800_000_000_000;
  let requests = 0;
  const ports = {
    read: async () => stored,
    write: async (value) => {
      stored = value;
    },
    available: async () => true,
    request: async () => {
      requests++;
    },
    now: () => now,
    ...overrides,
  };
  return {
    controller: createReviewController(ports),
    restart: () => createReviewController(ports),
    advance: (ms) => {
      now += ms;
    },
    requests: () => requests,
    state: () => JSON.parse(stored),
  };
}

test("the tenth aggregate visit requests once, persisting across launches", async () => {
  const f = fixture();
  for (let i = 0; i < 9; i++) await f.controller.visit(safe);
  assert.equal(f.requests(), 0);
  await f.restart().visit(safe);
  assert.equal(f.requests(), 1);
  f.advance(reviewIntervalMs);
  await f.restart().visit(safe);
  assert.equal(f.requests(), 1);
  await f.controller.completed(safe);
  assert.equal(f.requests(), 2);
});

test("both triggers share exactly 60 days, even when concurrent", async () => {
  const f = fixture();
  for (let i = 0; i < 9; i++) await f.controller.visit(safe);
  await Promise.all([f.controller.visit(safe), f.controller.completed(safe)]);
  assert.equal(f.requests(), 1);
  f.advance(reviewIntervalMs - 1);
  await f.controller.completed(safe);
  assert.equal(f.requests(), 1);
  f.advance(1);
  await f.controller.completed(safe);
  assert.equal(f.requests(), 2);
});

test("a blocked tenth visit waits for a free surface without incrementing again", async () => {
  const f = fixture();
  for (let i = 0; i < 10; i++) await f.controller.visit(() => false);
  assert.equal(f.requests(), 0);
  await f.controller.retryVisitMilestone(safe);
  assert.equal(f.requests(), 1);
  assert.equal(f.state().visits, 10);
});

test("completion before ten visits does not consume the later milestone", async () => {
  const f = fixture();
  await f.controller.completed(safe);
  for (let i = 0; i < 10; i++) await f.controller.visit(safe);
  assert.equal(f.requests(), 1);
  assert.equal(f.state().visitMilestoneConsumed, false);
  f.advance(reviewIntervalMs);
  await f.controller.retryVisitMilestone(safe);
  assert.equal(f.requests(), 2);
});

test("unsupported platforms and corrupt or unavailable storage never request", async () => {
  for (const overrides of [
    { available: async () => false },
    {
      read: async () => {
        throw new Error("unavailable");
      },
    },
    { read: async () => "invalid json" },
    { read: async () => JSON.stringify({ visits: -1 }) },
    {
      write: async () => {
        throw new Error("unavailable");
      },
    },
  ]) {
    const f = fixture(overrides);
    await f.controller.completed(safe);
    assert.equal(f.requests(), 0);
  }
});

test("a failed native request is still an attempt, and the queue recovers", async () => {
  let attempts = 0;
  const f = fixture({
    request: async () => {
      attempts++;
      throw new Error("store failed");
    },
  });
  await f.controller.completed(safe);
  await f.controller.completed(safe);
  assert.equal(attempts, 1);
  f.advance(reviewIntervalMs);
  await f.controller.completed(safe);
  assert.equal(attempts, 2);
});

test("navigation during the availability check cancels the request", async () => {
  let active = true;
  const f = fixture({
    available: async () => {
      active = false;
      return true;
    },
  });
  await f.controller.completed(() => active);
  assert.equal(f.requests(), 0);
  assert.equal(f.state(), null);
});
