import { APIUnexpectedResponseError, authenticatedApiFetch } from "@/api/fetch";
import {
  deleteAllCurrentAccountNotifications,
  deleteCurrentAccountNotification,
  getCurrentAccountUnreadNotificationCount,
  listCurrentAccountNotifications,
  markAllCurrentAccountNotificationsRead,
} from "@/api/generated/notifications/notifications";
import { parseNotificationItems } from "./response-parser";

export async function listNotifications() {
  const response = await listCurrentAccountNotifications(undefined, authenticatedApiFetch);
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  const items = parseNotificationItems(response.data);
  if (!items) throw new APIUnexpectedResponseError(response.status);
  return items;
}
export async function unreadNotificationCount() {
  const response = await getCurrentAccountUnreadNotificationCount(undefined, authenticatedApiFetch);
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  return response.data.count;
}
export async function markAllNotificationsRead() {
  const response = await markAllCurrentAccountNotificationsRead(undefined, authenticatedApiFetch);
  if (response.status !== 204) throw new APIUnexpectedResponseError(response.status);
}
export async function deleteNotification(id: string) {
  const response = await deleteCurrentAccountNotification(id, undefined, authenticatedApiFetch);
  if (response.status !== 204) throw new APIUnexpectedResponseError(response.status);
}
export async function deleteAllNotifications() {
  const response = await deleteAllCurrentAccountNotifications(undefined, authenticatedApiFetch);
  if (response.status !== 204) throw new APIUnexpectedResponseError(response.status);
}
