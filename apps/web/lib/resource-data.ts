import { apiFetch } from "@/lib/api-client";
import { fieldKey } from "@/lib/record-presentation";

export type RecordRow = Record<string, unknown>;
const pagedPaths = new Set([
  "/api/v1/products/items", "/api/v1/receivables/customers", "/api/v1/receivables/receipts",
  "/api/v1/payables/suppliers", "/api/v1/payables/invoices", "/api/v1/payables/payments",
  ...["quotations", "orders", "deliveries", "invoices", "credit-notes"].map(p => `/api/v1/sales/${p}`),
  ...["requests", "rfqs", "quotes", "orders", "receipts"].map(p => `/api/v1/purchasing/${p}`),
  ...["movements", "balances", "transfers", "adjustments", "counts"].map(p => `/api/v1/inventory/${p}`),
  "/api/v1/tax/transactions", "/api/v1/documents/", "/api/v1/notifications/",
]);
const searchablePaths = new Set([
  "/api/v1/receivables/customers", "/api/v1/receivables/receipts", "/api/v1/payables/suppliers",
  "/api/v1/payables/invoices", "/api/v1/payables/payments",
]);
export const hasServerPagination = (path: string) => pagedPaths.has(path.split("?")[0]);
export const hasServerSearch = (path: string) => searchablePaths.has(path.split("?")[0]);

export function toRecord(value: unknown): RecordRow {
  return value && typeof value === "object" && !Array.isArray(value)
    ? Object.fromEntries(Object.entries(value).map(([key, item]) => [fieldKey(key), item])) : { value };
}
export function normalizeRows(payload: unknown): RecordRow[] {
  if (Array.isArray(payload)) return payload.map(toRecord);
  if (payload == null) return [];
  if (typeof payload !== "object") return [{ value: payload }];
  const envelope = payload as RecordRow;
  for (const key of ["data", "items", "results", "records"]) {
    if (Array.isArray(envelope[key])) return (envelope[key] as unknown[]).map(toRecord);
    if (key in envelope && envelope[key] === null) return [];
  }
  if (envelope.data && typeof envelope.data === "object") return [toRecord(envelope.data)];
  return [toRecord(envelope)];
}
export function responseTotal(payload: unknown): number | null {
  if (!payload || typeof payload !== "object" || Array.isArray(payload)) return null;
  const data = payload as RecordRow;
  const total = data.total ?? (data.meta as RecordRow | undefined)?.total;
  return typeof total === "number" && Number.isSafeInteger(total) && total >= 0 ? total : null;
}
export function queryPath(path: string, params: Record<string, string>) {
  const [pathname, rawQuery = ""] = path.split("?", 2);
  const query = new URLSearchParams(rawQuery);
  Object.entries(params).forEach(([key, value]) => { if (value) query.set(key, value); });
  return query.size ? `${pathname}?${query}` : pathname;
}
// Reference selectors must include later pages as well as the first page.
export async function loadReferenceRows(path: string, entity: string | null, signal: AbortSignal, tenant = true) {
  const rows: RecordRow[] = [];
  const paginated = hasServerPagination(path);
  let offset = 0;
  while (true) {
    const payload = await apiFetch<unknown>(queryPath(path, {
      ...(tenant && entity ? { legal_entity_id: entity } : {}),
      ...(paginated ? { limit: "100", offset: String(offset) } : {}),
    }), { signal });
    const page = normalizeRows(payload);
    rows.push(...page);
    if (!paginated || page.length === 0 || (responseTotal(payload) !== null && rows.length >= responseTotal(payload)!)) break;
    // Some services enforce a lower maximum than requested; advance by the actual size.
    offset += page.length;
    if (rows.length > 10000) throw new Error("Reference list is too large; narrow the legal entity scope.");
    if (page.length < 100 && responseTotal(payload) === null) break;
  }
  return rows.filter(row => row.id && row.is_active !== false && row.status !== "INACTIVE");
}
