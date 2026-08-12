/* global Scalar */

Scalar.createApiReference("#app", {
  url: "/api-docs/openapi.yaml",
  servers: [
    {
      url: "https://dev-api.fasttourney.com/v1",
      description: "API pública de desarrollo",
    },
  ],
  telemetry: false,
});
