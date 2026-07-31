import { APIConnectionError } from "@/api/fetch";

import type { TranslationKey } from "@/shared/i18n/locale";

export type RequestFailureKind = "network-error" | "generic-error";

/**
 * Clasifica únicamente fallos que el cliente puede explicar con seguridad.
 * Las reglas de negocio siguen siendo responsabilidad de cada feature.
 */
export function getRequestFailure(error: unknown): {
  kind: RequestFailureKind;
  messageKey: TranslationKey;
} {
  if (error instanceof APIConnectionError) {
    return { kind: "network-error", messageKey: "common_network_error" };
  }

  return { kind: "generic-error", messageKey: "common_request_error" };
}
