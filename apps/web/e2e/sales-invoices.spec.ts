import { test, expect, type Page, type Route } from "@playwright/test";

const entity = "30000000-0000-4000-8000-000000000001";
const secondEntity = "30000000-0000-4000-8000-000000000002";
const customer = { id: "40000000-0000-4000-8000-000000000001", name_en: "Customer One", name_ar: "العميل الأول", tax_id: "310123456700003", currency_code: "SAR", status: "ACTIVE", address: { city: "Riyadh", country_code: "SA" } };
const product = { id: "50000000-0000-4000-8000-000000000001", code: "SERVICE", name_en: "Consulting", name_ar: "استشارات", base_uom_id: "60000000-0000-4000-8000-000000000001", list_price: "123456789012345.123456", is_sales_enabled: true, is_active: true, default_tax_code_id: "70000000-0000-4000-8000-000000000001" };
const uom = { id: product.base_uom_id, code: "HOUR", name_en: "Hour", name_ar: "ساعة", is_active: true };
const taxCode = { id: product.default_tax_code_id, code: "OUTPUT", name_en: "Output tax", name_ar: "ضريبة المخرجات", tax_type: "OUTPUT", is_active: true };
const invoice = { ID: "80000000-0000-4000-8000-000000000001", LegalEntityID: entity, InvoiceNumber: "INV-TEST-001", CustomerNameEn: customer.name_en, CustomerNameAr: customer.name_ar, IssueDate: "2026-10-07T00:00:00Z", DueDate: "2026-10-07T00:00:00Z", Status: "DRAFT", CurrencyCode: "SAR", ExchangeRate: "1", Subtotal: "246913578024690.246912", DiscountAmount: "0", TaxableAmount: "246913578024690.246912", TaxAmount: "0", TotalAmount: "246913578024690.246912", Lines: [{ Description: "Consulting", Quantity: "2", UOM: "HOUR", UnitPrice: product.list_price, TaxCode: "OUTPUT", TaxRatePercent: "0", TaxAmount: "0", NetAmount: "246913578024690.246912" }] };

async function respond(route: Route, body: unknown, status = 200) {
  await route.fulfill({ status, contentType: "application/json", body: JSON.stringify(body) });
}

async function setup(page: Page) {
  await page.context().addCookies([{ name: "erp_session", value: "invoice-browser-test", domain: "127.0.0.1", path: "/", httpOnly: true, sameSite: "Lax" }]);
  await page.route("**/api/v1/auth/me", (route) => respond(route, { user: { id: "user-1", username: "admin", display_name: "Administrator" } }));
  await page.route("**/api/v1/authz/assignments/me", (route) => respond(route, { data: [
    { id: "a1", role_code: "ADMIN", legal_entity_id: entity, legal_entity_name_en: "Company One", legal_entity_name_ar: "الشركة الأولى" },
    { id: "a2", role_code: "ADMIN", legal_entity_id: secondEntity, legal_entity_name_en: "Company Two", legal_entity_name_ar: "الشركة الثانية" },
  ] }));
  await page.route("**/api/v1/receivables/customers?**", (route) => respond(route, { data: [customer] }));
  await page.route("**/api/v1/products/items?**", (route) => respond(route, { data: [product] }));
  await page.route("**/api/v1/products/uoms?**", (route) => respond(route, { data: [uom] }));
  await page.route("**/api/v1/tax/codes?**", (route) => respond(route, { data: [taxCode] }));
  await page.route("**/api/v1/sales/orders?**", (route) => respond(route, []));
}

