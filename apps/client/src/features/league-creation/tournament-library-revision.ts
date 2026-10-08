// Invalida la proyección local tras una mutación confirmada, sin persistir datos de cuenta.
let revision = 0;
export function getTournamentLibraryRevision() {
  return revision;
}
export function invalidateTournamentLibrary() {
  revision += 1;
}
