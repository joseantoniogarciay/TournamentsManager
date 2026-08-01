let pendingToken: string | null = null;
export function setPendingPasswordResetToken(token: string) {
  pendingToken = token;
}
export function getPendingPasswordResetToken() {
  return pendingToken;
}
