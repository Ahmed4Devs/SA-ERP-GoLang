import { expect, test, type Page, type Route } from "@playwright/test";

const user = {
  id: "10000000-0000-0000-0000-000000000001",
  username: "admin",
  email: "admin@example.sa",
  display_name: "System Administrator",
};

const entities = [
  {
    id: "20000000-0000-0000-0000-000000000001",
    role_code: "ADMIN",
    legal_entity_id: "30000000-0000-0000-0000-000000000001",
    legal_entity_name_ar: "الشركة الأولى",
    legal_entity_name_en: "First Company",
  },
  {
    id: "20000000-0000-0000-0000-000000000002",
    role_code: "ADMIN",
    legal_entity_id: "30000000-0000-0000-0000-000000000002",
    legal_entity_name_ar: "الشركة الثانية",
    legal_entity_name_en: "Second Company",
  },
];

async function json(route: Route, status: number, body: unknown, headers?: Record<string, string>) {
  await route.fulfill({ status, contentType: "application/json", body: JSON.stringify(body), headers });
}

async function mockAuthenticatedSession(page: Page) {
  await page.context().addCookies([
    { name: "erp_session", value: "browser-test-session", domain: "127.0.0.1", path: "/", httpOnly: true, sameSite: "Lax" },
  ]);
  await page.route("**/api/v1/auth/me", (route) => json(route, 200, { user }));
  await page.route("**/api/v1/authz/assignments/me", (route) => json(route, 200, { data: entities }));
}

test("a protected route redirects an anonymous user to login", async ({ page }) => {
  await page.route("**/api/v1/auth/me", (route) => json(route, 401, { error: { code: "UNAUTHORIZED", message: "Unauthorized" } }));

  await page.goto("/accounting");

  await expect(page).toHaveURL(/\/login\?next=%2Faccounting$/);
  await expect(page.locator("#login-identifier")).toBeVisible();
});

test("password login restores the requested route and exposes tenant selection", async ({ page }) => {
  let authenticated = false;
  await page.route("**/api/v1/auth/me", (route) =>
    json(route, authenticated ? 200 : 401, authenticated ? { user } : { error: { code: "UNAUTHORIZED", message: "Unauthorized" } }),
  );
  await page.route("**/api/v1/auth/login", async (route) => {
    authenticated = true;
    await json(route, 200, { user }, { "set-cookie": "erp_session=browser-test-session; Path=/; HttpOnly; SameSite=Lax" });
  });
  await page.route("**/api/v1/authz/assignments/me", (route) => json(route, 200, { data: entities }));
  await page.route("**/api/v1/accounting/**", (route) => json(route, 200, { data: [] }));

  await page.goto("/login?next=%2Faccounting");
  await page.locator("#login-identifier").fill("admin");
  await page.locator("#login-password").fill("correct-password");
  await page.locator("#login-submit-btn").click();

  await expect(page).toHaveURL(/\/accounting$/);
  const tenantSelector = page.getByLabel(/الكيان|legal entity/i);
  await expect(tenantSelector).toHaveValue(entities[0].legal_entity_id);
  await tenantSelector.selectOption(entities[1].legal_entity_id);
  await expect(tenantSelector).toHaveValue(entities[1].legal_entity_id);
});

test("an MFA challenge does not create a session until the second factor succeeds", async ({ page }) => {
  let authenticated = false;
  await page.route("**/api/v1/auth/me", (route) =>
    json(route, authenticated ? 200 : 401, authenticated ? { user } : { error: { code: "UNAUTHORIZED", message: "Unauthorized" } }),
  );
  await page.route("**/api/v1/auth/login", (route) =>
    json(route, 202, { mfa_required: true, challenge_token: "challenge-token", expires_at: "2030-01-01T00:00:00Z" }),
  );
  await page.route("**/api/v1/auth/login/mfa", async (route) => {
    authenticated = true;
    await json(route, 200, { user }, { "set-cookie": "erp_session=browser-test-session; Path=/; HttpOnly; SameSite=Lax" });
  });
  await page.route("**/api/v1/authz/assignments/me", (route) => json(route, 200, { data: entities }));
  await page.route("**/api/v1/dashboard**", (route) => json(route, 503, { error: "dashboard unavailable" }));

  await page.goto("/login");
  await page.locator("#login-identifier").fill("admin");
  await page.locator("#login-password").fill("correct-password");
  await page.locator("#login-submit-btn").click();

  await expect(page.locator("#login-mfa-code")).toBeVisible();
  await expect(page).toHaveURL(/\/login$/);
  await page.locator("#login-mfa-code").fill("123456");
  await page.locator("#login-submit-btn").click();
  await expect(page).toHaveURL(/\/$/);
});

