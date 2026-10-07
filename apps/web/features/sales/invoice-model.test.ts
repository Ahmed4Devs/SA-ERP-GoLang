import { describe, expect, it } from "vitest";
import { invoiceInputsRounded, invoiceLinePayload, validLineAmounts, type InvoiceLineInput } from "./invoice-model";

const line: InvoiceLineInput = { key: "line-1", item_id: "item-1", description: "Consulting", uom: "HOUR", quantity: "2.000001", unit_price: "999999999999999999.123456", discount_percent: "0.00000001", tax_code: "OUTPUT", exemption_reason: "" };

describe("invoice financial inputs", () => {
  it("flags changed saved amounts while ignoring insignificant trailing zeros", () => {
    const input = { ...line, quantity: "2.0000", unit_price: "123.456789" };
    expect(invoiceInputsRounded([input], { Lines: [{ Quantity: "2", UnitPrice: "123.4568" }] })).toBe(true);
    expect(invoiceInputsRounded([input], { Lines: [{ Quantity: "2", UnitPrice: "123.4567890" }] })).toBe(false);
  });
  it("keeps decimal values exact and delegates totals and rates to the server", () => {
    expect(validLineAmounts(line)).toBe(true);
    const payload = invoiceLinePayload(line);
    expect(payload.quantity).toBe("2.000001");
    expect(payload.unit_price).toBe("999999999999999999.123456");
    expect(payload.discount_percent).toBe("0.00000001");
    expect(payload).not.toHaveProperty("tax_rate_percent");
    expect(payload).not.toHaveProperty("total_amount");
  });

  it.each([
    { quantity: "0" }, { quantity: "-1" }, { quantity: "1e3" }, { quantity: "0.0000001" },
    { unit_price: "NaN" }, { unit_price: "-0.01" }, { discount_percent: "100.00000001" }, { discount_percent: "-1" },
  ])("rejects invalid financial values %o", (invalid) => {
    expect(validLineAmounts({ ...line, ...invalid })).toBe(false);
  });
});
