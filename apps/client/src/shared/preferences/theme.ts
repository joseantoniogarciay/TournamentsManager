export type ThemePreference = "system" | "light" | "dark";
export type ResolvedTheme = "light" | "dark";

export const themePreferenceStorageKey = "tournaments-manager.theme-preference";

export const themeCanvasColors = {
  light: "#F8FAFC",
  dark: "#101828",
} satisfies Record<ResolvedTheme, string>;

export function getStoredWebThemePreference(): ThemePreference {
  try {
    const storedTheme = window.localStorage.getItem(themePreferenceStorageKey);
    return isThemePreference(storedTheme) ? storedTheme : "system";
  } catch {
    return "system";
  }
}

export function isThemePreference(value: string | null): value is ThemePreference {
  return value === "system" || value === "light" || value === "dark";
}
