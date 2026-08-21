import type { Notification } from "@/api/generated/models/notification";
import { NotificationKind } from "@/api/generated/models/notificationKind";

type RecordValue = Record<string, unknown>;

function isRecord(value: unknown): value is RecordValue {
  return !!value && typeof value === "object" && !Array.isArray(value);
}

function isUUID(value: unknown): value is string {
  return (
    typeof value === "string" &&
    /^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(value)
  );
}

function isDateTime(value: unknown): value is string {
  return typeof value === "string" && Number.isFinite(Date.parse(value));
}

function isNotificationKind(value: unknown): value is Notification["kind"] {
  return Object.values(NotificationKind).includes(value as Notification["kind"]);
}

/**
 * Construye una notificación solo tras comprobar el contrato en runtime. Un
 * elemento defectuoso de una colección nunca alcanza la interfaz.
 */
function parseNotification(value: unknown): Notification | null {
  if (!isRecord(value)) return null;
  if (
    !isUUID(value.id) ||
    !isNotificationKind(value.kind) ||
    !isUUID(value.leagueId) ||
    typeof value.leagueName !== "string" ||
    !isDateTime(value.createdAt) ||
    !(value.readAt === null || isDateTime(value.readAt))
  ) {
    return null;
  }

  return {
    id: value.id,
    kind: value.kind,
    leagueId: value.leagueId,
    leagueName: value.leagueName,
    createdAt: value.createdAt,
    readAt: value.readAt,
  };
}

/** Devuelve `null` cuando el contenedor no cumple el contrato. */
export function parseNotificationItems(value: unknown): Notification[] | null {
  if (!isRecord(value) || !Array.isArray(value.items)) return null;
  return value.items.flatMap((item) => {
    const notification = parseNotification(item);
    return notification ? [notification] : [];
  });
}
