import { expect, test, type Route } from "@playwright/test";

const entity = "30000000-0000-4000-8000-000000000001";
const actor = "10000000-0000-4000-8000-000000000001";
const maker = "10000000-0000-4000-8000-000000000002";
const account = "50000000-0000-4000-8000-000000000001";
const supplier = "60000000-0000-4000-8000-000000000001";
const tax = "70000000-0000-4000-8000-000000000001";
const branch = "80000000-0000-4000-8000-000000000001";
const json = (route: Route, body: unknown, status = 200) => route.fulfill({ status, contentType: "application/json", body: JSON.stringify(body) });

test.beforeEach(async ({ page }) => {
  await page.addInitScript(() => localStorage.setItem("erp_locale", "en"));
  await page.context().addCookies([{ name: "erp_session", value: "business-test", domain: "127.0.0.1", path: "/", httpOnly: true, sameSite: "Lax" }]);
  await page.route("**/api/v1/auth/me", route => json(route, { user: { id: actor, display_name: "Reviewer" } }));
  await page.route("**/api/v1/authz/assignments/me", route => json(route, { data: [{ id: "assignment", role_code: "ADMIN", legal_entity_id: entity, legal_entity_name_en: "Test Company" }] }));
  await page.route("**/api/v1/org/branches?**", route => json(route, { data: [{ id: branch, name_en: "Riyadh", name_ar: "الرياض" }] }));
});

test("paged resources advance offsets while unpaged resources advance locally", async ({ page }) => {
  const requests: URL[] = [];
  await page.route("**/api/v1/products/items?**", route => {
    const url = new URL(route.request().url()); requests.push(url);
    const offset = Number(url.searchParams.get("offset")); const limit = Number(url.searchParams.get("limit"));
    return json(route, { data: Array.from({ length: Math.min(limit, 60 - offset) }, (_, index) => ({ id: `id-${index + offset}`, code: `ITEM-${String(index + offset).padStart(3, "0")}` })), total: 60 });
  });
  await page.route("**/api/v1/products/categories?**", route => json(route, { data: Array.from({ length: 30 }, (_, index) => ({ id: `category-${index}`, code: `CAT-${String(index).padStart(3, "0")}` })) }));
  await page.goto("/products");
  await expect(page.getByRole("cell", { name: "ITEM-000", exact: true })).toBeVisible();
  await page.getByRole("button", { name: "Next", exact: true }).click();
  await expect(page.getByRole("cell", { name: "ITEM-025", exact: true })).toBeVisible();
  expect(requests.at(-1)?.searchParams.get("offset")).toBe("25");
  await page.getByRole("button", { name: "Categories", exact: true }).click();
  await expect(page.getByRole("cell", { name: "CAT-000", exact: true })).toBeVisible();
  await page.getByRole("button", { name: "Next", exact: true }).click();
  await expect(page.getByRole("cell", { name: "CAT-025", exact: true })).toBeVisible();
  await expect(page.getByRole("button", { name: "Next", exact: true })).toBeDisabled();
});

test("treasury approval uses authenticated actor and posting failures preserve the document", async ({ page }) => {
  const payment = { id: "90000000-0000-4000-8000-000000000001", payment_number: "PAY-1", status: "DRAFT", amount: "10.123456", created_by_user_id: maker };
  const writes: { path: string; body: unknown; entity: string | null }[] = [];
  await page.route("**/api/v1/treasury/**", route => {
    const url = new URL(route.request().url());
    if (route.request().method() === "POST") {
      writes.push({ path: url.pathname, body: route.request().postDataJSON(), entity: url.searchParams.get("legal_entity_id") });
      if (url.pathname.endsWith("/approve")) { payment.status = "APPROVED"; return json(route, payment); }
      return json(route, { error: "Accounting period is closed" }, 400);
    }
    return json(route, url.pathname.endsWith("/payments") ? [payment] : []);
  });
  await page.goto("/treasury");
  await page.getByRole("button", { name: "Payments", exact: true }).click();
  await page.getByRole("cell", { name: "PAY-1", exact: true }).click();
  await page.getByRole("button", { name: "Approve", exact: true }).click();
  await page.getByRole("button", { name: "Confirm action", exact: true }).click();
  await expect(page.getByRole("cell", { name: "Approved", exact: true })).toBeVisible();
  expect(writes[0]).toEqual({ path: `/api/v1/treasury/payments/${payment.id}/approve`, body: {}, entity });
  await page.getByRole("cell", { name: "PAY-1", exact: true }).click();
  await page.getByRole("button", { name: "Pay and post", exact: true }).click();
  await page.getByRole("button", { name: "Confirm action", exact: true }).click();
  await expect(page.getByRole("alert").filter({ hasText: "Accounting period is closed" })).toBeVisible();
  await expect(page.getByRole("cell", { name: "Approved", exact: true })).toBeVisible();
});