test("API failures are visible instead of being replaced with fabricated records", async ({ page }) => {
  await mockAuthenticatedSession(page);
  await page.route("**/api/v1/org/organizations", (route) => json(route, 503, { error: "backend unavailable" }));

  await page.goto("/organization");

  await expect(page.getByRole("alert").filter({ hasText: "backend unavailable" })).toBeVisible();
});

test("MFA settings expose enrollment secret and one-time recovery codes", async ({ page }) => {
  await mockAuthenticatedSession(page);
  await page.route("**/api/v1/auth/mfa/status", (route) => json(route, 200, { enabled: false }));
  await page.route("**/api/v1/auth/mfa/enroll", (route) =>
    json(route, 200, {
      secret: "JBSWY3DPEHPK3PXP",
      otpauth_uri: "otpauth://totp/ERP:admin?secret=JBSWY3DPEHPK3PXP",
      recovery_codes: ["AAAA-BBBB", "CCCC-DDDD"],
    }),
  );
  await page.route("**/api/v1/auth/me/permissions", (route) => json(route, 200, { data: [] }));

  await page.goto("/settings");
  await page.getByRole("button", { name: /بدء الإعداد|start setup/i }).click();

  await expect(page.getByTestId("mfa-secret")).toHaveText("JBSWY3DPEHPK3PXP");
  await expect(page.getByTestId("recovery-codes")).toContainText("AAAA-BBBB");
});

test("a module write is sent to the API with the selected tenant and refreshed", async ({ page }) => {
  await mockAuthenticatedSession(page);
  const created = { id: "40000000-0000-0000-0000-000000000001", code: "E2E-CAT", name_en: "E2E Category" };
  let receivedBody: Record<string, unknown> | null = null;

  await page.route("**/api/v1/products/**", async (route) => {
    const request = route.request();
    if (request.method() === "POST" && new URL(request.url()).pathname.endsWith("/categories")) {
      receivedBody = request.postDataJSON() as Record<string, unknown>;
      await json(route, 201, { data: created });
      return;
    }
    await json(route, 200, { data: receivedBody ? [created] : [] });
  });

  await page.goto("/products");
  await page.getByRole("button", { name: /الفئات|categories/i }).click();
  await page.getByRole("button", { name: /عملية جديدة|new operation/i }).click();
  await page.locator("#operation-body").fill(JSON.stringify({ code: "E2E-CAT", name_en: "E2E Category", name_ar: "فئة اختبار" }));
  await page.locator("#operation-submit").click();

  await expect(page.getByTestId("operation-success")).toBeVisible();
  await expect(page.getByText("E2E-CAT", { exact: true })).toBeVisible();
  expect(receivedBody).toMatchObject({ legal_entity_id: entities[0].legal_entity_id, code: "E2E-CAT" });
});

test("dependent module filters never send an incomplete banking request", async ({ page }) => {
  await mockAuthenticatedSession(page);
  const bankAccounts = [
    { id: "50000000-0000-0000-0000-000000000001", bank_name_ar: "البنك الأول", bank_name_en: "First Bank", account_code: "B1" },
    { id: "50000000-0000-0000-0000-000000000002", bank_name_ar: "البنك الثاني", bank_name_en: "Second Bank", account_code: "B2" },
  ];
  const statementRequests: URL[] = [];

  await page.route("**/api/v1/banking/**", async (route) => {
    const url = new URL(route.request().url());
    if (url.pathname.endsWith("/bank-accounts")) {
      await json(route, 200, { data: bankAccounts });
      return;
    }
    if (url.pathname.endsWith("/statements")) {
      statementRequests.push(url);
      await json(route, 200, { data: [] });
      return;
    }
    await json(route, 200, { data: [] });
  });

  await page.goto("/banking");
  await page.getByRole("button", { name: /كشوف الحساب|statements/i }).click();

  const accountSelector = page.getByLabel(/الحساب البنكي|bank account/i);
  await expect(accountSelector).toHaveValue(bankAccounts[0].id);
  await expect.poll(() => statementRequests.length).toBeGreaterThan(0);
  expect(statementRequests.every((url) => url.searchParams.has("bank_account_id"))).toBe(true);
  expect(statementRequests.at(-1)?.searchParams.get("legal_entity_id")).toBe(entities[0].legal_entity_id);

  await accountSelector.selectOption(bankAccounts[1].id);
  await expect.poll(() => statementRequests.at(-1)?.searchParams.get("bank_account_id")).toBe(bankAccounts[1].id);
});

