// Fundaciones agnósticas de plataforma para la dirección visual Pulse.
export const color = {
  brand: { primary: "#155EEF", primaryPressed: "#004EEB", accent: "#7F56D9" },
  surface: { canvas: "#F8FAFC", default: "#FFFFFF", subtle: "#F1F5F9", inverse: "#101828" },
  text: { primary: "#101828", secondary: "#475467", placeholder: "#98A2B3", inverse: "#FFFFFF" },
  border: { default: "#D0D5DD", focus: "#155EEF", error: "#D92D20" },
  feedback: { success: "#027A48", warning: "#B54708", error: "#D92D20", info: "#155EEF" },
} as const;

export const space = {
  0: 0,
  1: 4,
  2: 8,
  3: 12,
  4: 16,
  5: 20,
  6: 24,
  8: 32,
  10: 40,
  12: 48,
} as const;

export const radius = { control: 12, card: 16, pill: 999 } as const;

export const typography = {
  family: { system: "system-ui" },
  size: { caption: 12, body: 14, bodyLarge: 16, title: 20, display: 32 },
  weight: { regular: "400", medium: "500", semibold: "600", bold: "700" },
  lineHeight: { compact: 1.2, default: 1.5 },
} as const;

export const motion = { feedback: 160, enterExit: 240 } as const;

export const control = { minHeight: 44, horizontalPadding: 16, iconSize: 20 } as const;
