/**
 * @typedef {{ name: string, version: string, dependencies?: Record<string, string>,
 *   peerDependencies?: Record<string, string> }} PackageManifest
 */

export const hooks = {
  /** @param {PackageManifest} pkg */
  readPackage(pkg) {
    // ADR-0171: this optional Orval tool needs the TS 6 API, rather than the
    // project's native TS 7 compiler. Remove the peer to prevent root binding.
    if (pkg.name === "typedoc" && pkg.version === "0.28.20") {
      const peerDependencies = { ...pkg.peerDependencies };
      delete peerDependencies.typescript;
      return {
        ...pkg,
        peerDependencies,
        dependencies: { ...pkg.dependencies, typescript: "6.0.3" },
      };
    }
    return pkg;
  },
};
