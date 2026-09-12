import { ScrollViewStyleReset, useServerDocumentContext } from "expo-router/html";
import type { PropsWithChildren } from "react";

import { themeCanvasColors, themePreferenceStorageKey } from "@/shared/preferences/theme";

export default function RootHtml({ children }: PropsWithChildren) {
  const { bodyAttributes, bodyNodes, headNodes, htmlAttributes } = useServerDocumentContext();

  return (
    <html
      {...htmlAttributes}
      lang="en"
      data-theme-storage-key={themePreferenceStorageKey}
      data-light-canvas-color={themeCanvasColors.light}
      data-dark-canvas-color={themeCanvasColors.dark}
    >
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
        <script src="/theme-init.js"></script>
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
