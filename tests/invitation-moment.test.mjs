import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { runInNewContext } from "node:vm";
import test from "node:test";
import { transpileForTest } from "./helpers/transpile-typescript.mjs";

function fixture() {
  const module = { exports: {} };
  runInNewContext(
    transpileForTest(
      readFileSync("apps/client/src/features/league-creation/invitation-moment.ts", "utf8"),
    ),
    { module, exports: module.exports },
  );
  return module.exports;
}

test("only a confirmed invitation queued for this tournament and account is consumed once", () => {
  const f = fixture();
  assert.equal(f.takeInvitationMoment("one", "alice"), false);
  f.rememberInvitationMoment("one", "alice");
  assert.equal(f.takeInvitationMoment("two", "alice"), false);
  assert.equal(f.takeInvitationMoment("one", "alice"), true);
  assert.equal(f.takeInvitationMoment("one", "alice"), false);
});

test("an account change discards the old moment, and new outcomes replace old ones", () => {
  const f = fixture();
  f.rememberInvitationMoment("one", "alice");
  assert.equal(f.takeInvitationMoment("one", "bob"), false);
  assert.equal(f.takeInvitationMoment("one", "alice"), false);
  f.rememberInvitationMoment("one", "alice");
  f.rememberInvitationMoment("two", "alice");
  assert.equal(f.takeInvitationMoment("one", "alice"), false);
  assert.equal(f.takeInvitationMoment("two", "alice"), true);
});
