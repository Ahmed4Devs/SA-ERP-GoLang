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
    await page.getByRole("button", { name: /إضافة سجل|create record/i }).click();
    await page.locator("#create-code").fill(code);
    await page.locator("#create-name_en").fill("Real E2E Category");
    await page.locator("#create-name_ar").fill("فئة تكامل فعلية");
    await page.getByRole("button", { name: /^(حفظ|save)$/i }).click();

    await expect(page.getByRole("status")).toContainText(/تم حفظ|saved successfully/i);
    await expect(page.getByText(code, { exact: true })).toBeVisible();
  });

  test("administrator creates an account and assigns an entity role", async ({ page }) => {
    const username = `e2e.user.${Date.now()}`;
    await page.addInitScript(() => localStorage.setItem("erp_locale", "en"));
    await page.goto("/login");
    await page.locator("#login-identifier").fill(process.env.E2E_USERNAME || "admin");
    await page.locator("#login-password").fill(process.env.E2E_PASSWORD || "Admin@ERP123!");
    await page.locator("#login-submit-btn").click();
    await expect(page).toHaveURL(/\/$/);
    await page.goto("/authorization");
    await page.getByRole("button", { name: "Create user", exact: true }).click();
    await page.locator("#user-username").fill(username);
    await page.locator("#user-display_name").fill(username);
    await page.locator("#user-password").fill("E2E-Temporary-Password!");
    await page.getByRole("button", { name: "Save", exact: true }).click();
    await expect(page.getByRole("status")).toContainText("User created");
    await page.getByRole("button", { name: `Manage: ${username}` }).click();
    const roles = page.getByLabel("Role", { exact: true });
    await expect(roles.locator("option")).not.toHaveCount(1);
    await roles.selectOption({ index: 1 });
    await page.getByRole("button", { name: "Assign role", exact: true }).click();
    await expect(page.getByRole("status")).toContainText("Changes saved");
    await page.getByLabel("Status", { exact: true }).selectOption("SUSPENDED");
    await page.getByRole("button", { name: "Update status", exact: true }).click();
    await expect(page.getByRole("row").filter({ has: page.getByRole("button", { name: `Manage: ${username}` }) })).toContainText("Suspended");
    await page.getByLabel("New password", { exact: true }).fill("E2E-Changed-Password!");
    await page.getByRole("button", { name: "Reset password", exact: true }).click();
    await expect(page.getByLabel("New password", { exact: true })).toHaveValue("");
  });

  test("document upload persists and downloads the original content", async ({ page }) => {
    const title = `E2E document ${Date.now()}`;
    await page.addInitScript(() => localStorage.setItem("erp_locale", "en"));
    await page.goto("/login");
    await page.locator("#login-identifier").fill(process.env.E2E_USERNAME || "admin");
    await page.locator("#login-password").fill(process.env.E2E_PASSWORD || "Admin@ERP123!");
    await page.locator("#login-submit-btn").click();
    await expect(page).toHaveURL(/\/$/);
    await page.goto("/documents");
    await page.getByRole("button", { name: "Upload document", exact: true }).click();
    await page.getByLabel("File", { exact: true }).setInputFiles({ name: "integration.txt", mimeType: "text/plain", buffer: Buffer.from("ERP integration document") });
    await page.getByLabel("Title", { exact: true }).fill(title);
    await page.getByRole("button", { name: "Upload", exact: true }).click();
    const row = page.getByRole("row").filter({ has: page.getByRole("cell", { name: title, exact: true }) });
    await expect(row).toBeVisible();
    const downloadEvent = page.waitForEvent("download");
    await row.getByRole("button", { name: "Download", exact: true }).click();
    const download = await downloadEvent;
    expect(download.suggestedFilename()).toBe("integration.txt");
    const stream = await download.createReadStream();
    const chunks: Buffer[] = [];
    for await (const chunk of stream) chunks.push(Buffer.from(chunk));
    expect(Buffer.concat(chunks).toString()).toBe("ERP integration document");
  });
});
