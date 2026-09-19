import type { TipPaymentLink } from "./tip-payment-link";

export type { TipPaymentLink } from "./tip-payment-link";

// ADR-0129 prohíbe enlazar a Stripe u otro pago externo desde las apps.
export function getTipPaymentLinks(): TipPaymentLink[] | undefined {
  return undefined;
}
