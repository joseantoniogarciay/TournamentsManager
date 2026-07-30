import { apiFetch } from "@/api/fetch";
import { getUsernameAvailability } from "@/api/generated/usernames/usernames";

/** Adapta el cliente OpenAPI a las necesidades de la feature de registro. */
export function getRegistrationUsernameAvailability(username: string, signal: AbortSignal) {
  return getUsernameAvailability(username, { signal }, apiFetch);
}
