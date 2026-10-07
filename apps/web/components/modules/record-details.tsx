"use client";

import { useState } from "react";
import { useLocale } from "@/components/layout/shell";
import { apiFetch } from "@/lib/api-client";
import { queryPath, toRecord, type RecordRow } from "@/lib/resource-data";
import { fieldLabel, stateLabel } from "@/lib/record-presentation";

const detailPaths = new Set([
  "/api/v1/payables/invoices", "/api/v1/payables/payments", "/api/v1/payables/credit-notes",
  "/api/v1/payroll/runs", "/api/v1/payroll/payslips", "/api/v1/fixedassets/depreciation/runs",
  "/api/v1/fixedassets/assets", "/api/v1/inventory/transfers", "/api/v1/accounting/journal-entries",
  "/api/v1/inventory/adjustments", "/api/v1/inventory/counts",
  ...["orders", "quotations", "deliveries", "invoices", "credit-notes"].map(key => `/api/v1/sales/${key}`),
  ...["orders", "requests", "rfqs", "quotes", "receipts"].map(key => `/api/v1/purchasing/${key}`),
]);
function Value({ value, name, depth = 0 }: { value: unknown; name?: string; depth?: number }) {
  const { locale, t } = useLocale();
  if (value == null || value === "") return <>—</>;
  if (typeof value === "boolean") return <>{value ? "✓" : "✕"}</>;
  if (Array.isArray(value)) return value.length ? <ol className="space-y-3">{value.map((entry, index) => <li key={index} className="rounded-lg border p-3"><Value value={entry} depth={depth + 1} /></li>)}</ol> : <>{t.business.noValues}</>;
  if (typeof value === "object" && depth < 5) return <dl className={`grid gap-3 ${depth === 0 ? "sm:grid-cols-2" : ""}`}>{Object.entries(toRecord(value)).map(([key, item]) => <div key={key} className={item && typeof item === "object" ? "col-span-full" : "min-w-0"}><dt className="text-xs text-muted-foreground">{fieldLabel(key, locale)}</dt><dd className="mt-1 break-words text-sm"><Value value={item} name={key} depth={depth + 1} /></dd></div>)}</dl>;
  return <>{name === "status" || name?.endsWith("_status") ? stateLabel(value, locale) : String(value)}</>;
}
export function RecordDetails({ row, path, legalEntityId }: { row: RecordRow; path: string; legalEntityId: string | null }) {
  const { t } = useLocale();
  const [detail, setDetail] = useState<unknown>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  async function load() {
    setBusy(true); setError(null);
    try {
      const payload = await apiFetch<RecordRow>(queryPath(`${path}/${encodeURIComponent(String(row.id))}`, legalEntityId ? { legal_entity_id: legalEntityId } : {}));
      setDetail(payload.data ?? payload);
    } catch (failure) { setError(failure instanceof Error ? failure.message : t.business.detailsError); }
    finally { setBusy(false); }
  }
  return <div className="space-y-4">
    {detailPaths.has(path) && row.id ? <button type="button" disabled={busy} onClick={() => void load()} className="rounded-lg border px-3 py-2 text-sm disabled:opacity-50">{busy ? t.business.working : t.business.loadDetails}</button> : null}
    {error && <p role="alert" className="text-sm text-destructive">{error}</p>}
    <Value value={detail ?? row} />
  </div>;
}
