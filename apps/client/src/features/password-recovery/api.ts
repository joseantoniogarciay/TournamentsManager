import {
  captureProductOutcome,
  APIUnexpectedResponseError,
  apiFetch,
  saveMobileSession,
} from "@/api/fetch";
import {
  confirmPasswordReset,
  inspectPasswordResetLink,
  requestPasswordReset,
} from "@/api/generated/password-recovery/password-recovery";

export async function requestRecovery(email: string) {
  const response = await requestPasswordReset({ email }, undefined, apiFetch);
  if (response.status !== 202) throw new APIUnexpectedResponseError(response.status);
  captureProductOutcome("password_recovery_requested", response.headers);
}

export async function inspectRecovery(token: string) {
  const response = await inspectPasswordResetLink({ token }, undefined, apiFetch);
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  return response.data.email;
}

export async function confirmRecovery(
  token: string,
  password: string,
  sessionTransport: "bearer" | "cookie",
) {
  const response = await confirmPasswordReset(
    { token, password, sessionTransport },
    undefined,
    apiFetch,
  );
  if (response.status !== 200) throw new APIUnexpectedResponseError(response.status);
  if (sessionTransport === "bearer") await saveMobileSession(response.data);
  captureProductOutcome("password_recovery_completed", response.headers);
  return response.data;
}
