export function isShareCancellation(error: unknown) {
  return Boolean(
    error && typeof error === "object" && "name" in error && error.name === "AbortError",
  );
}
