import { describe, expect, it } from "vitest";
import { createResourceSchema, resourceFormPayload } from "./resource-forms";
import { transactionForms } from "./transaction-forms";
import { availableActions } from "@/components/modules/resource-actions";
import { normalizeRows, responseTotal, hasServerPagination } from "./resource-data";

const messages = { required: "required", invalid: "invalid" };
describe("business document contracts", () => {
  it("preserves supplier amounts and FX snapshots while validating actual dates", () => {
    const definition = transactionForms["/api/v1/payables/invoices"];
    const values = { supplier_id: "30000000-0000-4000-8000-000000000001", supplier_invoice_number: "INV-1", invoice_date: "2026-02-28", due_date: "2026-03-01", currency_code: "USD", exchange_rate: "3.750000000001" };
    const schema = createResourceSchema(definition, messages);
    expect(schema.safeParse(values).success).toBe(true);
    expect(resourceFormPayload(definition, values).exchange_rate).toBe("3.750000000001");
    expect(schema.safeParse({ ...values, invoice_date: "2026-02-30" }).success).toBe(false);
    expect(schema.safeParse({ ...values, exchange_rate: "3.7500000000001" }).success).toBe(false);
  });
  it("does not serialize bank IDs hidden by a cash transfer selection", () => {
    const definition = transactionForms["/api/v1/treasury/transfers"];
    const payload = resourceFormPayload(definition, { source_type: "CASH", source_cash_account_id: "cash-1", source_bank_account_id: "bank-1", transfer_date: "2026-10-07", created_by_user_id: "forged" });
    expect(payload.source_cash_account_id).toBe("cash-1");
    expect(payload.transfer_date).toBe("2026-10-07T00:00:00Z");
    expect(payload).not.toHaveProperty("source_bank_account_id");
    expect(payload).not.toHaveProperty("created_by_user_id");
  });
  it("only exposes actions for valid document stages", () => {
    expect(availableActions("/api/v1/treasury/payments", { id: "1", status: "DRAFT" }).map(rule => rule.action)).toEqual(["approve"]);
    expect(availableActions("/api/v1/treasury/payments", { id: "1", status: "APPROVED" }).map(rule => rule.action)).toEqual(["pay"]);
    expect(availableActions("/api/v1/treasury/payments", { id: "1", status: "PAID" })).toEqual([]);
    expect(availableActions("/api/v1/inventory/counts", { id: "1", status: "PLANNED" }).map(rule => rule.action)).toEqual(["reconcile"]);
    expect(availableActions("/api/v1/receivables/receipts", { id: "1", status: "DRAFT" })).toEqual([]);
  });
  it("normalizes Go field names without losing total metadata or inventing pagination", () => {
    const payload = { data: [{ ID: "1", PeriodNameAr: "فترة", GLAccountID: "account" }], total: 101 };
    expect(normalizeRows(payload)).toEqual([{ id: "1", period_name_ar: "فترة", gl_account_id: "account" }]);
    expect(responseTotal(payload)).toBe(101);
    expect(hasServerPagination("/api/v1/inventory/counts")).toBe(true);
    expect(hasServerPagination("/api/v1/payroll/periods")).toBe(false);
  });
});
