"use client";

import { useEffect, useState } from "react";
import { useLocale } from "@/components/layout/shell";
import { useAuth } from "@/lib/auth-context";
import { apiFetch } from "@/lib/api-client";
import { normalizeRows, queryPath, type RecordRow } from "@/lib/resource-data";
import { RecordDetails } from "@/components/modules/record-details";

const button = "rounded-lg border px-4 py-2 text-sm hover:bg-muted disabled:opacity-50";
const primary = "rounded-lg bg-primary px-4 py-2 text-sm text-primary-foreground disabled:opacity-50";
const control = "min-h-10 w-full rounded-lg border bg-background px-3 py-2 text-sm disabled:opacity-50";
const documentPaths: Record<string, string> = { SALES_INVOICE: "/api/v1/sales/invoices", TREASURY_PAYMENT: "/api/v1/treasury/payments", PAYROLL_RUN: "/api/v1/payroll/runs" };

export function ApprovalWorkspace() {
  const { legalEntityId } = useAuth();
  return <ScopedApprovals key={legalEntityId} entity={legalEntityId} />;
}
function ScopedApprovals({ entity }: { entity: string | null }) {
  const { locale, t } = useLocale();
  const { user } = useAuth();
  const copy = t.approval;
  const [items, setItems] = useState<RecordRow[]>([]);
  const [selected, setSelected] = useState<RecordRow | null>(null);
  const [decision, setDecision] = useState<"APPROVE" | "REJECT" | null>(null);
  const [notes, setNotes] = useState("");
  const [revision, setRevision] = useState(0);
  const [loading, setLoading] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const [limitsOpen, setLimitsOpen] = useState(false);
  useEffect(() => {
    const controller = new AbortController(); setLoading(true); setError("");
    apiFetch<unknown>("/api/v1/workflow/inbox", { signal: controller.signal })
      .then(payload => { if (!controller.signal.aborted) setItems(normalizeRows(payload).filter(row => !entity || row.legal_entity_id === entity)); })
      .catch(failure => { if (!controller.signal.aborted) setError(failure instanceof Error ? failure.message : t.resourceForm.loadError); })
      .finally(() => { if (!controller.signal.aborted) setLoading(false); });
    return () => controller.abort();
  }, [entity, revision, t.resourceForm.loadError]);
  async function decide() {
    if (!selected || !decision) return;
    setBusy(true); setError(""); setNotice("");
    try {
      await apiFetch(queryPath(`/api/v1/workflow/requests/${encodeURIComponent(String(selected.request_id))}/decide`, { legal_entity_id: String(selected.legal_entity_id) }), { method: "POST", body: JSON.stringify({ decision, notes: notes.trim() }) });
      setSelected(null); setDecision(null); setNotes(""); setNotice(copy.saved); setRevision(value => value + 1);
    } catch (failure) { setError(failure instanceof Error ? failure.message : t.resourceForm.saveError); }
    finally { setBusy(false); }
  }
  return <section className="space-y-4 rounded-xl border bg-card p-5" aria-label={copy.title}>
    <div className="flex flex-wrap justify-between gap-3"><h1 className="text-xl font-bold">{copy.title}</h1><div className="flex gap-2"><button type="button" className={button} disabled={loading || busy} onClick={() => setRevision(value => value + 1)}>{copy.refresh}</button><button type="button" className={button} onClick={() => setLimitsOpen(value => !value)}>{copy.manage}</button></div></div>
    {error && <p role="alert" className="text-sm text-destructive">{error}</p>}
    {notice && <p role="status" className="text-sm text-emerald-700 dark:text-emerald-300">{notice}</p>}
    {loading ? <p className="text-sm">{copy.loading}</p> : !items.length ? <p className="text-sm text-muted-foreground">{copy.empty}</p> : <div className="overflow-x-auto"><table className="w-full text-sm"><thead className="bg-muted"><tr>{[copy.document, copy.initiator, copy.step, copy.review].map(label => <th key={label} className="p-3 text-start">{label}</th>)}</tr></thead><tbody>{items.map(item => <tr key={String(item.request_id)} data-document-id={String(item.resource_id)} className="border-t"><td className="p-3">{item.resource_type === "SALES_INVOICE" ? copy.sales : item.resource_type === "TREASURY_PAYMENT" ? copy.treasury : item.resource_type === "PAYROLL_RUN" ? copy.payroll : String(item.resource_type)}</td><td className="p-3">{String(item.initiator_name)}</td><td className="p-3">{String((locale === "ar" ? item.step_name_ar : item.step_name_en) ?? "")}</td><td className="p-3"><button type="button" className={button} disabled={busy} onClick={() => { setSelected(item); setDecision(null); setNotes(""); setNotice(""); }}>{copy.review}</button></td></tr>)}</tbody></table></div>}
    {selected && <section aria-label={copy.reviewDocument} className="space-y-4 rounded-lg border p-4"><div className="flex justify-between"><h2 className="font-semibold">{copy.reviewDocument}</h2><button type="button" className={button} disabled={busy} onClick={() => { setSelected(null); setDecision(null); }}>{copy.close}</button></div>
      <RecordDetails key={String(selected.request_id)} row={{ ...(selected.context_data && typeof selected.context_data === "object" ? selected.context_data as RecordRow : {}), id: selected.resource_id }} path={documentPaths[String(selected.resource_type)] ?? ""} legalEntityId={String(selected.legal_entity_id)} />
      <label className="block space-y-1 text-sm">{copy.notes}<textarea className={control} value={notes} maxLength={2000} disabled={busy} onChange={event => setNotes(event.target.value)} /></label>
      <div className="flex flex-wrap gap-3"><button type="button" className={primary} disabled={busy || selected.initiator_id === user?.id} onClick={() => setDecision("APPROVE")}>{copy.approve}</button><button type="button" className={button} disabled={busy || selected.initiator_id === user?.id} onClick={() => setDecision("REJECT")}>{copy.reject}</button></div>
      {selected.initiator_id === user?.id && <p className="text-sm text-muted-foreground">{t.business.ownDocument}</p>}
      {decision && <div className="space-y-3 rounded-lg bg-muted p-4"><p>{decision === "APPROVE" ? copy.approve : copy.reject}</p><div className="flex gap-3"><button type="button" className={primary} disabled={busy} onClick={() => void decide()}>{busy ? t.business.working : copy.confirm}</button><button type="button" className={button} disabled={busy} onClick={() => setDecision(null)}>{t.business.cancel}</button></div></div>}
    </section>}
    {limitsOpen && entity && <ApprovalLimits entity={entity} />}
  </section>;
}

