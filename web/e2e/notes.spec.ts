import { expect, test } from "@playwright/test"

test("serves the version it was built as", async ({ page, request }) => {
  const health = await (await request.get("/api/health")).json()
  expect(health.status).toBe("ok")
  if (process.env.EXPECT_VERSION) {
    expect(health.version).toBe(process.env.EXPECT_VERSION)
  }

  await page.goto("/")
  await expect(page.getByText(`v${health.version}`)).toBeVisible()
})

test("adds and deletes a note", async ({ page }) => {
  const text = `e2e note ${Date.now()}`
  await page.goto("/")

  await page.getByLabel("New note").fill(text)
  await page.getByRole("button", { name: "Add" }).click()
  await expect(page.getByText(text)).toBeVisible()

  await page.reload()
  await expect(page.getByText(text)).toBeVisible()

  await page.getByRole("button", { name: `Delete "${text}"` }).click()
  await expect(page.getByText(text)).toHaveCount(0)
})

test("serves the app for client-side routes", async ({ page }) => {
  await page.goto("/some/client/route")
  await expect(page.getByLabel("New note")).toBeVisible()
})
