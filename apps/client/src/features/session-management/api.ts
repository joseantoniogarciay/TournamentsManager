import { APIUnexpectedResponseError, authenticatedApiFetch } from "@/api/fetch";
import {
  listCurrentAccountSessions,
  revokeCurrentAccountSession,
  revokeOtherCurrentAccountSessions,
} from "@/api/generated/session/session";
import type { ActiveSession, ActiveSessionPage } from "@/api/generated/models";
import { ActiveSessionClient } from "@/api/generated/models/activeSessionClient";
import { ActiveSessionPlatform } from "@/api/generated/models/activeSessionPlatform";

const isRecord = (value: unknown): value is Record<string, unknown> =>
  !!value && typeof value === "object" && !Array.isArray(value);
const isUUID = (value: unknown): value is string =>
  typeof value === "string" &&
  /^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(value);
const isDate = (value: unknown): value is string =>
  typeof value === "string" && Number.isFinite(Date.parse(value));

/** Reject the whole page if a session cannot be safely displayed or targeted. */
export function parseSessionPage(value: unknown): ActiveSessionPage | null {
  if (
    !isRecord(value) ||
    !Array.isArray(value.items) ||
    value.items.length > 50 ||
    !(value.nextCursor === null || isUUID(value.nextCursor))
  )
    return null;
  const ids = new Set<string>();
  const items: ActiveSession[] = [];
  let currentCount = 0;
  for (const item of value.items) {
    if (
      !isRecord(item) ||
      !isUUID(item.id) ||
      !isDate(item.createdAt) ||
      !isDate(item.lastSeenAt) ||
      Date.parse(item.lastSeenAt) < Date.parse(item.createdAt) ||
      typeof item.current !== "boolean" ||
      !Object.values(ActiveSessionClient).includes(item.client as ActiveSessionClient) ||
      !Object.values(ActiveSessionPlatform).includes(item.platform as ActiveSessionPlatform) ||
      ids.has(item.id)
    )
      return null;
    ids.add(item.id);
    if (item.current && ++currentCount > 1) return null;
    items.push({
      id: item.id,
      createdAt: item.createdAt,
      lastSeenAt: item.lastSeenAt,
      current: item.current,
      client: item.client as ActiveSessionClient,
      platform: item.platform as ActiveSessionPlatform,
    });
  }
  if (
    value.nextCursor !== null &&
    (value.items.length !== 50 || value.nextCursor !== value.items.at(-1)?.id)
  )
    return null;
  return { items, nextCursor: value.nextCursor };
}

export async function listSessions(cursor?: string, signal?: AbortSignal) {
  const response = await listCurrentAccountSessions(
    cursor ? { cursor } : undefined,
    { signal },
    authenticatedApiFetch,
  );
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  const page = parseSessionPage(response.data);
  if (!page) throw new APIUnexpectedResponseError(response.status);
  return page;
}

export async function revokeSession(id: string, signal?: AbortSignal) {
  const response = await revokeCurrentAccountSession(id, { signal }, authenticatedApiFetch);
  if (response.status !== 204) throw new APIUnexpectedResponseError(response.status);
}

export async function revokeOthers(signal?: AbortSignal) {
  const response = await revokeOtherCurrentAccountSessions({ signal }, authenticatedApiFetch);
  if (response.status !== 204) throw new APIUnexpectedResponseError(response.status);
}
