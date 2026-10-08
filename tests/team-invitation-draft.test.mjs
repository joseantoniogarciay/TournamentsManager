import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { URL } from "node:url";
import { runInNewContext } from "node:vm";
import test from "node:test";
import ts from "typescript";

const sourcePath = resolve("apps/client/src/features/league-creation/team-invitation.ts");
const code = ts.transpileModule(readFileSync(sourcePath, "utf8"), {
  compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 },
}).outputText;

function loadInvitation(platform, values) {
  const storage = {
    getItem: async (key) => values.get(key) ?? null,
    setItem: async (key, value) => values.set(key, value),
    removeItem: async (key) => values.delete(key),
    getItemAsync: async (key) => values.get(key) ?? null,
    setItemAsync: async (key, value) => values.set(key, value),
    deleteItemAsync: async (key) => values.delete(key),
  };
  const module = { exports: {} };
  const require = (name) => {
    if (name === "@react-native-async-storage/async-storage")
      return { __esModule: true, default: storage };
    if (name === "expo-secure-store") return storage;
    if (name === "react-native") return { Platform: { OS: platform } };
    throw new Error(`Unexpected module: ${name}`);
  };
  runInNewContext(`(function(require,module,exports){${code}\n})`, {}, { filename: sourcePath })(
    require,
    module,
    module.exports,
  );
  return module.exports;
}

for (const platform of ["web", "ios", "android"]) {
  test(`${platform}: invitation name survives login remount, stays scoped and clears with invitation`, async () => {
    const values = new Map();
    const invitation = loadInvitation(platform, values);
    const token = "a".repeat(43);
    await invitation.rememberPendingTeamInvitation(token);
    await invitation.rememberPendingTeamInvitationName(token, "Cóndores elegidos antes del login");

    const restored = loadInvitation(platform, values);
    assert.equal(await restored.getPendingTeamInvitation(), token);
    assert.equal(
      await restored.getPendingTeamInvitationName(token),
      "Cóndores elegidos antes del login",
    );
    assert.equal(await restored.getPendingTeamInvitationName("b".repeat(43)), null);
    await restored.clearPendingTeamInvitation();
    assert.equal(await restored.getPendingTeamInvitation(), null);
    assert.equal(await restored.getPendingTeamInvitationName(token), null);
    assert.equal(values.size, 0);
  });
}

test("legacy token-only invitations restore without a draft and corrupt drafts are ignored", async () => {
  const token = "a".repeat(43);
  const values = new Map([["tm-pending-team-invitation", token]]);
  const invitation = loadInvitation("web", values);
  assert.equal(await invitation.getPendingTeamInvitation(), token);
  assert.equal(await invitation.getPendingTeamInvitationName(token), null);
  values.set("tm-pending-team-invitation-draft", "{invalid JSON");
  assert.equal(await invitation.getPendingTeamInvitationName(token), null);
  values.set("tm-pending-team-invitation-draft", JSON.stringify({ token, name: 42 }));
  assert.equal(await invitation.getPendingTeamInvitationName(token), null);
});

const nativeIntentPath = resolve("apps/client/src/app/+native-intent.ts");
const nativeIntentCode = ts.transpileModule(readFileSync(nativeIntentPath, "utf8"), {
  compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 },
}).outputText;

function loadPlainModule(path) {
  const code = ts.transpileModule(readFileSync(resolve(path), "utf8"), {
    compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 },
  }).outputText;
  const module = { exports: {} };
  runInNewContext(`(function(module,exports){${code}\n})`, { URL })(module, module.exports);
  return module.exports;
}

function loadNativeIntent(invitation) {
  const gate = loadPlainModule("apps/client/src/shared/navigation/deep-link-gate.ts");
  const social = loadPlainModule("apps/client/src/features/social-authentication/config.ts");
  const module = { exports: {} };
  runInNewContext(
    `(function(require,module,exports){${nativeIntentCode}\n})`,
    { URL },
    { filename: nativeIntentPath },
  )(
    (name) => {
      if (name === "@/features/league-creation/team-invitation") return invitation;
      if (name === "@/shared/navigation/deep-link-gate") return gate;
      if (name === "@/features/social-authentication/config") return social;
      throw new Error(`Unexpected module: ${name}`);
    },
    module,
    module.exports,
  );
  return { ...module.exports, gate };
}

