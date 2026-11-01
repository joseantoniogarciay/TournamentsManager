// Fundaciones agnósticas de plataforma para la identidad Grada (ADR-0176).
export const color = {
  brand: { primary: "#B93B19", primaryPressed: "#A33216", accent: "#A33216" },
  surface: { canvas: "#F5F1E8", default: "#FFFDF8", subtle: "#EAE4D8", inverse: "#24241F" },
  text: { primary: "#24241F", secondary: "#625D52", placeholder: "#6A6358", inverse: "#FFFFFF" },
  border: { default: "#827A6C", focus: "#B93B19", error: "#B42318" },
  feedback: { success: "#25623E", warning: "#854A0E", error: "#B42318", info: "#B93B19" },
} as const;

// El color de tinta/indicador cambia en oscuro; los fondos de marca conservan
// tinta blanca. No usar el tinte claro como fondo de una acción con texto blanco.
export const themeColors = {
  light: {
    surface: {
      canvas: color.surface.canvas,
      default: color.surface.default,
      subtle: color.surface.subtle,
    },
    text: color.text,
    indicator: { default: color.brand.primary },
    border: color.border,
    feedback: { success: color.feedback.success, error: color.feedback.error },
  },
  dark: {
    surface: { canvas: "#1C1B17", default: "#24231E", subtle: "#35332B" },
    text: { primary: "#F5F1E8", secondary: "#D4CCBD", placeholder: "#B4AA98", inverse: "#24241F" },
    indicator: { default: "#FF9B73" },
    border: { default: "#908778", focus: "#FF9B73", error: "#FFB4A4" },
    feedback: { success: "#91D4A6", error: "#FFB4A4" },
  },
} as const;

export const gradient = {
  brand: {
    colors: ["#A3321600", "#A3321600", color.brand.accent],
    end: { x: 1, y: 1 },
    locations: [0, 0.35, 1],
    start: { x: 0, y: 0 },
  },
  brandButton: {
    colors: ["#A3321600", "#A3321600", color.brand.accent],
    end: { x: 1, y: 1 },
    locations: [0, 0.35, 1],
    start: { x: 0, y: 0 },
  },
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

export const webLayout = {
  maxWidth: 1200,
  readingMaxWidth: 720,
  columnsMinWidth: 960,
  sidebarWidth: 360,
} as const;

export const typography = {
  family: {
    regular: "DMSans_400Regular",
    medium: "DMSans_500Medium",
    semibold: "DMSans_600SemiBold",
    bold: "DMSans_700Bold",
  },
  heading: { title: "BarlowCondensed_600SemiBold", display: "BarlowCondensed_700Bold" },
  numeric: { regular: "IBMPlexMono_400Regular", semibold: "IBMPlexMono_600SemiBold" },
  size: { caption: 12, body: 14, bodyLarge: 16, title: 20, display: 32 },
  lineHeight: { compact: 1.2, default: 1.5 },
} as const;

export const motion = { feedback: 160, enterExit: 240 } as const;

export const control = { minHeight: 44, horizontalPadding: 16, iconSize: 20 } as const;