function ApprovalLimits({ entity }: { entity: string }) {
  const { t } = useLocale(); const copy = t.approval;
  const [limits, setLimits] = useState<RecordRow[]>([]);
  const [users, setUsers] = useState<RecordRow[]>([]);
  const [userId, setUserId] = useState("");
  const [type, setType] = useState("SALES_INVOICE");
  const [amount, setAmount] = useState("");
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const [revision, setRevision] = useState(0);
  const existing = limits.find(item => item.user_id === userId && item.resource_type === type);
  useEffect(() => {
    const controller = new AbortController(); setLoading(true); setError("");
    async function load() {
      try {
        const limits = normalizeRows(await apiFetch<unknown>(queryPath("/api/v1/workflow/approval-limits", { legal_entity_id: entity }), { signal: controller.signal }));
        const allUsers: RecordRow[] = [];
        for (let offset = 0; ; offset += 100) {
          const batch = normalizeRows(await apiFetch<unknown>(`/api/v1/auth/users?limit=100&offset=${offset}`, { signal: controller.signal }));
          allUsers.push(...batch); if (batch.length < 100) break;
        }
        if (!controller.signal.aborted) { setLimits(limits); setUsers(allUsers.filter(user => user.status === "ACTIVE")); }
      } catch (failure) { if (!controller.signal.aborted) setError(failure instanceof Error ? failure.message : t.resourceForm.loadError); }
      finally { if (!controller.signal.aborted) setLoading(false); }
    }
    void load(); return () => controller.abort();
  }, [entity, revision, t.resourceForm.loadError]);
  useEffect(() => { setAmount(String(existing?.maximum_base_amount ?? "")); }, [existing]);
  async function save(event: React.FormEvent) {
    event.preventDefault(); setError(""); setNotice("");
    if (!/^\d{1,18}(\.\d{1,6})?$/.test(amount) || !users.some(user => user.id === userId)) { setError(t.resourceForm.invalid); return; }
    setBusy(true);
    try {
      await apiFetch("/api/v1/workflow/approval-limits", { method: "PUT", body: JSON.stringify({ legal_entity_id: entity, user_id: userId, resource_type: type, maximum_base_amount: amount, version: Number(existing?.version ?? 0) }) });
      setNotice(copy.limitSaved); setRevision(value => value + 1);
    } catch (failure) { setError(failure instanceof Error ? failure.message : t.resourceForm.saveError); }
    finally { setBusy(false); }
  }
  return <section className="space-y-4 border-t pt-4" aria-label={copy.limits}><h2 className="font-semibold">{copy.limits}</h2><p className="text-sm text-muted-foreground">{copy.limitDescription}</p>
    {error && <p role="alert" className="text-sm text-destructive">{error}</p>}{notice && <p role="status" className="text-sm text-emerald-700 dark:text-emerald-300">{notice}</p>}
    <form onSubmit={event => void save(event)} className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <label className="space-y-1 text-sm">{copy.user}<select aria-label={copy.user} required disabled={loading || busy} className={control} value={userId} onChange={event => setUserId(event.target.value)}><option value="">{copy.selectUser}</option>{users.map(user => <option key={String(user.id)} value={String(user.id)}>{String(user.display_name ?? user.username)}</option>)}</select></label>
      <label className="space-y-1 text-sm">{copy.type}<select aria-label={copy.type} className={control} disabled={busy} value={type} onChange={event => setType(event.target.value)}><option value="SALES_INVOICE">{copy.sales}</option><option value="TREASURY_PAYMENT">{copy.treasury}</option><option value="PAYROLL_RUN">{copy.payroll}</option></select></label>
      <label className="space-y-1 text-sm">{copy.maximum}<input aria-label={copy.maximum} required inputMode="decimal" dir="ltr" className={control} value={amount} disabled={loading || busy} onChange={event => setAmount(event.target.value)} /></label>
      <div className="col-span-full flex items-center gap-4"><button type="submit" className={primary} disabled={loading || busy || !userId}>{busy ? t.business.working : copy.save}</button>{existing && <span className="text-xs text-muted-foreground">{copy.version} {String(existing.version)}</span>}</div>
    </form>
  </section>;
}
