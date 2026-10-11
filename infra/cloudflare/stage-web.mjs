import { createHash } from "node:crypto";
import { execFileSync } from "node:child_process";
import { cp, mkdir, readFile, readdir, rename, rm, writeFile } from "node:fs/promises";
import { join, resolve } from "node:path";
import { fileURLToPath, pathToFileURL, URL } from "node:url";
import process from "node:process";
import { error as logError, log } from "node:console";
import { assetHeaders } from "./web-headers.mjs";

const repository = fileURLToPath(new URL("../../", import.meta.url));
const shaPattern = /^[a-f0-9]{40}$/;

export async function fileInventory(directory, prefix = "") {
  const files = [];
  for (const item of await readdir(directory, { withFileTypes: true })) {
    const name = `${prefix}${item.name}`;
    if (item.isSymbolicLink())
      throw new Error("Un release no puede contener enlaces simbólicos internos");
    if (item.isDirectory())
      files.push(...(await fileInventory(join(directory, item.name), `${name}/`)));
    else {
      if (/\.(map|pem|key)$|(^|\/)\.env(?:\.|$)/i.test(name))
        throw new Error("El release contiene archivos privados o source maps públicos");
      const content = await readFile(join(directory, item.name));
      if (name.endsWith(".js") && content.includes("sourceMappingURL="))
        throw new Error("El bundle conserva referencias a source maps públicos");
      files.push([name, createHash("sha256").update(content).digest("hex")]);
    }
  }
  return files.sort(([first], [second]) => first.localeCompare(second));
}

export function webConfig(name, redirect = false) {
  return {
    name,
    main: "./web-worker.mjs",
    compatibility_date: "2026-10-01",
    workers_dev: false,
    preview_urls: !redirect,
    observability: { enabled: false },
    routes: [
      { pattern: `${redirect ? "www." : ""}fasttourney.com/*`, zone_name: "fasttourney.com" },
    ],
    ...(redirect
      ? {}
      : {
          assets: {
            directory: "./assets",
            binding: "ASSETS",
            html_handling: "drop-trailing-slash",
            not_found_handling: "none",
            run_worker_first: ["/", "/tournament/*"],
          },
        }),
  };
}

export async function stageWeb(webSHA, infrastructureSHA, source, releasesRoot) {
  if (!shaPattern.test(webSHA) || !shaPattern.test(infrastructureSHA))
    throw new Error("Ambas revisiones deben ser SHA completos");
  const manifest = JSON.parse(await readFile(join(source, "deployment.json"), "utf8"));
  if (manifest.commit !== webSHA)
    throw new Error("El manifiesto del artefacto no corresponde al SHA indicado");
  const shell = await readFile(join(source, "index.html"), "utf8");
  if (!shell.includes('href="https://fasttourney.com/"'))
    throw new Error("El artefacto debe declarar el canonical productivo");
  await fileInventory(source);
  const release = join(releasesRoot, `${webSHA}-${infrastructureSHA}`);
  await mkdir(releasesRoot, { recursive: true });
  const staging = `${release}.staging`;
  await mkdir(staging); // Fail closed rather than overwriting a candidate.
  try {
    await cp(source, join(staging, "assets"), { recursive: true, dereference: false });
    for (const file of ["web-worker.mjs", "web-headers.mjs"])
      await cp(join(repository, "infra/cloudflare", file), join(staging, file));
    await writeFile(join(staging, "assets/_headers"), assetHeaders());
    await writeFile(
      join(staging, "wrangler.json"),
      `${JSON.stringify(webConfig("fasttourney-web"), null, 2)}\n`,
    );
    await writeFile(
      join(staging, "wrangler-www.json"),
      `${JSON.stringify(webConfig("fasttourney-web-redirect", true), null, 2)}\n`,
    );
    const files = await fileInventory(staging);
    await writeFile(
      join(staging, "release.json"),
      `${JSON.stringify({ webSHA, infrastructureSHA, files }, null, 2)}\n`,
    );
    // rename must not replace an already prepared directory.
    await rename(staging, release);
    return release;
  } finally {
    await rm(staging, { recursive: true, force: true });
  }
}

if (process.argv[1] && import.meta.url === pathToFileURL(resolve(process.argv[1])).href) {
  try {
    const infrastructureSHA = execFileSync("git", ["rev-parse", "HEAD"], {
      cwd: repository,
      encoding: "utf8",
    }).trim();
    if (
      execFileSync("git", ["status", "--porcelain"], { cwd: repository, encoding: "utf8" }).trim()
    )
      throw new Error("El árbol Git debe estar limpio antes de preparar Cloudflare");
    const webSHA = process.argv[2];
    if (!shaPattern.test(webSHA ?? ""))
      throw new Error("Uso: node infra/cloudflare/stage-web.mjs <SHA-web-completo>");
    const source = `/opt/homebrew/var/www/fasttourney/prod/releases/${webSHA}`;
    log(
      await stageWeb(
        webSHA,
        infrastructureSHA,
        source,
        join(repository, ".config/cloudflare/releases"),
      ),
    );
  } catch (error) {
    logError(error instanceof Error ? error.message : "No se pudo preparar Cloudflare");
    process.exitCode = 1;
  }
}
