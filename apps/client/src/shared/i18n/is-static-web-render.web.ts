export function isStaticWebRender() {
  return typeof window === "undefined";
}