for (const locale of ["ar", "en"] as const) {
  for (const theme of ["light", "dark"] as const) {
    test(`customer form saves exact values in ${locale} ${theme}`, async ({ page }) => {
      await page.addInitScript(({ locale, theme }) => {
        localStorage.setItem("erp_locale", locale);
        localStorage.setItem("theme", theme);
      }, { locale, theme });
      await mockAuthenticatedSession(page);
      let received: Record<string, unknown> | null = null;
      await page.route("**/api/v1/receivables/**", async (route) => {
        if (route.request().method() === "POST") {
          received = route.request().postDataJSON();
          await json(route, 201, { id: "created-customer", ...received });
          return;
        }
        const customers = new URL(route.request().url()).pathname.endsWith("/customers");
        await json(route, 200, { data: customers && received ? [{ id: "created-customer", ...received }] : [] });
      });
      await page.goto("/customers");
      await expect(page.locator("html")).toHaveAttribute("dir", locale === "ar" ? "rtl" : "ltr");
      await expect(page.locator("html")).toHaveClass(new RegExp(theme));
      await page.getByRole("button", { name: /إضافة سجل|create record/i }).click();
      await page.locator("#create-name_ar").fill("عميل اختبار");
      await page.locator("#create-name_en").fill("Form Customer");
      await page.locator("#create-city").fill("Riyadh");
      await page.locator("#create-credit_limit").fill("999999999999999999.999999");
      await page.locator("#create-max_overdue_days").fill("30");
      await page.evaluate(() => window.scrollTo(0, 0));
      await page.screenshot({ path: `test-results/customer-form-${locale}-${theme}.png`, fullPage: true });
      await page.getByRole("button", { name: /^(حفظ|save)$/i }).click();
      await expect(page.getByRole("status")).toContainText(/تم حفظ|saved successfully/i);
      expect(received).toMatchObject({
        legal_entity_id: entities[0].legal_entity_id, name_en: "Form Customer", credit_limit: "999999999999999999.999999", max_overdue_days: 30,
      });
      expect(received).not.toHaveProperty("payment_terms_id");
      await expect(page.getByRole("cell", { name: "Form Customer", exact: true })).toBeVisible();
    });
  }
}

test("a structured form validates input and retains it after a denied save", async ({ page }) => {
  await mockAuthenticatedSession(page);
  let posts = 0;
  await page.route("**/api/v1/products/**", async (route) => {
    if (route.request().method() === "POST") {
      posts++;
      await json(route, 403, { error: { code: "FORBIDDEN", message: "Permission denied" } });
    } else await json(route, 200, { data: [] });
  });
  await page.goto("/products");
  await page.getByRole("button", { name: /الفئات|categories/i }).click();
  await page.getByRole("button", { name: /إضافة سجل|create record/i }).click();
  await page.getByRole("button", { name: /^(حفظ|save)$/i }).click();
  await expect(page.locator("#create-code")).toHaveAttribute("aria-invalid", "true");
  expect(posts).toBe(0);
  await page.locator("#create-code").fill("FORM-CATEGORY");
  await page.locator("#create-name_ar").fill("فئة");
  await page.locator("#create-name_en").fill("Category");
  await page.getByRole("button", { name: /^(حفظ|save)$/i }).click();
  await expect(page.getByRole("alert").filter({ hasText: "Permission denied" })).toBeVisible();
  await expect(page.locator("#create-code")).toHaveValue("FORM-CATEGORY");
  expect(posts).toBe(1);
});

