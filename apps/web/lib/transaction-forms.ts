import type { FormField, ResourceFormDefinition } from "@/lib/resource-forms";

const field = (key: string, options: Partial<FormField> = {}): FormField => ({ key, label: key, required: true, ...options });
const select = (key: string, source: string, required = true) => field(key, { type: "select", source, required });
const choices = (key: string, values: string[], initial = values[0]) => field(key, { type: "select", initial, options: values.map(value => ({ value, label: value.toLowerCase() })) });
const date = (key: string, dateTime = false) => field(key, { type: "date", dateTime });
const decimal = (key: string, initial?: string) => field(key, { type: "decimal", initial });
const account = (key: string, required = true) => select(key, "/api/v1/accounting/accounts", required);
const currency = field("currency_code", { initial: "SAR", pattern: "^[A-Z]{3}$" });
const notes = field("notes", { type: "textarea", required: false });
const reference = field("reference_number", { required: false });
const paymentMethod = choices("payment_method", ["CASH", "BANK_TRANSFER", "CHEQUE", "MADA", "SADAD"]);
const amount = field("amount", { type: "decimal", min: 0.000001 });
const year = field("year", { type: "integer", min: 1900, max: 9999 });
const month = field("month", { type: "integer", min: 1, max: 12 });
const stockAccounts = [account("inventory_account_id"), account("offset_account_id")];
const bin = select("bin_id", "/api/v1/inventory/warehouses/{warehouse_id}/bins", false);
const stockItem = select("item_id", "/api/v1/products/items");

export const transactionForms: Record<string, ResourceFormDefinition> = {
  "/api/v1/treasury/cash-accounts": { fields: [field("account_code"), field("name_ar"), field("name_en"), choices("account_type", ["PETTY_CASH", "CASH_REGISTER", "MAIN_VAULT"]), account("gl_account_id"), select("branch_id", "/api/v1/org/branches", false), currency, decimal("max_limit", "0") ] },
  "/api/v1/treasury/receipts": { fields: [field("receipt_number"), select("cash_account_id", "/api/v1/treasury/cash-accounts"), date("receipt_date", true), choices("receipt_type", ["MISCELLANEOUS", "CUSTOMER_PAYMENT", "CASH_SALE", "DEPOSIT", "INTER_BRANCH"]), field("received_from"), amount, currency, choices("payment_method", ["CASH", "CHEQUE"]), account("gl_offset_account_id"), field("cheque_number", { required: false }), reference, field("description", { type: "textarea", required: false })] },
  "/api/v1/treasury/payments": { fields: [field("payment_number"), select("cash_account_id", "/api/v1/treasury/cash-accounts"), date("payment_date", true), choices("payment_type", ["PETTY_CASH_EXPENSE", "SUPPLIER_PAYMENT", "CASH_ADVANCE", "MISCELLANEOUS"]), field("paid_to"), amount, currency, account("gl_offset_account_id"), reference, field("description", { type: "textarea", required: false })] },
  "/api/v1/treasury/transfers": { fields: [field("transfer_number"), date("transfer_date", true), choices("source_type", ["CASH", "BANK"]),
    { ...select("source_cash_account_id", "/api/v1/treasury/cash-accounts"), showWhen: { key: "source_type", value: "CASH" } },
    { ...select("source_bank_account_id", "/api/v1/banking/bank-accounts"), showWhen: { key: "source_type", value: "BANK" } },
    choices("destination_type", ["CASH", "BANK"]),
    { ...select("destination_cash_account_id", "/api/v1/treasury/cash-accounts"), showWhen: { key: "destination_type", value: "CASH" } },
    { ...select("destination_bank_account_id", "/api/v1/banking/bank-accounts"), showWhen: { key: "destination_type", value: "BANK" } }, amount, currency, decimal("fee_amount", "0"), account("fee_gl_account_id", false), reference, notes] },
  "/api/v1/payables/invoices": { fields: [select("supplier_id", "/api/v1/payables/suppliers"), field("supplier_invoice_number"), date("invoice_date"), date("due_date"), currency, field("exchange_rate", { type: "decimal", required: false, min: 0.000000000001, decimalPlaces: 12 }), notes], lines: { key: "lines", fields: [select("item_id", "/api/v1/products/items", false), account("expense_account_id"), field("description"), field("quantity", { type: "decimal", min: 0.000001, initial: "1" }), decimal("unit_price"), field("discount_percent", { type: "decimal", initial: "0", max: 100 }), select("tax_code_id", "/api/v1/tax/codes")] } },
  "/api/v1/payables/payments": { fields: [select("supplier_id", "/api/v1/payables/suppliers"), date("payment_date"), paymentMethod, account("disbursement_account_id"), amount, currency, reference, notes], lines: { key: "allocations", optional: true, fields: [select("open_item_id", "/api/v1/payables/suppliers/{supplier_id}/open-items"), amount] } },
  "/api/v1/receivables/receipts": { notice: "receiptPostingHint", submitLabel: "recordReceipt", fields: [select("customer_id", "/api/v1/receivables/customers"), date("receipt_date"), paymentMethod, account("deposit_account_id"), amount, currency, reference, notes], lines: { key: "allocations", optional: true, fields: [select("open_item_id", "/api/v1/receivables/customers/{customer_id}/open-items"), amount] } },
  "/api/v1/payroll/periods": { fields: [year, month, field("name_ar"), field("name_en"), date("start_date"), date("end_date"), date("payment_due_date")] },
  "/api/v1/payroll/runs": { fields: [select("period_id", "/api/v1/payroll/periods"), field("run_number"), date("run_date"), date("cutoff_date")] },
  "/api/v1/fixedassets/depreciation/runs": { createPath: "/api/v1/fixedassets/depreciation/runs/calculate", fields: [select("book_id", "/api/v1/fixedassets/books"), year, month] },
  "/api/v1/inventory/adjustments": { fields: [select("warehouse_id", "/api/v1/inventory/warehouses"), choices("adjustment_type", ["INCREASE", "DECREASE", "SCRAP", "FOUND", "DAMAGED", "CYCLE_COUNT"]), field("reason_code"), ...stockAccounts, notes], lines: { key: "lines", fields: [stockItem, bin, field("quantity", { type: "decimal", min: 0.000001 }), decimal("unit_cost", "0"), field("reason", { required: false }), field("lot_number", { required: false }), field("serial_number", { required: false })] } },
  "/api/v1/inventory/counts": { fields: [select("warehouse_id", "/api/v1/inventory/warehouses"), choices("count_type", ["FULL", "CYCLE", "SPOT"]), ...stockAccounts], lines: { key: "lines", fields: [stockItem, bin, decimal("counted_quantity"), decimal("unit_cost", "0"), notes] } },
};
