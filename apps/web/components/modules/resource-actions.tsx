"use client";

import { useState } from "react";
import { useLocale } from "@/components/layout/shell";
import { useAuth } from "@/lib/auth-context";
import { apiFetch } from "@/lib/api-client";
import { queryPath, type RecordRow } from "@/lib/resource-data";

type Action = "post" | "approve" | "pay" | "complete" | "calculate" | "reconcile" | "dispatch" | "receive" | "confirm" | "submit";
type Rule = { action: Action; states: string[]; independent?: boolean };
const post: Rule = { action: "post", states: ["DRAFT", "APPROVED"] };
const rules: Record<string, Rule[]> = {
  "/api/v1/treasury/receipts": [{ ...post, independent: true }],
  "/api/v1/treasury/payments": [{ action: "approve", states: ["DRAFT"], independent: true }, { action: "pay", states: ["APPROVED"], independent: true }],
  "/api/v1/treasury/transfers": [{ action: "complete", states: ["DRAFT", "IN_TRANSIT"], independent: true }],
  "/api/v1/payroll/runs": [{ action: "calculate", states: ["DRAFT"] }, { action: "approve", states: ["CALCULATED"], independent: true }, { action: "post", states: ["APPROVED"], independent: true }],
  "/api/v1/payables/invoices": [post],
  "/api/v1/payables/payments": [post],
  "/api/v1/inventory/adjustments": [{ ...post, independent: true }],
  "/api/v1/inventory/counts": [{ action: "reconcile", states: ["PLANNED", "IN_PROGRESS", "COMPLETED"], independent: true }],
  "/api/v1/inventory/transfers": [{ action: "dispatch", states: ["DRAFT"] }, { action: "receive", states: ["DISPATCHED", "IN_TRANSIT"] }],
  "/api/v1/fixedassets/depreciation/runs": [{ action: "post", states: ["DRAFT", "CALCULATED"], independent: true }],
  "/api/v1/sales/invoices": [{ action: "submit", states: ["DRAFT"] }, { action: "post", states: ["APPROVED"], independent: true }],
  "/api/v1/sales/credit-notes": [post],
  "/api/v1/sales/orders": [{ action: "confirm", states: ["DRAFT"] }],
  "/api/v1/sales/deliveries": [{ action: "dispatch", states: ["DRAFT"] }],
};
export function availableActions(path: string, row: RecordRow) {
  return row.id ? (rules[path] ?? []).filter(rule => rule.states.includes(String(row.status))) : [];
}
export function ResourceActions({ path, row, legalEntityId, onSaved }: { path: string; row: RecordRow; legalEntityId: string | null; onSaved: () => Promise<void> }) {
  const { t } = useLocale();
  const { user } = useAuth();
  const copy = t.business;
  const [confirm, setConfirm] = useState<Rule | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const actions = availableActions(path, row);
  const creator = row.created_by_user_id ?? row.created_by ?? row.initiated_by_user_id;
  const isOwn = creator === user?.id;
  if (!actions.length) return null;
  const execute = async () => {
    if (!confirm) return;
    setBusy(true); setError(null);
    try {
      await apiFetch(queryPath(`${path}/${encodeURIComponent(String(row.id))}/${confirm.action}`, legalEntityId ? { legal_entity_id: legalEntityId } : {}), { method: "POST", body: "{}" });
      setConfirm(null); await onSaved();
    } catch (failure) { setError(failure instanceof Error ? failure.message : t.resourceForm.saveError); }
    finally { setBusy(false); }
  };
  return <div className="mb-5 space-y-3 border-b pb-4" aria-label={copy.actions}>
    <div className="flex flex-wrap gap-2">{actions.map(rule => <button key={rule.action} type="button" disabled={busy || (rule.independent && isOwn) || (rule.action === "submit" && !isOwn)} onClick={() => { setConfirm(rule); setError(null); }} className="rounded-lg border bg-primary px-4 py-2 text-sm text-primary-foreground disabled:opacity-50">{copy[rule.action]}</button>)}</div>
    {isOwn && actions.some(rule => rule.independent) && <p className="text-xs text-muted-foreground">{copy.ownDocument}</p>}
    {confirm && <div className="space-y-3 rounded-lg border bg-muted/40 p-4"><h3 className="font-medium">{copy.confirmAction}: {copy[confirm.action]}</h3><p className="text-sm">{copy.confirmHint}</p><div className="flex gap-3"><button type="button" disabled={busy} onClick={() => void execute()} className="rounded-lg bg-primary px-4 py-2 text-sm text-primary-foreground disabled:opacity-50">{busy ? copy.working : copy.confirmAction}</button><button type="button" disabled={busy} onClick={() => setConfirm(null)} className="rounded-lg border px-4 py-2 text-sm">{copy.cancel}</button></div></div>}
    {error && <p role="alert" className="text-sm text-destructive">{error}</p>}
  </div>;
}
