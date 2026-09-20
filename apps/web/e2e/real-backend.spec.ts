import { expect, test } from "@playwright/test";

test.describe("real backend", () => {
  test.skip(process.env.E2E_REAL_BACKEND !== "1", "Set E2E_REAL_BACKEND=1 to run against the Go API and PostgreSQL.");

  test("login, select a tenant, create a category, and read it back", async ({ page }) => {
    const code = `E2E-${Date.now()}`;
    await page.goto("/login");
    await page.locator("#login-identifier").fill(process.env.E2E_USERNAME || "admin");
    await page.locator("#login-password").fill(process.env.E2E_PASSWORD || "Admin@ERP123!");
    await page.locator("#login-submit-btn").click();
    await expect(page).toHaveURL(/\/$/);

    await page.goto("/products");
    await page.getByRole("button", { name: /الفئات|categories/i }).click();
    await page.getByRole("button", { name: /عملية جديدة|new operation/i }).click();
    await page.locator("#operation-body").fill(JSON.stringify({ code, name_en: "Real E2E Category", name_ar: "فئة تكامل فعلية" }));
    await page.locator("#operation-submit").click();

    await expect(page.getByTestId("operation-success")).toBeVisible();
    await expect(page.getByText(code, { exact: true })).toBeVisible();
  });
});
