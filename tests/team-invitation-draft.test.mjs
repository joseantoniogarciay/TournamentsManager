import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";
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
