import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { runInNewContext } from "node:vm";
import test from "node:test";
import { transpileForTest } from "./helpers/transpile-typescript.mjs";

const module = { exports: {} };
runInNewContext(transpileForTest(readFileSync("packages/design-tokens/src/tokens.ts", "utf8")), {
  module,
  exports: module.exports,
});
const { color, gradient, themeColors } = module.exports;

function luminance(hex) {
  const channels = [1, 3, 5].map((offset) => {
    const value = parseInt(hex.slice(offset, offset + 2), 16) / 255;
    return value <= 0.04045 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4;
  });
  return channels[0] * 0.2126 + channels[1] * 0.7152 + channels[2] * 0.0722;
}

function contrast(foreground, background, minimum, role) {
  const values = [luminance(foreground), luminance(background)].sort((a, b) => a - b);
  const ratio = (values[1] + 0.05) / (values[0] + 0.05);
  assert.ok(ratio >= minimum, `${role}: ${ratio.toFixed(2)} < ${minimum}`);
}

for (const [theme, colors] of Object.entries(themeColors)) {
  test(`${theme}: normal text, placeholders and feedback meet WCAG AA on every surface`, () => {
    for (const background of Object.values(colors.surface)) {
      for (const foreground of [
        colors.text.primary,
        colors.text.secondary,
        colors.text.placeholder,
        colors.feedback.error,
        colors.feedback.success,
      ]) {
        contrast(foreground, background, 4.5, `${theme} text ${foreground} on ${background}`);
      }
    }
  });
  test(`${theme}: navigation tint, focus and field boundaries remain distinguishable`, () => {
    for (const background of Object.values(colors.surface)) {
      for (const foreground of [
        colors.indicator.default,
        colors.border.focus,
        colors.border.default,
      ]) {
        contrast(foreground, background, 3, `${theme} control ${foreground} on ${background}`);
      }
    }
  });
}

test("filled brand controls retain readable white labels across the gradient", () => {
  // Sample the rendered sRGB overlay, including transparent stops; testing only
  // the two named colours would miss contrast failures within a gradient.
  for (const definition of Object.values(gradient)) {
    for (let sample = 0; sample <= 100; sample++) {
      const alpha = sample / 100;
      const mixed = [1, 3, 5]
        .map((offset) => {
          const base = parseInt(color.brand.primary.slice(offset, offset + 2), 16);
          const overlay = parseInt(definition.colors.at(-1).slice(offset, offset + 2), 16);
          return Math.round(base * (1 - alpha) + overlay * alpha)
            .toString(16)
            .padStart(2, "0");
        })
        .join("");
      contrast(color.text.inverse, `#${mixed}`, 4.5, "onBrand");
    }
  }
});
