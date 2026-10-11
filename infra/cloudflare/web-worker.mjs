import { noindex, securityHeaders } from "./web-headers.mjs";

const tournamentPath =
  /^\/tournament\/([0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12})$/i;
const publicAPI = "https://api.fasttourney.com/v1";
const publicWeb = "https://fasttourney.com";

function safeResponse(message, status, method = "GET") {
  return new globalThis.Response(method === "HEAD" ? null : message, {
    status,
    headers: {
      ...securityHeaders,
      "Content-Type": "text/plain; charset=utf-8",
      "Cache-Control": "no-store",
      "X-Robots-Tag": noindex,
      ...(status === 405 ? { Allow: "GET, HEAD" } : {}),
    },
  });
}

function escapeHTML(value) {
  return value.replace(
    /[&<>"']/g,
    (character) =>
      ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[character],
  );
}

export function previewHead(tournament, path) {
  const characters = [...tournament.name.trim()];
  const name =
    characters.length > 120 ? `${characters.slice(0, 119).join("")}…` : characters.join("");
  const title = `${name} | FastTourney`;
  const state =
    {
      published: "Published",
      in_progress: "In progress",
      completed: "Completed",
      cancelled: "Cancelled",
    }[tournament.state] ?? "Available";
  // Preserve the existing renderer's metadata during the hosting migration.
  const description = `Football league · ${tournament.teams.length} ${tournament.teams.length === 1 ? "team" : "teams"} · ${state} on FastTourney.`;
  const image = `${publicWeb}/fasttourney-league-preview.png`;
  const canonical = `${publicWeb}${path}`;
  const tags = [
    ["name", "description", description],
    ["name", "robots", noindex],
    ["property", "og:type", "website"],
    ["property", "og:title", title],
    ["property", "og:description", description],
    ["property", "og:url", canonical],
    ["property", "og:image", image],
    ["property", "og:image:width", "1200"],
    ["property", "og:image:height", "630"],
    ["property", "og:image:alt", "FastTourney sports tournament"],
    ["name", "twitter:card", "summary_large_image"],
    ["name", "twitter:title", title],
    ["name", "twitter:description", description],
    ["name", "twitter:image", image],
  ];
  return [
    `<title>${escapeHTML(title)}</title>`,
    `<link rel="canonical" href="${escapeHTML(canonical)}">`,
    ...tags.map(
      ([attribute, key, value]) => `<meta ${attribute}="${key}" content="${escapeHTML(value)}">`,
    ),
  ].join("");
}

async function renderPreview(request, env, path, id) {
  try {
    // Infrastructure adapter for the public OpenAPI projection. Never forward
    // incoming cookies, credentials, query strings or visitor-supplied headers.
    const apiResponse = await globalThis.fetch(`${publicAPI}/tournaments/${id.toLowerCase()}`, {
      signal: globalThis.AbortSignal.timeout(3000),
      redirect: "manual",
      headers: { Accept: "application/json" },
    });
    if (apiResponse.status === 404) return safeResponse("Not found", 404, request.method);
    if (apiResponse.status !== 200)
      return safeResponse("Tournament preview is temporarily unavailable", 503, request.method);
    if (!apiResponse.headers.get("Content-Type")?.includes("application/json"))
      throw new Error("Invalid projection");
    const reader = apiResponse.body.getReader();
    const chunks = [];
    let size = 0;
    try {
      for (;;) {
        const { value, done } = await reader.read();
        if (done) break;
        size += value.byteLength;
        if (size > 1 << 20) throw new Error("Projection too large");
        chunks.push(value);
      }
    } finally {
      await reader.cancel();
    }
    const bytes = new Uint8Array(size);
    let offset = 0;
    for (const chunk of chunks) {
      bytes.set(chunk, offset);
      offset += chunk.byteLength;
    }
    const tournament = JSON.parse(new globalThis.TextDecoder().decode(bytes));
    if (
      typeof tournament.name !== "string" ||
      typeof tournament.state !== "string" ||
      !Array.isArray(tournament.teams)
    )
      throw new Error("Invalid projection");
    const shell = await env.ASSETS.fetch(
      new globalThis.Request(`${publicWeb}/`, { method: "GET" }),
    );
    if (shell.status !== 200 || !shell.headers.get("Content-Type")?.includes("text/html"))
      throw new Error("Missing shell");
    const remove = {
      element(element) {
        element.remove();
      },
    };
    const rewritten = new globalThis.HTMLRewriter()
      .on("head title", remove)
      .on('head link[rel="canonical"]', remove)
      .on(
        'head meta[name="description"], head meta[name="robots"], head meta[property^="og:"], head meta[name^="twitter:"]',
        remove,
      )
      .on("head", {
        element(element) {
          element.append(previewHead(tournament, path), { html: true });
        },
      })
      .transform(shell);
    return new globalThis.Response(request.method === "HEAD" ? null : rewritten.body, {
      headers: {
        ...securityHeaders,
        "Content-Type": "text/html; charset=utf-8",
        "Cache-Control": "no-store",
        "X-Robots-Tag": noindex,
      },
    });
  } catch {
    return safeResponse("Tournament preview is temporarily unavailable", 503, request.method);
  }
}

export default {
  async fetch(request, env) {
    const url = new globalThis.URL(request.url);
    if (url.hostname === "www.fasttourney.com") {
      return new globalThis.Response(null, {
        status: 308,
        headers: { ...securityHeaders, Location: `${publicWeb}${url.pathname}${url.search}` },
      });
    }
    if (request.method !== "GET" && request.method !== "HEAD")
      return safeResponse("Method not allowed", 405, request.method);
    const match = tournamentPath.exec(url.pathname);
    if (match) return renderPreview(request, env, url.pathname, match[1]);
    // Assets run first. Missing associations and asset files must not masquerade
    // as successful HTML. Deep application routes retain the Expo SPA shell.
    if (
      url.pathname.startsWith("/.well-known/") ||
      /\.[a-z0-9]+$/i.test(url.pathname) ||
      url.pathname.startsWith("/_expo/") ||
      url.pathname.startsWith("/assets/")
    )
      return safeResponse("Not found", 404, request.method);
    const shell = await env.ASSETS.fetch(
      new globalThis.Request(`${publicWeb}/`, { method: request.method }),
    );
    const headers = new globalThis.Headers(shell.headers);
    for (const [key, value] of Object.entries(securityHeaders)) headers.set(key, value);
    // Keep the canonical home indexable even when the edge asset service
    // supplies a noindex header. Preview hosts and SPA fallbacks stay private.
    if (url.hostname === "fasttourney.com" && url.pathname === "/") headers.delete("X-Robots-Tag");
    else headers.set("X-Robots-Tag", noindex);
    return new globalThis.Response(shell.body, { status: shell.status, headers });
  },
};
