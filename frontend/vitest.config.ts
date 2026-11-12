import { defineConfig } from "vitest/config";
import path from "node:path";

export default defineConfig({
  test: {
    environment: "jsdom",
  },
  resolve: {
    // Mirrors the "@/*" path alias in tsconfig.json. Vitest does not read
    // tsconfig paths on its own, so without this the components importing
    // "@/lib/..." fail to resolve under test.
    alias: {
      "@": path.resolve(__dirname, "."),
    },
  },
});
