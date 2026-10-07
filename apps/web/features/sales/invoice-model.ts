export type Row = Record<string, unknown>;

export function rows(payload: unknown): Row[] {
  if (Array.isArray(payload)) return payload as Row[];
  if (payload && typeof payload === "object" && "data" in payload && Array.isArray(payload.data)) return payload.data as Row[];
  return [];
}

// Existing sales endpoints expose Go field names; master data uses JSON tags.
export function field(record: Row, key: string): string {
  const snakeKey = key.replace(/([A-Z]+)([A-Z][a-z])/g, "$1_$2").replace(/([a-z\d])([A-Z])/g, "$1_$2").toLowerCase();
  const value = record[key] ?? record[snakeKey];
  return value == null ? "" : String(value);
}

export function invoiceID(record: Row): string { return String(record.ID ?? record.id ?? ""); }

export interface InvoiceLineInput {
  key: string; item_id: string; description: string; uom: string;
  quantity: string; unit_price: string; discount_percent: string; tax_code: string; exemption_reason: string;
}

export function emptyLine(): InvoiceLineInput {
  return { key: crypto.randomUUID(), item_id: "", description: "", uom: "", quantity: "1", unit_price: "0", discount_percent: "0", tax_code: "", exemption_reason: "" };
}

export function validLineAmounts(line: InvoiceLineInput): boolean {
  const amount = /^(?:0|[1-9]\d{0,17})(?:\.\d{1,6})?$/;
  const percent = /^(?:(?:0|[1-9]\d?)(?:\.\d{1,8})?|100(?:\.0{1,8})?)$/;
  return amount.test(line.quantity) && /[1-9]/.test(line.quantity) && amount.test(line.unit_price) && percent.test(line.discount_percent);
}

// Preserve exact decimal text. Totals and tax rates deliberately are not sent.
export function invoiceLinePayload(line: InvoiceLineInput) {
  return { item_id: line.item_id, description: line.description.trim(), uom: line.uom,
    quantity: line.quantity, unit_price: line.unit_price, discount_percent: line.discount_percent,
    tax_code: line.tax_code, exemption_reason: line.exemption_reason.trim() || null };
}

export function invoiceInputsRounded(lines: InvoiceLineInput[], invoice: Row): boolean {
  const saved = rows(invoice.Lines ?? invoice.lines);
  const normalize = (value: string) => value.includes(".") ? value.replace(/0+$/, "").replace(/\.$/, "") : value;
  return lines.some((line, index) => saved[index] && (
    normalize(line.quantity) !== normalize(field(saved[index], "Quantity")) ||
    normalize(line.unit_price) !== normalize(field(saved[index], "UnitPrice"))
  ));
}
