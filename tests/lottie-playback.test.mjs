import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { runInNewContext } from "node:vm";
import test from "node:test";
import { transpileForTest } from "./helpers/transpile-typescript.mjs";

function fixture(initialPlayback) {
  const slots = [];
  const effects = [];
  const players = [];
  let index = 0;
  let playback = initialPlayback;
  let colors = {
    indicator: { default: "accent" },
    text: { secondary: "ink" },
    surface: { subtle: "subtle" },
  };
  const changed = (previous, deps) =>
    !previous || deps.some((value, i) => value !== previous.deps[i]);
  const modules = {
    react: {
      useRef: (current) => (slots[index++] ??= { current }),
      useMemo(factory, deps) {
        const i = index++;
        if (changed(slots[i], deps)) slots[i] = { deps, value: factory() };
        return slots[i].value;
      },
      useEffect(effect, deps) {
        const i = index++;
        if (changed(slots[i], deps))
          effects.push(() => {
            slots[i]?.cleanup?.();
            slots[i] = { deps, cleanup: effect() };
          });
      },
    },
    "react/jsx-runtime": { jsx: (type, props) => ({ type, props }) },
    "lottie-web/build/player/lottie_light": {
      loadAnimation() {
        const listeners = new Map();
        const calls = [];
        // Lottie checkLoaded sets this synchronously, then queues DOMLoaded.
        const player = {
          isLoaded: true,
          calls,
          goToAndStop: (frame) => calls.push(["still", frame]),
          goToAndPlay: (frame) => calls.push(["play", frame]),
          addEventListener: (name, callback) => listeners.set(name, callback),
          removeEventListener: (name) => listeners.delete(name),
          destroy: () => calls.push(["destroy"]),
          ready: () => listeners.get("DOMLoaded")?.(),
        };
        players.push(player);
        return player;
      },
    },
    "@tournaments-manager/design-tokens": {
      illustration: { celebrationWidth: 192, emptyWidth: 256, aspectRatio: 4 / 3 },
    },
    "@/shared/preferences/preferences-provider": { usePreferences: () => ({ colors }) },
    "./tournament-illustration-data": { createTournamentIllustration: () => ({ op: 48 }) },
    "./use-illustration-playback": { useIllustrationPlayback: () => playback },
  };
  const module = { exports: {} };
  runInNewContext(
    transpileForTest(
      readFileSync("apps/client/src/shared/ui/tournament-illustration.web.tsx", "utf8"),
      { jsx: true },
    ),
    {
      module,
      exports: module.exports,
      structuredClone: globalThis.structuredClone,
      require: (name) => {
        assert.ok(name in modules, `Unexpected module ${name}`);
        return modules[name];
      },
    },
  );
  return {
    players,
    render(next = playback, props = {}) {
      playback = next;
      index = 0;
      const tree = module.exports.TournamentIllustration({ variant: "trophy", ...props });
      tree.props.ref.current ??= {};
      effects.splice(0).forEach((effect) => effect());
    },
    changeTheme() {
      colors = globalThis.structuredClone(colors);
    },
    unmount() {
      slots.forEach((slot) => slot?.cleanup?.());
    },
  };
}

const moving = { reducedMotion: false, motionReady: true, playing: true };

test("web starts after DOMLoaded even when isLoaded is already true", () => {
  const f = fixture(moving);
  f.render();
  const player = f.players[0];
  assert.deepEqual(player.calls, [], "Do not consume playback before SVG readiness");
  player.ready();
  assert.deepEqual(player.calls, [
    ["still", 47],
    ["play", 0],
  ]);
  f.render({ ...moving });
  assert.equal(player.calls.filter(([action]) => action === "play").length, 1);
  f.unmount();
});

test("web waits for the motion preference and never replays after leaving", () => {
  const f = fixture({ reducedMotion: true, motionReady: false, playing: false });
  f.render();
  const player = f.players[0];
  player.ready();
  assert.deepEqual(player.calls, [["still", 47]]);
  f.render(moving);
  f.render({ ...moving, playing: false });
  f.render(moving);
  assert.deepEqual(player.calls, [
    ["still", 47],
    ["play", 0],
    ["still", 47],
  ]);
  f.unmount();
  assert.deepEqual(player.calls.at(-1), ["destroy"]);
});

test("web keeps reduced motion static and recreates a themed SVG without replay", () => {
  const f = fixture({ reducedMotion: true, motionReady: true, playing: false });
  f.render();
  f.players[0].ready();
  assert.ok(f.players[0].calls.every(([action]) => action !== "play"));
  f.changeTheme();
  f.render(moving);
  assert.deepEqual(f.players[0].calls.at(-1), ["destroy"]);
  assert.deepEqual(f.players[1].calls, []);
  f.players[1].ready();
  assert.ok(f.players[1].calls.every(([action]) => action !== "play"));
  f.unmount();
});

test("a welcome already presented stays static despite a motion-enabled system", () => {
  const f = fixture(moving);
  f.render(moving, { variant: "welcome", animate: false });
  f.players[0].ready();
  assert.ok(f.players[0].calls.every(([action]) => action !== "play"));
  f.unmount();
});
