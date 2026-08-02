import { APIUnexpectedResponseError, apiFetch } from "@/api/fetch";
import { getCurrentSession } from "@/api/generated/session/session";

/** Consulta la cookie HttpOnly web al arrancar; una sesión ausente es anónima. */
export async function restoreWebSession() {
  const response = await getCurrentSession(undefined, apiFetch);
  if (response.status === 200) return response.data.user;
  if (response.status === 401) return null;
  throw new APIUnexpectedResponseError((response as { status: number }).status);
}
