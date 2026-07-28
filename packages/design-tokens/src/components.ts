// Contratos de estado que web y nativo deberán representar de la misma manera.
export type FieldState = "default" | "focus" | "filled" | "error" | "disabled";
export type ButtonVariant = "primary" | "secondary" | "ghost" | "destructive";
export type BannerKind = "network-error" | "generic-error" | "success";

export const feedbackCopy = {
  networkError: "No hemos podido conectarnos. Revisa tu conexión e inténtalo de nuevo.",
  genericError: "Estamos teniendo problemas. Lo sentimos, inténtalo más tarde.",
} as const;

export const banner = { autoDismissMs: 6000, swipeToDismiss: true } as const;