test("tenant selection and system theme remain usable on mobile", async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 });
  await mockAuthenticatedSession(page);
  await page.route("**/api/v1/products/**", (route) => json(route, 200, { data: [] }));
  await page.goto("/products");
  const entity = page.getByLabel(/الكيان|legal entity/i);
  await expect(entity).toBeVisible();
  await entity.selectOption(entities[1].legal_entity_id);
  await expect(entity).toHaveValue(entities[1].legal_entity_id);
  await page.locator("#theme-toggle-btn select").selectOption("system");
  await expect(page.locator("#theme-toggle-btn select")).toHaveValue("system");
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
});

test("switching entity discards the open form and ignores a delayed previous response", async ({ page }) => {
  await mockAuthenticatedSession(page);
  let finishOldRequest: (() => void) | undefined;
  const oldRequestPending = new Promise<void>((resolve) => { finishOldRequest = resolve; });
  let started = false;
  await page.route("**/api/v1/products/items?**", async (route) => {
    const entity = new URL(route.request().url()).searchParams.get("legal_entity_id");
    if (entity === entities[0].legal_entity_id) {
      started = true;
      await oldRequestPending;
      await json(route, 200, { data: [{ id: "old", name_en: "OLD-ENTITY-SECRET" }] });
    } else await json(route, 200, { data: [{ id: "new", name_en: "NEW-ENTITY-RECORD" }] });
  });
  await page.route("**/api/v1/products/categories?**", (route) => json(route, 200, { data: [] }));
  await page.route("**/api/v1/products/uoms?**", (route) => json(route, 200, { data: [] }));
  await page.goto("/products");
  await expect.poll(() => started).toBe(true);
  await page.getByRole("button", { name: /إضافة سجل|create record/i }).click();
  await page.locator("#create-code").fill("UNSAVED-OLD-ENTITY");
  await page.getByLabel(/الكيان|legal entity/i).selectOption(entities[1].legal_entity_id);
  await expect(page.locator("#create-code")).toHaveCount(0);
  await expect(page.getByText("NEW-ENTITY-RECORD", { exact: true })).toBeVisible();
  finishOldRequest?.();
  await expect(page.getByText("OLD-ENTITY-SECRET", { exact: true })).toHaveCount(0);
});

test("user administration creates accounts, assigns scope, and uses versioned commands", async ({ page }) => {
  await mockAuthenticatedSession(page);
  await page.addInitScript(() => localStorage.setItem("erp_locale", "en"));
  const created = { id: "60000000-0000-4000-8000-000000000001", username: "new.user", display_name: "New Employee", email: "employee@example.sa", status: "ACTIVE", version: 1 };
  const requests: { path: string; body: Record<string, unknown> }[] = [];
  let exists = false;
  await page.route("**/api/v1/auth/users**", async (route) => {
    const path = new URL(route.request().url()).pathname;
    if (route.request().method() === "GET") { await json(route, 200, { data: exists ? [created] : [] }); return; }
    const body = route.request().postDataJSON() as Record<string, unknown>;
    requests.push({ path, body });
    if (path.endsWith("/status")) { created.status = String(body.status); created.version++; }
    else if (path.endsWith("/password")) created.version++;
    else exists = true;
    await json(route, path.endsWith("/users") ? 201 : 200, created);
  });
  const role = { id: "70000000-0000-4000-8000-000000000001", code: "ACCOUNTANT", name_en: "Accountant", name_ar: "محاسب" };
  await page.route("**/api/v1/authz/roles", (route) => json(route, 200, { data: [role] }));
  await page.route("**/api/v1/authz/assignments?user_id=*", (route) => json(route, 200, { data: [] }));
  await page.route("**/api/v1/authz/assignments", async (route) => {
    requests.push({ path: new URL(route.request().url()).pathname, body: route.request().postDataJSON() });
    await json(route, 201, { data: {} });
  });
  await page.goto("/authorization");
  await page.getByRole("button", { name: "Create user", exact: true }).click();
  await page.locator("#user-username").fill(created.username);
  await page.locator("#user-display_name").fill(created.display_name);
  await page.locator("#user-email").fill(created.email);
  await page.locator("#user-password").fill("New-test-password!");
  await page.getByRole("button", { name: "Save", exact: true }).click();
  await expect(page.getByRole("status")).toContainText("User created");
  expect(requests[0].body).not.toHaveProperty("legal_entity_id");
  await expect(page.locator("#user-password")).toHaveCount(0);
  await page.getByRole("button", { name: "Manage: New Employee" }).click();
  await page.getByLabel("Role", { exact: true }).selectOption(role.id);
  await page.getByRole("button", { name: "Assign role", exact: true }).click();
  await expect.poll(() => requests.length).toBe(2);
  expect(requests[1].body).toEqual({ user_id: created.id, role_id: role.id, legal_entity_id: entities[0].legal_entity_id });
  await page.getByLabel("Status", { exact: true }).selectOption("SUSPENDED");
  await page.getByRole("button", { name: "Update status", exact: true }).click();
  await expect(page.getByRole("cell", { name: "Suspended", exact: true })).toBeVisible();
  expect(requests[2].body).toEqual({ status: "SUSPENDED", expected_version: 1 });
  await page.getByLabel("New password", { exact: true }).fill("Changed-test-password!");
  await page.getByRole("button", { name: "Reset password", exact: true }).click();
  await expect(page.getByLabel("New password", { exact: true })).toHaveValue("");
  expect(requests[3].body).toEqual({ password: "Changed-test-password!", expected_version: 2 });
});

