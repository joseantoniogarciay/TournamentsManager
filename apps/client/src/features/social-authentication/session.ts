import type { SessionEstablishment, Transport } from "@/api/generated/models";

/** A 200 status alone cannot turn malformed data into a client session. */
export function isSocialSession(
  value: unknown,
  transport: Transport,
): value is SessionEstablishment {
  if (!value || typeof value !== "object") return false;
  const session = value as Partial<SessionEstablishment>;
  return (
    session.delivery === transport &&
    typeof session.expiresAt === "string" &&
    Date.parse(session.expiresAt) > Date.now() &&
    typeof session.refreshExpiresAt === "string" &&
    Date.parse(session.refreshExpiresAt) > Date.now() &&
    Boolean(
      session.user &&
      typeof session.user.id === "string" &&
      session.user.id &&
      typeof session.user.username === "string" &&
      session.user.username,
    ) &&
    (transport === "cookie" ||
      (typeof session.sessionToken === "string" &&
        Boolean(session.sessionToken) &&
        typeof session.refreshToken === "string" &&
        Boolean(session.refreshToken)))
  );
}
