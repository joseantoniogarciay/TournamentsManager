import { defineConfig } from "orval";

export default defineConfig({
  tournamentsManager: {
    input: {
      target: "./contracts/openapi/v1/openapi.yaml",
    },
    output: {
      target: "./apps/client/src/api/generated/endpoints.ts",
      schemas: "./apps/client/src/api/generated/models",
      client: "fetch",
      mode: "tags-split",
      clean: true,
      formatter: "prettier",
      indexFiles: true,
      tagsSplitDeduplication: true,
      override: {
        fetch: {
          useRuntimeFetcher: true,
        },
      },
    },
  },
});
