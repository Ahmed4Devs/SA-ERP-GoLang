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