for (const locale of ["en", "ar"] as const) {
  for (const theme of ["light", "dark"] as const) {
    test(`sales invoice draft review and posting in ${locale} ${theme}`, async ({ page }) => {
      await page.addInitScript(({ locale, theme }) => { localStorage.setItem("erp_locale", locale); localStorage.setItem("theme", theme); }, { locale, theme });
      await setup(page);
      let received: Record<string, unknown> | null = null;
      let posted = false;
      await page.route("**/api/v1/sales/invoices**", async (route) => {
        const request = route.request();
        const url = new URL(request.url());
        expect(url.searchParams.get("legal_entity_id")).toBe(entity);
        if (request.method() === "POST" && url.pathname.endsWith("/post")) {
          expect(request.headers()["idempotency-key"]).toBe(`sales-invoice-post-${invoice.ID}`);
          posted = true;
          await respond(route, { ...invoice, Status: "POSTED", JournalEntryID: "90000000-0000-4000-8000-000000000001" });
        } else if (request.method() === "POST") {
          received = request.postDataJSON();
          await respond(route, invoice, 201);
        } else await respond(route, received ? [{ ...invoice, Status: posted ? "POSTED" : "DRAFT" }] : []);
      });
      await page.goto("/sales");
      await expect(page.locator("html")).toHaveAttribute("dir", locale === "ar" ? "rtl" : "ltr");
      await expect(page.locator("html")).toHaveClass(new RegExp(theme));
      await page.getByRole("button", { name: /^(New invoice|فاتورة جديدة)$/ }).click();
      const form = page.getByRole("form", { name: /^(New invoice|فاتورة جديدة)$/ });
      await form.getByRole("combobox", { name: /^(Customer|العميل)$/ }).selectOption(customer.id);
      await form.getByRole("combobox", { name: /Product or service|المنتج أو الخدمة/ }).selectOption(product.id);
      await form.getByRole("textbox", { name: /^(Quantity|الكمية)$/ }).fill("2");
      await expect(form.getByRole("textbox", { name: /^(Unit price|سعر الوحدة)$/ })).toHaveValue(product.list_price);
      await form.getByRole("button", { name: /Save draft and review|حفظ المسودة ومراجعتها/ }).click();
      const review = page.getByRole("region", { name: /^(Invoice review|مراجعة الفاتورة)$/ });
      await expect(review).toBeVisible();
      await expect(review).toContainText("246913578024690.246912");
      expect(received).toMatchObject({ legal_entity_id: entity, customer_id: customer.id, lines: [{ unit_price: product.list_price, quantity: "2", tax_code: "OUTPUT", uom: "HOUR" }] });
      const sentLine = (received as unknown as { lines: Record<string, unknown>[] }).lines[0];
      expect(sentLine).not.toHaveProperty("tax_rate_percent");
      expect(sentLine).not.toHaveProperty("total_amount");
      const postButton = review.getByRole("button", { name: /^(Post invoice|ترحيل الفاتورة)$/ });
      await expect(postButton).toBeDisabled();
      await review.getByRole("checkbox").check();
      await postButton.click();
      await expect(page.getByRole("status").filter({ hasText: /Invoice posted successfully|تم ترحيل الفاتورة بنجاح/ })).toBeVisible();
      await expect(review.getByRole("button", { name: /^(Post invoice|ترحيل الفاتورة)$/ })).toHaveCount(0);
      await page.screenshot({ path: `test-results/sales-invoice-${locale}-${theme}.png`, fullPage: true });
    });
  }
}

test("invoice form discards entity-specific choices when entity changes", async ({ page }) => {
  await page.addInitScript(() => localStorage.setItem("erp_locale", "en"));
  await setup(page);
  await page.route("**/api/v1/sales/invoices**", (route) => respond(route, []));
  await page.goto("/sales");
  await page.getByRole("button", { name: "New invoice", exact: true }).click();
  const form = page.getByRole("form", { name: "New invoice" });
  await form.getByRole("combobox", { name: "Customer", exact: true }).selectOption(customer.id);
  await page.getByLabel("Legal entity", { exact: true }).selectOption(secondEntity);
  await expect(form).toHaveCount(0);
  await page.getByRole("button", { name: "New invoice", exact: true }).click();
  await expect(page.getByRole("form", { name: "New invoice" }).getByRole("combobox", { name: "Customer", exact: true })).toHaveValue("");
});

test("invoice lookup permission failure prevents draft submission", async ({ page }) => {
  await page.addInitScript(() => localStorage.setItem("erp_locale", "en"));
  await setup(page);
  await page.route("**/api/v1/sales/invoices**", (route) => respond(route, []));
  await page.route("**/api/v1/tax/codes?**", (route) => respond(route, { error: { message: "Forbidden" } }, 403));
  await page.goto("/sales");
  await page.getByRole("button", { name: "New invoice", exact: true }).click();
  await expect(page.getByRole("alert").filter({ hasText: "Could not load" })).toBeVisible();
  await expect(page.getByRole("button", { name: "Save draft and review" })).toBeDisabled();
});