test("user management permission errors remain visible", async ({ page }) => {
  await mockAuthenticatedSession(page);
  await page.route("**/api/v1/auth/users**", (route) => json(route, 403, { error: { code: "FORBIDDEN", message: "Global user administration permission is required" } }));
  await page.route("**/api/v1/authz/roles", (route) => json(route, 200, { data: [] }));
  await page.goto("/authorization");
  await expect(page.getByRole("alert").filter({ hasText: "Global user administration permission is required" })).toBeVisible();
});

test("document library uploads files, versions and downloads with the selected entity", async ({ page }) => {
  await mockAuthenticatedSession(page);
  await page.addInitScript(() => localStorage.setItem("erp_locale", "en"));
  let uploaded = false;
  let version = 1;
  const writes: { contentType: string; body: string; entity: string | null }[] = [];
  await page.route("**/api/v1/documents/**", async (route) => {
    const request = route.request();
    const url = new URL(request.url());
    if (request.method() === "POST") {
      writes.push({ contentType: request.headers()["content-type"], body: request.postData() ?? "", entity: url.searchParams.get("legal_entity_id") });
      if (uploaded) version++;
      uploaded = true;
      await json(route, 201, { id: "document-1" });
    } else if (url.pathname.endsWith("/download")) {
      await route.fulfill({ status: 200, contentType: "text/plain", body: "Document content", headers: { "content-disposition": 'attachment; filename="document.txt"' } });
    } else await json(route, 200, { data: uploaded ? [{ id: "document-1", title: "Test Document", category: "GENERAL", current_version: { version_number: version, original_filename: "document.txt" } }] : [], total: uploaded ? 1 : 0 });
  });
  await page.goto("/documents");
  await page.getByRole("button", { name: "Upload document", exact: true }).click();
  await page.getByLabel("File", { exact: true }).setInputFiles({ name: "document.txt", mimeType: "text/plain", buffer: Buffer.from("Document content") });
  await page.getByLabel("Title", { exact: true }).fill("Test Document");
  await page.getByRole("button", { name: "Upload", exact: true }).click();
  await expect(page.getByRole("cell", { name: "Test Document", exact: true })).toBeVisible();
  expect(writes[0].contentType).toContain("multipart/form-data; boundary=");
  expect(writes[0].body).toContain("Document content");
  expect(writes[0].entity).toBe(entities[0].legal_entity_id);
  const download = page.waitForEvent("download");
  await page.getByRole("button", { name: "Download", exact: true }).click();
  expect((await download).suggestedFilename()).toBe("document.txt");
  await page.getByRole("button", { name: "Upload new version", exact: true }).click();
  await page.getByLabel("File", { exact: true }).setInputFiles({ name: "document.txt", mimeType: "text/plain", buffer: Buffer.from("Updated content") });
  await page.getByRole("button", { name: "Upload", exact: true }).click();
  await expect(page.getByRole("cell", { name: "2", exact: true })).toBeVisible();
  expect(writes[1].body).toContain("Updated content");
});
