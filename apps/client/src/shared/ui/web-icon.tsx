import { createElement } from "react";

type IconName = "account" | "add" | "back" | "close" | "home" | "more" | "settings" | "tournament";

type WebIconProps = {
  color: string;
  name: IconName;
  size: number;
};

const paths: Record<IconName, readonly string[]> = {
  account: ["M20 21a8 8 0 0 0-16 0", "M12 13a4 4 0 1 0 0-8 4 4 0 0 0 0 8Z"],
  add: ["M12 5v14", "M5 12h14"],
  back: ["m15 18-6-6 6-6"],
  close: ["m18 6-12 12", "m6 6 12 12"],
  home: ["m3 10 9-7 9 7v10a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1V10Z", "M9 21v-6h6v6"],
  more: ["M5 12h.01", "M12 12h.01", "M19 12h.01"],
  settings: [
    "M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6Z",
    "M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 1 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.6 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 1 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.6h.01a1.65 1.65 0 0 0 1-1.51V3a2 2 0 1 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9v.01a1.65 1.65 0 0 0 1.51 1H21a2 2 0 1 1 0 4h-.09A1.65 1.65 0 0 0 19.4 15Z",
  ],
  tournament: [
    "M7 4h10v4a5 5 0 0 1-10 0V4Z",
    "M7 6H4v1a4 4 0 0 0 4 4M17 6h3v1a4 4 0 0 1-4 4M12 13v5M8 21h8M9 18h6",
  ],
};

export function WebIcon({ color, name, size }: WebIconProps) {
  return createElement(
    "svg",
    {
      "aria-hidden": true,
      fill: "none",
      focusable: "false",
      height: size,
      stroke: color,
      strokeLinecap: "round",
      strokeLinejoin: "round",
      strokeWidth: 1.8,
      viewBox: "0 0 24 24",
      width: size,
    },
    paths[name].map((path, index) => createElement("path", { d: path, key: index })),
  );
}
