import { defineConfig } from "orval"

export default defineConfig({
  pcc: {
    input: "./openapi.json",
    output: {
      target: "./src/lib/orval/endpoints.ts",
      schemas: "./src/lib/orval/schemas",
      client: "react-query",
      mode: "tags-split",
      mock: false,
      prettier: true,
    },
    hooks: {
      afterAllFilesWrite: "npx prettier --write",
    },
  },
})