for (const platform of ["ios", "android"]) {
  for (const initial of [true, false]) {
    test(`${platform}: ${initial ? "cold" : "warm"} link stores token before sanitized navigation and login remount`, async () => {
      const values = new Map();
      const invitation = loadInvitation(platform, values);
      const intent = loadNativeIntent(invitation);
      const token = "c".repeat(43);
      const route = await intent.redirectSystemPath({
        path: `fasttourney-local://join-team#${token}`,
        initial,
      });
      assert.equal(route, initial ? "/" : "/join-team?invitationRevision=1");
      if (initial)
        assert.equal(
          intent.gate.consumeDeferredInitialDeepLink(),
          "/join-team?invitationRevision=1",
        );
      const screen = loadInvitation(platform, values);
      assert.equal(await screen.getPendingTeamInvitation(), token);
      await screen.rememberPendingTeamInvitationName(token, "Equipo antes del login");
      const returnedScreen = loadInvitation(platform, values);
      assert.equal(await returnedScreen.getPendingTeamInvitation(), token);
      assert.equal(
        await returnedScreen.getPendingTeamInvitationName(token),
        "Equipo antes del login",
      );
    });
  }
}

test("native links accept universal and Expo Go fragments and reject malformed invitation without reusing prior draft", async () => {
  const values = new Map();
  const invitation = loadInvitation("ios", values);
  const intent = loadNativeIntent(invitation);
  const token = "d".repeat(43);
  for (const base of ["https://example.test/join-team", "exp://127.0.0.1:8082/--/join-team"]) {
    assert.match(
      await intent.redirectSystemPath({ path: `${base}#${token}`, initial: false }),
      /^\/join-team\?invitationRevision=\d+$/,
    );
    assert.equal(await invitation.getPendingTeamInvitation(), token);
  }
  await invitation.rememberPendingTeamInvitationName(token, "Borrador anterior");
  assert.equal(
    await intent.redirectSystemPath({
      path: "fasttourney-local://join-team#invalid",
      initial: false,
    }),
    "/join-team?invitationRevision=3",
  );
  assert.equal(await invitation.getPendingTeamInvitation(), null);
  assert.equal(await invitation.getPendingTeamInvitationName(token), null);
  const unrelated = "fasttourney-local://reset-password#unrelated";
  assert.equal(
    await intent.redirectSystemPath({ path: unrelated, initial: false }),
    "/reset-password#unrelated",
  );
});

test("native invitation navigation waits for persistence and fails closed without exposing the fragment", async () => {
  let finish;
  let stored = false;
  const invitation = loadInvitation("android", new Map());
  const intent = loadNativeIntent({
    ...invitation,
    rememberPendingTeamInvitation: () =>
      new Promise((resolve) => {
        finish = resolve;
      }),
  });
  const pending = intent.redirectSystemPath({
    path: `fasttourney-local://join-team#${"e".repeat(43)}`,
    initial: false,
  });
  pending.then(() => {
    stored = true;
  });
  await Promise.resolve();
  assert.equal(stored, false);
  finish();
  assert.equal(await pending, "/join-team?invitationRevision=1");
  const failing = loadNativeIntent({
    ...invitation,
    rememberPendingTeamInvitation: async () => {
      throw new Error("storage unavailable");
    },
  });
  assert.equal(
    await failing.redirectSystemPath({
      path: `fasttourney-local://join-team#${"e".repeat(43)}`,
      initial: true,
    }),
    "/",
  );
});

test("native gate keeps OAuth returns and cold navigation behavior while repeated invitations get distinct revisions", async () => {
  const invitation = loadInvitation("ios", new Map());
  const intent = loadNativeIntent(invitation);
  const oauth = "fasttourney://oauth/apple-complete?code=test";
  assert.equal(await intent.redirectSystemPath({ path: oauth, initial: false }), null);
  assert.equal(await intent.redirectSystemPath({ path: oauth, initial: true }), "/account");
  assert.equal(await intent.redirectSystemPath({ path: null, initial: true }), null);
  assert.equal(
    await intent.redirectSystemPath({ path: "fasttourney-local://tournament/test", initial: true }),
    "/",
  );
  assert.equal(intent.gate.consumeDeferredInitialDeepLink(), "/tournament/test");
  const token = "f".repeat(43);
  assert.equal(
    await intent.redirectSystemPath({
      path: `fasttourney-local://join-team#${token}`,
      initial: false,
    }),
    "/join-team?invitationRevision=1",
  );
  assert.equal(
    await intent.redirectSystemPath({
      path: `fasttourney-local://join-team#${token}`,
      initial: false,
    }),
    "/join-team?invitationRevision=2",
  );
});
