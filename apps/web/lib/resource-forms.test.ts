import { describe, expect, it } from "vitest";
import { createResourceSchema, resourceFormPayload, resourceForms } from "./resource-forms";

const messages = { required: "required", invalid: "invalid" };

describe("domain form contracts", () => {
  it("preserves exact monetary values and only serializes declared fields", () => {
    const definition = resourceForms["/api/v1/receivables/customers"];
    const values = {
      name_ar: "عميل", name_en: "Customer", customer_type: "B2B", city: "Riyadh",
      country_code: "SA", currency_code: "SAR", credit_limit: "999999999999999999.999999", max_overdue_days: "30",
      legal_entity_id: "untrusted-entity", status: "POSTED", payment_terms_id: "",
    };
    expect(createResourceSchema(definition, messages).safeParse(values).success).toBe(true);
    const payload = resourceFormPayload(definition, values);
    expect(payload.credit_limit).toBe("999999999999999999.999999");
    expect(payload.max_overdue_days).toBe(30);
    expect(payload).not.toHaveProperty("legal_entity_id");
    expect(payload).not.toHaveProperty("status");
    expect(payload).not.toHaveProperty("payment_terms_id");
  });

  it("rejects invalid types, monetary precision and missing required references", () => {
    const definition = resourceForms["/api/v1/products/items"];
    const valid = { code: "ITEM", name_ar: "صنف", name_en: "Item", item_type: "STOCK", tracking_type: "NONE", base_uom_id: "30000000-0000-4000-8000-000000000001", cost_price: "0", list_price: "9.999999" };
    const schema = createResourceSchema(definition, messages);
    expect(schema.safeParse(valid).success).toBe(true);
    for (const invalid of [
      { item_type: "POSTED" }, { list_price: "1e3" }, { cost_price: "-1" }, { list_price: "1.1234567" }, { base_uom_id: "" }, { base_uom_id: "not-a-uuid" },
    ]) expect(schema.safeParse({ ...valid, ...invalid }).success).toBe(false);
  });

  it("validates optional contact data without requiring it", () => {
    const definition = resourceForms["/api/v1/payables/suppliers"];
    const valid = { name_ar: "مورد", name_en: "Supplier", supplier_type: "LOCAL", city: "Riyadh", country_code: "SA", currency_code: "SAR" };
    const schema = createResourceSchema(definition, messages);
    expect(schema.safeParse(valid).success).toBe(true);
    expect(schema.safeParse({ ...valid, email: "invalid" }).success).toBe(false);
    expect(schema.safeParse({ ...valid, tax_id: "123" }).success).toBe(false);
    expect(schema.safeParse({ ...valid, tax_id: "300000000000003" }).success).toBe(true);
  });
});