test("creators cannot approve their own treasury documents", async ({ page }) => {
  await page.route("**/api/v1/treasury/**", route => json(route, new URL(route.request().url()).pathname.endsWith("/payments") ? [{ id: "own", payment_number: "OWN-1", status: "DRAFT", created_by_user_id: actor }] : []));
  await page.goto("/treasury");
  await page.getByRole("button", { name: "Payments", exact: true }).click();
  await page.getByRole("cell", { name: "OWN-1", exact: true }).click();
  await expect(page.getByRole("button", { name: "Approve", exact: true })).toBeDisabled();
  await expect(page.getByText("A different user must approve or post this document.")).toBeVisible();
});

test("supplier invoice form preserves decimal lines and tax code without raw payload input", async ({ page }) => {
  let saved: Record<string, unknown> | null = null;
  await page.route("**/api/v1/accounting/accounts?**", route => json(route, { data: [{ id: account, code: "523000", name_en: "Expense", allow_posting: true }] }));
  await page.route("**/api/v1/tax/codes?**", route => json(route, { data: [{ id: tax, code: "VAT15", name_en: "VAT 15%" }] }));
  await page.route("**/api/v1/products/items?**", route => json(route, { data: [] }));
  await page.route("**/api/v1/payables/**", route => {
    if (route.request().method() === "POST") { saved = route.request().postDataJSON(); return json(route, { id: "invoice", ...saved }, 201); }
    return json(route, { data: new URL(route.request().url()).pathname.endsWith("/suppliers") ? [{ id: supplier, name_en: "Supplier" }] : [] });
  });
  await page.goto("/payables");
  await page.getByRole("button", { name: "Create record", exact: true }).click();
  await page.locator("#create-supplier_id").selectOption(supplier);
  await page.locator("#create-supplier_invoice_number").fill("SUP-1");
  await page.locator("#create-invoice_date").fill("2026-10-07");
  await page.locator("#create-due_date").fill("2026-11-07");
  await page.locator("#create-currency_code").fill("USD");
  await page.locator("#create-exchange_rate").fill("3.750000000001");
  await page.locator("#line-0-expense_account_id").selectOption(account);
  await page.locator("#line-0-description").fill("Exact precision");
  await page.locator("#line-0-quantity").fill("2.123456");
  await page.locator("#line-0-unit_price").fill("999999999999.123456");
  await page.locator("#line-0-tax_code_id").selectOption(tax);
  await page.evaluate(() => window.scrollTo(0, 0));
  await page.screenshot({ path: "test-results/supplier-invoice-form-en.png", fullPage: true });
  await page.getByRole("button", { name: "Save", exact: true }).click();
  await expect(page.getByRole("status")).toBeVisible();
  expect(saved).toMatchObject({ legal_entity_id: entity, exchange_rate: "3.750000000001", lines: [{ expense_account_id: account, quantity: "2.123456", unit_price: "999999999999.123456", tax_code_id: tax, discount_percent: "0" }] });
  expect(saved).not.toHaveProperty("created_by");
  await expect(page.locator("#operation-body")).toHaveCount(0);
});

