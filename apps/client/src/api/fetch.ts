const defaultAPIBaseURL = "http://127.0.0.1:8080/v1";

function getAPIBaseURL() {
  return (process.env.EXPO_PUBLIC_API_BASE_URL ?? defaultAPIBaseURL).replace(/\/$/, "");
}

function isAbsoluteURL(url: string) {
  return /^[a-z][a-z\d+.-]*:\/\//i.test(url);
}

/**
 * Transporte común del cliente OpenAPI.
 *
 * Las operaciones generadas aportan paths relativos; este borde añade el
 * origen configurable. Las credenciales de sesión se incorporarán aquí cuando
 * existan, sin obligar a cada feature a reconstruir la petición.
 */
export const apiFetch: typeof globalThis.fetch = (input, init) => {
  if (typeof input !== "string" || isAbsoluteURL(input)) {
    return globalThis.fetch(input, init);
  }

  const path = input.startsWith("/") ? input : `/${input}`;
  return globalThis.fetch(`${getAPIBaseURL()}${path}`, init);
};
