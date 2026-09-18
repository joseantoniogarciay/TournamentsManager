/* global document, window */

(() => {
  try {
    const root = document.documentElement;
    const storedTheme = window.localStorage.getItem(root.dataset.themeStorageKey);
    const systemTheme = window.matchMedia("(prefers-color-scheme: dark)").matches
      ? "dark"
      : "light";
    const resolvedTheme =
      storedTheme === "dark" || storedTheme === "light" ? storedTheme : systemTheme;
    const backgroundColor =
      resolvedTheme === "dark" ? root.dataset.darkCanvasColor : root.dataset.lightCanvasColor;

    if (!backgroundColor) return;

    root.dataset.theme = resolvedTheme;
    root.style.backgroundColor = backgroundColor;
    root.style.colorScheme = resolvedTheme;
    root.style.setProperty("--initial-canvas-color", backgroundColor);
    document.getElementById("initial-theme-color")?.setAttribute("content", backgroundColor);
  } catch {
    // El CSS conserva como fallback la preferencia del sistema si storage falla.
  }
})();
