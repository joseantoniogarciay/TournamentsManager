import type { TipPaymentLink } from "./tip-payment-link";

export type { TipPaymentLink } from "./tip-payment-link";

const configuredLinks: readonly TipPaymentLink[] = [
  { amount: 2, url: process.env.EXPO_PUBLIC_TIP_PAYMENT_LINK_2_EUR ?? "" },
  { amount: 5, url: process.env.EXPO_PUBLIC_TIP_PAYMENT_LINK_5_EUR ?? "" },
  { amount: 10, url: process.env.EXPO_PUBLIC_TIP_PAYMENT_LINK_10_EUR ?? "" },
];

export function getTipPaymentLinks(): TipPaymentLink[] | undefined {
  if (!configuredLinks.every(({ url }) => isHTTPSURL(url))) return undefined;
  return [...configuredLinks];
}

function isHTTPSURL(value: string) {
  try {
    return new URL(value).protocol === "https:";
  } catch {
    return false;
  }
}
