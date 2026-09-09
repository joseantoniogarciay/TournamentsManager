import { ScrollViewStyleReset, useServerDocumentContext } from "expo-router/html";
import type { PropsWithChildren } from "react";

import { themeCanvasColors, themePreferenceStorageKey } from "@/shared/preferences/theme";

const initialThemeScript = `(() => {
  try {
    const storedTheme = window.localStorage.getItem(${JSON.stringify(themePreferenceStorageKey)});
    const systemTheme = window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
    const resolvedTheme = storedTheme === "dark" || storedTheme === "light" ? storedTheme : systemTheme;
    const backgroundColor = resolvedTheme === "dark" ? ${JSON.stringify(themeCanvasColors.dark)} : ${JSON.stringify(themeCanvasColors.light)};
    const root = document.documentElement;
    root.dataset.theme = resolvedTheme;
    root.style.backgroundColor = backgroundColor;
    root.style.colorScheme = resolvedTheme;
    root.style.setProperty("--initial-canvas-color", backgroundColor);
    document.getElementById("initial-theme-color")?.setAttribute("content", backgroundColor);
  } catch {}
})();`;

export default function RootHtml({ children }: PropsWithChildren) {
  const { bodyAttributes, bodyNodes, headNodes, htmlAttributes } = useServerDocumentContext();

  return (
    <html {...htmlAttributes} lang="en">
      <head>
        <meta charSet="utf-8" />
        <meta id="initial-theme-color" name="theme-color" content={themeCanvasColors.light} />
        <style>{`
          :root {
            --initial-canvas-color: ${themeCanvasColors.light};
            background-color: var(--initial-canvas-color);
            color-scheme: light;
          }
          @media (prefers-color-scheme: dark) {
            :root {
              --initial-canvas-color: ${themeCanvasColors.dark};
              color-scheme: dark;
            }
          }
          body {
            background-color: var(--initial-canvas-color);
          }
        `}</style>
        <script dangerouslySetInnerHTML={{ __html: initialThemeScript }} />
        <ScrollViewStyleReset />
        {headNodes}
      </head>
      <body {...bodyAttributes}>
        {children}
        {bodyNodes}
      </body>
    </html>
  );
}
