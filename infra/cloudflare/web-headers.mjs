export const securityHeaders = {
  "Strict-Transport-Security": "max-age=31536000",
  "X-Content-Type-Options": "nosniff",
  "Referrer-Policy": "strict-origin-when-cross-origin",
  "Permissions-Policy": "camera=(), geolocation=(), microphone=(), payment=(), usb=()",
  "Content-Security-Policy":
    "default-src 'self'; base-uri 'self'; object-src 'none'; frame-ancestors 'none'; form-action 'self'; script-src 'self' https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data: https://fonts.scalar.com; connect-src 'self' https://api.fasttourney.com https://accounts.google.com https://www.googleapis.com",
};

export const noindex = "noindex, nofollow, noarchive";

export function assetHeaders() {
  const common = Object.entries(securityHeaders).map(([key, value]) => `  ${key}: ${value}`);
  const indexable = ["/", "/apple-touch-icon.png", "/favicon.ico", "/robots.txt", "/sitemap.xml"];
  return [
    "/*",
    ...common,
    `  X-Robots-Tag: ${noindex}`,
    "",
    ...indexable.flatMap((path) => [path, "  ! X-Robots-Tag", ""]),
    "/.well-known/*",
    "  Content-Type: application/json",
    "",
    // A version preview must never become another indexed copy of the home.
    "https://:version.:account.workers.dev/*",
    `  X-Robots-Tag: ${noindex}`,
    "",
  ].join("\n");
}
