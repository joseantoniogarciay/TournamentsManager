import type { WebBrowserCustomTabsResults } from "expo-web-browser";

export function customTabsBrowserOptions(browsers: WebBrowserCustomTabsResults) {
  const browserPackage = [
    browsers.preferredBrowserPackage,
    browsers.defaultBrowserPackage,
    ...browsers.browserPackages,
  ].find(
    (candidate) =>
      candidate &&
      browsers.browserPackages.includes(candidate) &&
      browsers.servicePackages.includes(candidate),
  );
  return browserPackage ? { browserPackage } : {};
}
