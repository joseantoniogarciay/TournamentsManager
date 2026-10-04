import { spawnSync } from "node:child_process";
import { error as logError } from "node:console";
import process from "node:process";
import { readdir, readFile, rm, writeFile } from "node:fs/promises";
import { join, resolve } from "node:path";
import { pathToFileURL } from "node:url";
import { telemetryConfig } from "../../apps/client/src/shared/analytics/telemetry-config.ts";

async function removePublicSourceMaps(directory) {
  for (const item of await readdir(directory, { withFileTypes: true })) {
    const path = join(directory, item.name);
    if (item.isDirectory()) await removePublicSourceMaps(path);
    else if (item.name.endsWith(".map")) await rm(path);
    else if (item.name.endsWith(".js")) {
      const source = await readFile(path, "utf8");
      const clean = source.replace(/\/\/[#@]\s*sourceMappingURL=[^\r\n]*/g, "");
      if (source !== clean) await writeFile(path, clean);
    }
  }
}

export async function prepareWebTelemetry(environment, directory, release, options = {}) {
  const env = options.env ?? process.env;
  const configuration = telemetryConfig(environment, env.EXPO_PUBLIC_POSTHOG_PRODUCTION_API_KEY);
  if (configuration) {
    const credentialFile = env.FASTTOURNEY_PROD_POSTHOG_SYMBOLS_CONFIG;
    if (!credentialFile)
      throw new Error("Falta el archivo privado de símbolos del entorno PostHog");
    if (!/^[a-f0-9]{40}$/.test(release)) throw new Error("El release debe ser un SHA completo");
    const credentials = await readFile(credentialFile, "utf8");
    if (
      ["PROJECT_ID", "API_KEY", "HOST"].some(
        (field) =>
          (credentials.match(new RegExp(`^POSTHOG_CLI_${field}=`, "gm")) ?? []).length !== 1,
      ) ||
      !/^POSTHOG_CLI_PROJECT_ID=255144\s*$/m.test(credentials) ||
      !/^POSTHOG_CLI_API_KEY=phx_[A-Za-z0-9]+\s*$/m.test(credentials) ||
      !/^POSTHOG_CLI_HOST=https:\/\/eu\.posthog\.com\s*$/m.test(credentials)
    )
      throw new Error(
        "El archivo de símbolos requiere proyecto 255144, clave CLI y host UE reales",
      );
    // Ignore credentials inherited from another build/project. The explicit file
    // supplies the upload key and project ID; the destination is always EU.
    const uploadEnvironment = Object.fromEntries(
      Object.entries(env).filter(([key]) => !key.startsWith("POSTHOG_")),
    );
    const result = (options.run ?? spawnSync)(
      "posthog-cli",
      [
        "--host",
        "https://eu.posthog.com",
        "--dotenv-file",
        resolve(credentialFile),
        "sourcemap",
        "process",
        "--directory",
        join(resolve(directory), "_expo/static/js/web"),
        "--release-name",
        `fasttourney-web-${environment}`,
        "--release-version",
        release,
        "--skip-on-conflict",
        "--delete-after",
      ],
      { env: uploadEnvironment, stdio: "inherit", timeout: 120_000 },
    );
    if (result.error || result.status !== 0)
      throw new Error("No se ha confirmado la subida de símbolos web; el release no se publica");
  }
  await removePublicSourceMaps(directory);
}

if (process.argv[1] && import.meta.url === pathToFileURL(resolve(process.argv[1])).href) {
  try {
    await prepareWebTelemetry(process.argv[2], process.argv[3], process.argv[4]);
  } catch (error) {
    logError(error instanceof Error ? error.message : "No se pudo preparar la telemetría web");
    process.exitCode = 1;
  }
}
