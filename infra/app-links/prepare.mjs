import console from "node:console";
import process from "node:process";
import { existsSync, mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { resolve } from "node:path";
import { fileURLToPath } from "node:url";

const packages = {
  development: "com.fasttourney.app.dev",
  production: "com.fasttourney.app",
};
const paths = ["/link/*", "/join-team", "/tournament/*"];

// Format and environment checks cannot establish ownership of a signing key.
export function validateAssociations(environment, apple, android) {
  const packageName = packages[environment];
  if (!packageName) throw new Error("Entorno de asociaciones desconocido.");
  const details = apple?.applinks?.details;
  const appID = details?.[0]?.appIDs?.[0];
  const prefix = typeof appID === "string" ? appID.slice(0, -(packageName.length + 1)) : "";
  if (
    !/^[A-Z0-9]{10}$/.test(prefix) ||
    appID !== `${prefix}.${packageName}` ||
    details.length !== 1 ||
    details[0].appIDs.length !== 1 ||
    apple.webcredentials?.apps?.length !== 1 ||
    apple.webcredentials.apps[0] !== appID ||
    details[0].appID ||
    details[0].paths ||
    Object.keys(apple).some((key) => !["applinks", "webcredentials"].includes(key)) ||
    apple.applinks.apps?.length
  ) {
    throw new Error("AASA debe asociar únicamente la app del entorno y su App ID Prefix real.");
  }
  const components = details[0].components;
  if (
    !Array.isArray(components) ||
    components.length !== paths.length ||
    !paths.every((path) => components.some((component) => component["/"] === path)) ||
    components.some((component) => Object.keys(component).length !== 1)
  ) {
    throw new Error(
      "AASA debe cubrir /link/*, /join-team y /tournament/* sin condiciones adicionales.",
    );
  }
  const statement = android?.[0];
  const target = statement?.target;
  const fingerprints = target?.sha256_cert_fingerprints;
  if (
    !Array.isArray(android) ||
    android.length !== 1 ||
    statement.relation?.length !== 1 ||
    statement.relation[0] !== "delegate_permission/common.handle_all_urls" ||
    target?.namespace !== "android_app" ||
    target.package_name !== packageName ||
    !Array.isArray(fingerprints) ||
    fingerprints.length === 0 ||
    fingerprints.some((fingerprint) => !/^(?:[A-F0-9]{2}:){31}[A-F0-9]{2}$/.test(fingerprint)) ||
    new Set(fingerprints).size !== fingerprints.length ||
    statement.relation_extensions
  ) {
    throw new Error(
      "DAL debe asociar únicamente el paquete del entorno con huellas SHA-256 completas.",
    );
  }
}

export function prepareAssociations(environment, source, destination, required = false) {
  if (!packages[environment]) throw new Error("Entorno de asociaciones desconocido.");
  const names = ["apple-app-site-association", "assetlinks.json"];
  const present = names.map((name) => existsSync(resolve(source, name)));
  if (!present.some(Boolean) && !required) return false;
  if (!present.every(Boolean))
    throw new Error("Faltan asociaciones: se requieren ambos ficheros reales.");
  const contents = names.map((name) => readFileSync(resolve(source, name), "utf8"));
  validateAssociations(environment, ...contents.map((content) => JSON.parse(content)));
  if (destination) {
    const directory = resolve(destination, ".well-known");
    mkdirSync(directory, { recursive: true, mode: 0o755 });
    names.forEach((name, index) =>
      writeFileSync(resolve(directory, name), contents[index], { mode: 0o644 }),
    );
  }
  return true;
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  try {
    const [, , environment, source, destination] = process.argv;
    if (!source)
      throw new Error(
        "Uso: node infra/app-links/prepare.mjs <development|production> <origen> [release]",
      );
    const present = prepareAssociations(
      environment,
      source,
      destination,
      process.env.FASTTOURNEY_REQUIRE_APP_LINKS === "1",
    );
    console.log(
      present
        ? "Asociaciones validadas; comprobar su firma real antes de publicar."
        : "Release solo web: sin asociaciones móviles.",
    );
  } catch (error) {
    console.error(`Asociaciones rechazadas: ${error.message}`);
    process.exitCode = 1;
  }
}
