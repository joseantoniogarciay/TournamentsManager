import * as WebBrowser from "expo-web-browser";
import { customTabsBrowserOptions } from "@/features/social-authentication/browser-options";

/** Prefer Android Custom Tabs; allow the system browser when none is available. */
export async function androidAuthBrowserOptions() {
  try {
    return customTabsBrowserOptions(await WebBrowser.getCustomTabsSupportingBrowsersAsync());
  } catch {
    // Package discovery must not prevent the system from resolving a browser.
    return {};
  }
}
