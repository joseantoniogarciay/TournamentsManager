import { transformSync } from "esbuild";

// Keep the existing VM fixtures and dependency stubs independent of the
// compiler's programmatic API. Type checking runs separately through tsc.
/**
 * @param {string} source
 * @param {{ jsx?: boolean }} options
 */
export function transpileForTest(source, { jsx = false } = {}) {
  return transformSync(source, {
    loader: jsx ? "tsx" : "ts",
    format: "cjs",
    target: "es2022",
    jsx: "automatic",
  }).code;
}