test("payroll references normalize Go fields and navigation preserves the selected branch", async ({ page }) => {
  let saved: Record<string, unknown> | null = null;
  const period = "90000000-0000-4000-8000-000000000001";
  await page.route("**/api/v1/payroll/**", route => {
    if (route.request().method() === "POST") { saved = route.request().postDataJSON(); return json(route, saved, 201); }
    return json(route, new URL(route.request().url()).pathname.endsWith("/periods") ? [{ ID: period, PeriodNameEn: "October 2026", PeriodYear: 2026, PeriodMonth: 10 }] : []);
  });
  await page.goto("/payroll");
  await page.getByRole("button", { name: "Create record", exact: true }).click();
  await page.locator("#create-period_id").selectOption(period);
  await page.locator("#create-run_number").fill("PAY-2026-10");
  await page.locator("#create-run_date").fill("2026-10-31");
  await page.locator("#create-cutoff_date").fill("2026-10-31");
  await page.getByRole("button", { name: "Save", exact: true }).click();
  await expect(page.getByRole("status")).toBeVisible();
  expect(saved).toMatchObject({ period_id: period, legal_entity_id: entity });
  await page.getByRole("button", { name: "Payslips", exact: true }).click();
  await expect(page.getByLabel("Payroll period", { exact: true })).toHaveValue(period);
  await page.getByLabel("Branch", { exact: true }).selectOption(branch);
  await page.route("**/api/v1/treasury/**", route => json(route, []));
  await page.getByLabel("Find a page", { exact: true }).fill("Treasury");
  await page.getByRole("region", { name: "Find a page" }).getByRole("link", { name: "Treasury" }).click();
  await expect(page).toHaveURL(/\/treasury$/);
  await expect(page.getByLabel("Branch", { exact: true })).toHaveValue(branch);
});

test("approval limit updates carry the saved version and retain values on a conflict", async ({ page }) => {
  const writes: Record<string, unknown>[] = [];
  await page.route("**/api/v1/auth/users?**", route => json(route, { data: [{ id: actor, display_name: "Reviewer", status: "ACTIVE" }] }));
  await page.route("**/api/v1/workflow/**", route => {
    if (route.request().method() === "PUT") {
      writes.push(route.request().postDataJSON());
      return json(route, { error: { code: "CONFLICT", message: "The approval limit changed; reload and retry" } }, 409);
    }
    return json(route, { data: new URL(route.request().url()).pathname.endsWith("/approval-limits") ? [{ legal_entity_id: entity, user_id: actor, resource_type: "SALES_INVOICE", maximum_base_amount: "1000.123456", version: 2 }] : [] });
  });
  await page.goto("/workflow");
  await page.getByRole("button", { name: "Manage limits", exact: true }).click();
  await page.getByLabel("Approver", { exact: true }).selectOption(actor);
  await expect(page.getByLabel("Maximum base amount", { exact: true })).toHaveValue("1000.123456");
  await page.getByLabel("Maximum base amount", { exact: true }).fill("2000.654321");
  await page.getByRole("button", { name: "Save limit", exact: true }).click();
  await expect(page.getByRole("alert").filter({ hasText: "The approval limit changed" })).toBeVisible();
  expect(writes[0]).toEqual({ legal_entity_id: entity, user_id: actor, resource_type: "SALES_INVOICE", maximum_base_amount: "2000.654321", version: 2 });
  await expect(page.getByLabel("Maximum base amount", { exact: true })).toHaveValue("2000.654321");
});

test("cash account creation uses the selected branch and explicit ledger mapping", async ({ page }) => {
  let saved: Record<string, unknown> | null = null;
  await page.route("**/api/v1/accounting/accounts?**", route => json(route, { data: [{ id: account, name_en: "Cash ledger", allow_posting: true }] }));
  await page.route("**/api/v1/treasury/cash-accounts?**", route => {
    if (route.request().method() === "POST") { saved = route.request().postDataJSON(); return json(route, { id: "new-cash", ...saved }, 201); }
    return json(route, []);
  });
  await page.goto("/treasury");
  await page.getByLabel("Branch", { exact: true }).selectOption(branch);
  await page.getByRole("button", { name: "Create record", exact: true }).click();
  await expect(page.locator("#create-branch_id")).toHaveValue(branch);
  await page.locator("#create-account_code").fill("CASH-RUH");
  await page.locator("#create-name_ar").fill("صندوق الرياض");
  await page.locator("#create-name_en").fill("Riyadh Cash");
  await page.locator("#create-gl_account_id").selectOption(account);
  await page.getByRole("button", { name: "Save", exact: true }).click();
  await expect(page.getByRole("status")).toBeVisible();
  expect(saved).toMatchObject({ legal_entity_id: entity, branch_id: branch, gl_account_id: account });
});
