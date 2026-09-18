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

async function json(route: Route, status: number, body: unknown) {
  await route.fulfill({ status, contentType: "application/json", body: JSON.stringify(body) });
}

async function mockAuthenticatedSession(page: Page) {
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
    await json(route, 200, { user });
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
    await json(route, 200, { user });
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
