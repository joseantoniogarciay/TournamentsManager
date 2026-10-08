import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { runInNewContext } from "node:vm";
import test from "node:test";
import { setImmediate } from "node:timers";
import ts from "typescript";

const settle = () => new Promise((resolve) => setImmediate(resolve));
function deferred() {
  let resolve, reject;
  const promise = new Promise((a, b) => {
    resolve = a;
    reject = b;
  });
  return { promise, resolve, reject };
}
function harness() {
  const slots = [];
  const effects = [];
  let index = 0;
  let visible = true;
  let tree;
  const feedback = [];
  const calls = [];
  const pending = deferred();
  let dismissals = 0;
  let preparations = 0;
  const proof = {
    error: null,
    isConfigured: false,
    isLoading: false,
    prepare: () => preparations++,
    start: () => {},
  };
  const react = {
    useState(initial) {
      const i = index++;
      if (!(i in slots)) slots[i] = initial;
      return [
        slots[i],
        (value) => (slots[i] = typeof value === "function" ? value(slots[i]) : value),
      ];
    },
    useRef(initial) {
      return (slots[index++] ??= { current: initial });
    },
    useCallback(fn, deps) {
      const i = index++;
      const previous = slots[i];
      if (!previous || deps.some((v, n) => v !== previous.deps[n])) slots[i] = { deps, fn };
      return slots[i].fn;
    },
    useEffect(fn, deps) {
      const i = index++;
      const previous = slots[i];
      if (!previous || deps.some((v, n) => v !== previous.deps[n])) {
        effects.push(() => {
          previous?.cleanup?.();
          slots[i] = { deps, cleanup: fn() };
        });
      }
    },
  };
  const show = (value) => feedback.push(value);
  const translate = (key) => key;
  const imports = {
    react,
    "react/jsx-runtime": {
      jsx: (type, props) => ({ type, props }),
      jsxs: (type, props) => ({ type, props }),
      Fragment: "Fragment",
    },
    "expo-router": { router: { back() {} } },
    "react-native": { StyleSheet: { create: (value) => value }, View: "View" },
    "@tournaments-manager/design-tokens": { space: {} },
    "@/features/account-access/api": {
      GoogleLinkError: class extends Error {},
      getAccountAccessMethods: async () => ({ methods: { password: true } }),
      reauthenticateWithPassword: (...args) => {
        calls.push(args);
        return pending.promise;
      },
    },
    "@/features/federated-google/use-google-identity-proof": {
      useGoogleIdentityProof: () => proof,
    },
    "@/shared/feedback/feedback-provider": { useFeedback: () => ({ show }) },
    "@/shared/feedback/request-failure": {
      getRequestFailure: () => ({ messageKey: "common_request_error" }),
    },
    "@/shared/i18n/locale": { getTranslator: () => translate },
    "@/shared/ui": Object.fromEntries(
      ["Button", "ModalDialog", "Text", "TextField"].map((v) => [v, v]),
    ),
  };
  const code = ts.transpileModule(
    readFileSync("apps/client/src/app/(tabs)/account/google-link.tsx", "utf8"),
    {
      compilerOptions: {
        module: ts.ModuleKind.CommonJS,
        target: ts.ScriptTarget.ES2022,
        jsx: ts.JsxEmit.ReactJSX,
      },
    },
  ).outputText;
  const module = { exports: {} };
  runInNewContext(`(function(require,module,exports){${code}\n})`, {})(
    (name) => {
      if (!(name in imports)) throw new Error(`Unexpected import: ${name}`);
      return imports[name];
    },
    module,
    module.exports,
  );
  const dismiss = () => dismissals++;
  const linked = () => {};
  function render() {
    index = 0;
    tree = module.exports.GoogleLinkDialog({ visible, onDismiss: dismiss, onLinked: linked });
    effects.splice(0).forEach((fn) => fn());
    return tree;
  }
  function nodes(node) {
    if (!node || typeof node !== "object") return [];
    const children = node.props?.children;
    return [
      node,
      ...(Array.isArray(children) ? children : [children]).flatMap((child) =>
        Array.isArray(child) ? child.flatMap(nodes) : nodes(child),
      ),
    ];
  }
  return {
    pending,
    calls,
    feedback,
    proof,
    render,
    get dismissals() {
      return dismissals;
    },
    get preparations() {
      return preparations;
    },
    setVisible(value) {
      visible = value;
      render();
    },
    field: () => nodes(tree).find((node) => node.type === "TextField").props,
    button: (label) =>
      nodes(tree).find((node) => node.type === "Button" && node.props.label === label)?.props,
    async open() {
      render();
      await settle();
      render();
    },
    enter() {
      this.field().onChangeText("existing-test-password");
      render();
    },
  };
}

test("reauthentication blocks duplicate events before render and leaves unconfigured Google disabled", async () => {
  const h = harness();
  await h.open();
  h.enter();
  const submit = h.button("account_google_link_reauthenticate").onPress;
  submit();
  submit();
  assert.equal(h.calls.length, 1);
  h.render();
  assert.equal(h.button("account_google_link_reauthenticate").disabled, true);
  assert.equal(h.button("account_google_link_reauthenticate").loading, true);
  h.pending.resolve("one-use-ticket");
  await settle();
  h.render();
  assert.equal(h.button("account_google_link_continue").disabled, true);
  assert.equal(h.preparations, 0);
});

for (const outcome of ["success", "failure"]) {
  test(`closing reauthentication ignores late ${outcome} and reopening starts empty`, async () => {
    const h = harness();
    await h.open();
    h.enter();
    h.button("account_google_link_reauthenticate").onPress();
    h.setVisible(false);
    if (outcome === "success") h.pending.resolve("old-ticket");
    else h.pending.reject(new Error("private backend detail"));
    await settle();
    assert.equal(h.feedback.length, 0);
    assert.equal(h.dismissals, 0);
    h.setVisible(true);
    await settle();
    h.render();
    assert.equal(h.field().value, "");
    assert.equal(h.button("account_google_link_reauthenticate").disabled, true);
    assert.equal(h.button("account_google_link_continue"), undefined);
  });
}

test("provider error is consumed once and cannot repeat feedback after dismissal or reopening", async () => {
  const h = harness();
  await h.open();
  h.proof.error = new Error("private provider error");
  h.render();
  h.render();
  h.setVisible(false);
  h.setVisible(true);
  assert.equal(h.dismissals, 1);
  assert.equal(h.feedback.length, 1);
  assert.equal(h.feedback[0].message, "common_request_error");
});

test("a provider error arriving after close cannot dismiss a newly opened dialog", async () => {
  const h = harness();
  await h.open();
  h.setVisible(false);
  h.proof.error = new Error("late provider error");
  h.render();
  h.setVisible(true);
  assert.equal(h.dismissals, 0);
  assert.equal(h.feedback.length, 0);
});
