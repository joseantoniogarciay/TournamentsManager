import { execFileSync, spawnSync } from "node:child_process";
import { readFile } from "node:fs/promises";
import { resolve, join } from "node:path";
import { fileURLToPath, URL } from "node:url";
import process from "node:process";
import { error as logError } from "node:console";
import { fileInventory } from "./stage-web.mjs";

const repository = fileURLToPath(new URL("../../", import.meta.url));
const [mode, candidate] = process.argv.slice(2);
const modes = ["check", "preview", "publish"];

try {
  if (!modes.includes(mode) || !candidate)
    throw new Error(
      "Uso: node infra/cloudflare/publish-web.mjs <check|preview|publish> <directorio-release>",
    );
  const release = resolve(candidate);
  const manifest = JSON.parse(await readFile(join(release, "release.json"), "utf8"));
  const files = (await fileInventory(release)).filter(([name]) => name !== "release.json");
  if (JSON.stringify(files) !== JSON.stringify(manifest.files))
    throw new Error("El candidato cambió tras prepararlo; no se publica");
  if (execFileSync("git", ["status", "--porcelain"], { cwd: repository, encoding: "utf8" }).trim())
    throw new Error("El árbol Git debe estar limpio antes de publicar");
  const head = execFileSync("git", ["rev-parse", "HEAD"], {
    cwd: repository,
    encoding: "utf8",
  }).trim();
  if (head !== manifest.infrastructureSHA)
    throw new Error("Usa el checkout de infraestructura que preparó este candidato");
  const env = {
    ...process.env,
    WRANGLER_SEND_METRICS: "false",
    WRANGLER_LOG_PATH: join(repository, ".config/cloudflare/wrangler-logs"),
  };
  const configs = mode === "preview" ? ["wrangler.json"] : ["wrangler.json", "wrangler-www.json"];
  for (const config of configs) {
    const command = mode === "preview" ? ["versions", "upload"] : ["deploy"];
    const result = spawnSync(
      "pnpm",
      [
        "exec",
        "wrangler",
        ...command,
        "--config",
        join(release, config),
        ...(mode === "check"
          ? ["--dry-run", "--outdir", join(repository, ".config/cloudflare/dry-run", config)]
          : []),
      ],
      { cwd: repository, env, stdio: "inherit" },
    );
    if (result.error || result.status !== 0)
      throw new Error("Wrangler no confirmó el paso; revisar estado antes de repetir");
  }
} catch (error) {
  logError(error instanceof Error ? error.message : "No se pudo publicar Cloudflare");
  process.exitCode = 1;
}
