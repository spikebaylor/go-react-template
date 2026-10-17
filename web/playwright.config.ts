import { defineConfig, devices } from "@playwright/test"

// End-to-end tests run against a running build of the app, never a dev server:
// `just e2e` starts the combined image with compose and points BASE_URL at it.
export default defineConfig({
  testDir: "./e2e",
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  reporter: process.env.CI
    ? [["list"], ["junit", { outputFile: "../out/reports/e2e-junit.xml" }]]
    : [["list"]],
  use: {
    baseURL: process.env.BASE_URL ?? "http://localhost:8080",
    trace: "retain-on-failure",
  },
  projects: [{ name: "chromium", use: { ...devices["Desktop Chrome"] } }],
})
