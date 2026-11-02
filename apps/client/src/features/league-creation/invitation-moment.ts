// One transient, confirmed invitation outcome. Never serialized into a URL or storage.
let pending: { tournamentId: string; accountId: string } | undefined;

export function rememberInvitationMoment(tournamentId: string, accountId: string) {
  pending = { tournamentId, accountId };
}

export function takeInvitationMoment(tournamentId: string, accountId: string): boolean {
  if (!pending) return false;
  if (pending.accountId !== accountId) {
    pending = undefined;
    return false;
  }
  if (pending.tournamentId !== tournamentId) return false;
  pending = undefined;
  return true;
}
