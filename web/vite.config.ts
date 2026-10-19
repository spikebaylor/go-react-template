/// <reference types="vitest/config" />
import tailwindcss from "@tailwindcss/vite"
import react from "@vitejs/plugin-react"
import { fileURLToPath } from "node:url"
import { defineConfig } from "vite"

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: { "@": fileURLToPath(new URL("./src", import.meta.url)) },
  },
  server: {
    // `just dev-web` runs Vite here; `just dev-api` runs the Go server on :8080.
    proxy: { "/api": "http://localhost:8080" },
  },
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: ["./src/test-setup.ts"],
    include: ["src/**/*.test.{ts,tsx}"],
    reporters: process.env.CI ? ["default", "junit"] : ["default"],
    outputFile: { junit: "../out/reports/web-junit.xml" },
    // The web coverage floor `just verify` enforces. README.md's "Test
    // coverage" section explains the numbers and the exclusions.
    coverage: {
      provider: "v8",
      include: ["src/**/*.{ts,tsx}"],
      exclude: [
        "src/**/*.test.{ts,tsx}",
        "src/test-setup.ts",
        // Copied in by `npx shadcn add`; not code we write.
        "src/components/ui/**",
        // Generated from api/openapi.yaml by `just generate`.
        "src/api/schema.d.ts",
        // Only mounts <App />; `just e2e` runs it in a real browser.
        "src/main.tsx",
      ],
      reporter: ["text-summary", "html", "lcov"],
      reportsDirectory: "../out/reports/web-coverage",
      thresholds: { statements: 60, branches: 60, functions: 60, lines: 60 },
    },
  },
})
