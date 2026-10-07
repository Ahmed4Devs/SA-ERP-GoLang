"use client";

import { useCallback, useEffect, useRef, useState, type FormEvent } from "react";
import Link from "next/link";
import { Plus, RefreshCw, X } from "lucide-react";
import { useLocale } from "@/components/layout/shell";
import { apiFetch } from "@/lib/api-client";
import { useAuth } from "@/lib/auth-context";
import ar from "@/messages/sales-invoice.ar.json";
import en from "@/messages/sales-invoice.en.json";
import { emptyLine, field, invoiceID, invoiceInputsRounded, invoiceLinePayload, rows, validLineAmounts, type InvoiceLineInput, type Row } from "./invoice-model";

type Copy = typeof en;
const inputClass = "h-10 w-full rounded-lg border bg-background px-3 text-sm disabled:opacity-60";
const buttonClass = "inline-flex items-center justify-center gap-2 rounded-lg border px-4 py-2 text-sm font-semibold hover:bg-muted disabled:opacity-50";
const primaryClass = "inline-flex items-center justify-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground disabled:opacity-50";
const invoicePath = "/api/v1/sales/invoices";

function scoped(path: string, entity: string) {
  return `${path}${path.includes("?") ? "&" : "?"}legal_entity_id=${encodeURIComponent(entity)}`;
}

function unwrap(payload: Row): Row {
  return payload.data && typeof payload.data === "object" && !Array.isArray(payload.data) ? payload.data as Row : payload;
}

function displayName(row: Row, locale: string): string {
  return String((locale === "ar" ? row.name_ar || row.name_en : row.name_en || row.name_ar) || row.code || row.id);
}

export function InvoiceWorkspace() {
  const { locale } = useLocale();
  const { legalEntityId } = useAuth();
  const copy = locale === "ar" ? ar : en;
  if (!legalEntityId) return <p className="rounded-xl border bg-card p-5">{copy.selectEntity}</p>;
  // Entity changes discard the previous entity's drafts and in-flight requests.
  return <ScopedInvoiceWorkspace key={legalEntityId} entity={legalEntityId} copy={copy} locale={locale} />;
}

function ScopedInvoiceWorkspace({ entity, copy, locale }: { entity: string; copy: Copy; locale: string }) {
  const { user } = useAuth();
  const [invoices, setInvoices] = useState<Row[]>([]);
  const [selected, setSelected] = useState<Row | null>(null);
  const [creating, setCreating] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const [page, setPage] = useState(0);
  const [revision, setRevision] = useState(0);
  const [confirmed, setConfirmed] = useState(false);
  const detailController = useRef<AbortController | null>(null);
  const [posting, setPosting] = useState(false);

  useEffect(() => {
    const abort = new AbortController();
    setBusy(true);
    setError("");
    apiFetch<unknown>(scoped(`${invoicePath}?limit=20&offset=${page * 20}`, entity), { signal: abort.signal })
      .then((payload) => { if (!abort.signal.aborted) setInvoices(rows(payload)); })
      .catch((cause) => { if (!abort.signal.aborted) setError(cause instanceof Error ? cause.message : copy.error); })
      .finally(() => { if (!abort.signal.aborted) setBusy(false); });
    return () => abort.abort();
  }, [entity, page, revision, copy.error]);

  useEffect(() => () => detailController.current?.abort(), []);

  const open = async (id: string) => {
    detailController.current?.abort();
    const abort = new AbortController();
    detailController.current = abort;
    setError(""); setNotice(""); setSelected(null); setConfirmed(false); setCreating(false);
    try {
      const result = unwrap(await apiFetch<Row>(scoped(`${invoicePath}/${encodeURIComponent(id)}`, entity), { signal: abort.signal }));
      if (!abort.signal.aborted) setSelected(result);
    } catch (cause) {
      if (!abort.signal.aborted) setError(cause instanceof Error ? cause.message : copy.error);
    }
  };

  const transition = async (action: "submit" | "post") => {
    if (!selected || !confirmed || posting) return;
    setPosting(true); setError(""); setNotice("");
    try {
      const result = unwrap(await apiFetch<Row>(scoped(`${invoicePath}/${encodeURIComponent(invoiceID(selected))}/${action}`, entity), {
        method: "POST", headers: { "Idempotency-Key": `sales-invoice-${action}-${invoiceID(selected)}` }, body: "{}",
      }));
      setSelected(result); setConfirmed(false); setNotice(action === "submit" ? copy.submitted : copy.posted); setRevision((value) => value + 1);
    } catch (cause) { setError(cause instanceof Error ? cause.message : copy.error); }
    finally { setPosting(false); }
  };

  const statusLabel = (status: string) => status === "DRAFT" ? copy.draft : status === "POSTED" ? copy.postedStatus : status === "SUBMITTED" ? copy.submittedStatus : status === "APPROVED" ? copy.approvedStatus : status === "REJECTED" ? copy.rejectedStatus : status;

  return <section className="space-y-5" aria-labelledby="sales-invoice-title">
    <header className="flex flex-wrap items-start justify-between gap-4 rounded-2xl border bg-card p-6">
      <div><h1 id="sales-invoice-title" className="text-2xl font-bold">{copy.title}</h1><p className="mt-2 text-sm text-muted-foreground">{copy.description}</p></div>
      <div className="flex flex-wrap gap-2">
        <button className={primaryClass} onClick={() => { detailController.current?.abort(); setCreating(true); setSelected(null); setNotice(""); setError(""); }} disabled={posting}><Plus className="h-4 w-4" />{copy.newInvoice}</button>
        <button className={buttonClass} onClick={() => setRevision((value) => value + 1)} disabled={busy}><RefreshCw className="h-4 w-4" />{copy.refresh}</button>
      </div>
    </header>
    {error && <p role="alert" className="rounded-lg border border-destructive/40 bg-destructive/10 p-3 text-sm text-destructive">{error}</p>}
    {notice && <p role="status" className="rounded-lg border border-emerald-500/40 bg-emerald-500/10 p-3 text-sm">{notice}</p>}
    {creating && <InvoiceForm entity={entity} locale={locale} copy={copy} onCancel={() => setCreating(false)} onSaved={(invoice, rounded) => {
      setCreating(false); setSelected(invoice); setConfirmed(false); setNotice(rounded ? copy.savedRounded : copy.saved); setPage(0); setRevision((value) => value + 1);
    }} />}
    {selected && <section aria-label={copy.review} className="space-y-5 rounded-xl border bg-card p-5">
      <div className="flex items-center justify-between gap-3"><h2 className="text-lg font-bold">{copy.review}: <bdi>{field(selected, "InvoiceNumber")}</bdi></h2><button aria-label={copy.close} disabled={posting} className={buttonClass} onClick={() => setSelected(null)}><X className="h-4 w-4" /></button></div>
      <dl className="grid gap-4 sm:grid-cols-3">
        <Value label={copy.customer} value={field(selected, locale === "ar" ? "CustomerNameAr" : "CustomerNameEn")} />
        <Value label={copy.issueDate} value={field(selected, "IssueDate").slice(0, 10)} />
        <Value label={copy.dueDate} value={field(selected, "DueDate").slice(0, 10)} />
        <Value label={copy.status} value={statusLabel(field(selected, "Status"))} />
        <Value label={copy.currency} value={field(selected, "CurrencyCode")} />
        <Value label={copy.journal} value={field(selected, "JournalEntryID")} />
      </dl>
      <div className="overflow-x-auto"><table className="w-full text-start text-sm"><thead className="bg-muted"><tr>{[copy.lineDescription, copy.quantity, copy.uom, copy.unitPrice, copy.taxCode, copy.taxRate, copy.taxAmount, copy.netAmount].map((label) => <th key={label} className="whitespace-nowrap p-3 text-start">{label}</th>)}</tr></thead>
        <tbody>{rows(selected.Lines ?? selected.lines).map((line, index) => <tr key={index} className="border-b">{["Description", "Quantity", "UOM", "UnitPrice", "TaxCode", "TaxRatePercent", "TaxAmount", "NetAmount"].map((key) => <td key={key} className="p-3"><bdi>{field(line, key) || "—"}</bdi></td>)}</tr>)}</tbody></table></div>
      <p className="text-sm text-muted-foreground">{copy.serverTotals}</p>
      <dl className="grid gap-3 rounded-lg bg-muted/40 p-4 sm:grid-cols-2 lg:grid-cols-3">
        {([ [copy.subtotal, "Subtotal"], [copy.discountAmount, "DiscountAmount"], [copy.taxableAmount, "TaxableAmount"], [copy.taxAmount, "TaxAmount"], [copy.total, "TotalAmount"], [copy.exchangeRate, "ExchangeRate"] ] as const).map(([label, key]) => <Value key={key} label={label} value={field(selected, key)} />)}
      </dl>
      {field(selected, "Status") === "DRAFT" && <div className="space-y-3 border-t pt-4">
        {field(selected, "CreatedByUserID") !== user?.id && <p className="text-sm text-muted-foreground">{copy.makerOnly}</p>}
        <label className="flex items-start gap-3 text-sm"><input type="checkbox" className="mt-1" checked={confirmed} onChange={event => setConfirmed(event.target.checked)} disabled={posting || field(selected, "CreatedByUserID") !== user?.id} />{copy.submitConfirm}</label>
        <button className={primaryClass} onClick={() => void transition("submit")} disabled={!confirmed || posting || field(selected, "CreatedByUserID") !== user?.id}>{posting ? copy.submitting : copy.submit}</button>
      </div>}
      {field(selected, "Status") === "SUBMITTED" && <div className="space-y-2 border-t pt-4"><p className="text-sm text-muted-foreground">{copy.waitingApproval}</p><Link href="/workflow" className="text-sm text-primary underline">{copy.openInbox}</Link></div>}
      {field(selected, "Status") === "APPROVED" && <div className="space-y-3 border-t pt-4">
        {field(selected, "CreatedByUserID") === user?.id && <p className="text-sm text-muted-foreground">{copy.independentPost}</p>}
        <label className="flex items-start gap-3 text-sm"><input type="checkbox" className="mt-1" checked={confirmed} onChange={event => setConfirmed(event.target.checked)} disabled={posting || field(selected, "CreatedByUserID") === user?.id} />{copy.postConfirm}</label>
        <button className={primaryClass} onClick={() => void transition("post")} disabled={!confirmed || posting || field(selected, "CreatedByUserID") === user?.id}>{posting ? copy.posting : copy.post}</button>
      </div>}
    </section>}
    <div className="overflow-hidden rounded-xl border bg-card">
      {busy ? <p role="status" className="p-6 text-sm">{copy.loading}</p> : invoices.length === 0 ? <p className="p-6 text-sm text-muted-foreground">{copy.empty}</p> : <div className="overflow-x-auto"><table className="w-full text-sm"><thead className="bg-muted/50"><tr>{[copy.number, copy.customer, copy.issueDate, copy.status, copy.total, ""].map((label, index) => <th key={index} className="whitespace-nowrap p-3 text-start">{label}</th>)}</tr></thead><tbody>{invoices.map((invoice) => <tr key={invoiceID(invoice)} className="border-t">
        <td className="p-3"><bdi>{field(invoice, "InvoiceNumber")}</bdi></td><td className="p-3">{field(invoice, locale === "ar" ? "CustomerNameAr" : "CustomerNameEn")}</td><td className="whitespace-nowrap p-3"><bdi>{field(invoice, "IssueDate").slice(0, 10)}</bdi></td><td className="p-3">{statusLabel(field(invoice, "Status"))}</td><td className="whitespace-nowrap p-3"><bdi>{field(invoice, "TotalAmount")} {field(invoice, "CurrencyCode")}</bdi></td><td className="p-3"><button className={buttonClass} disabled={posting} onClick={() => void open(invoiceID(invoice))}>{copy.open}</button></td>
      </tr>)}</tbody></table></div>}
      <div className="flex items-center justify-between gap-2 border-t p-3 text-sm"><button className={buttonClass} disabled={busy || page === 0} onClick={() => setPage((value) => value - 1)}>{copy.previous}</button><span>{copy.page} {page + 1}</span><button className={buttonClass} disabled={busy || invoices.length < 20} onClick={() => setPage((value) => value + 1)}>{copy.next}</button></div>
    </div>
  </section>;
}

function Value({ label, value }: { label: string; value: string }) {
  return <div className="min-w-0"><dt className="text-xs text-muted-foreground">{label}</dt><dd className="mt-1 break-words text-sm font-medium"><bdi>{value || "—"}</bdi></dd></div>;
}

async function allPages(path: string, signal: AbortSignal): Promise<Row[]> {
  const result: Row[] = [];
  for (let offset = 0; ; offset += 100) {
    const batch = rows(await apiFetch<unknown>(`${path}&limit=100&offset=${offset}`, { signal }));
    result.push(...batch);
    if (batch.length < 100) return result;
  }
}

function InvoiceForm({ entity, locale, copy, onCancel, onSaved }: { entity: string; locale: string; copy: Copy; onCancel: () => void; onSaved: (invoice: Row, rounded: boolean) => void }) {
  const [customers, setCustomers] = useState<Row[]>([]);
  const [items, setItems] = useState<Row[]>([]);
  const [uoms, setUoms] = useState<Row[]>([]);
  const [taxCodes, setTaxCodes] = useState<Row[]>([]);
  const [loading, setLoading] = useState(true);
  const [lookupError, setLookupError] = useState(false);
  const [revision, setRevision] = useState(0);
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);
  const [customerID, setCustomerID] = useState("");
  const [invoiceType, setInvoiceType] = useState("STANDARD");
  const [issueDate, setIssueDate] = useState(() => new Intl.DateTimeFormat("en-CA", { timeZone: "Asia/Riyadh", year: "numeric", month: "2-digit", day: "2-digit" }).format(new Date()));
  const [dueDate, setDueDate] = useState(issueDate);
  const [currency, setCurrency] = useState("SAR");
  const [exchangeRate, setExchangeRate] = useState("1");
  const [notes, setNotes] = useState("");
  const [lines, setLines] = useState<InvoiceLineInput[]>(() => [emptyLine()]);
  const requestController = useRef<AbortController | null>(null);
  useEffect(() => () => requestController.current?.abort(), []);

  useEffect(() => {
    const controller = new AbortController();
    setLoading(true); setLookupError(false);
    Promise.all([
      allPages(scoped("/api/v1/receivables/customers", entity), controller.signal),
      allPages(scoped("/api/v1/products/items", entity), controller.signal),
      apiFetch<unknown>(scoped("/api/v1/products/uoms", entity), { signal: controller.signal }).then(rows),
      apiFetch<unknown>(scoped("/api/v1/tax/codes", entity), { signal: controller.signal }).then(rows),
    ]).then(([loadedCustomers, loadedItems, loadedUoms, loadedTax]) => {
      if (controller.signal.aborted) return;
      setCustomers(loadedCustomers.filter((row) => row.status === "ACTIVE" && !row.is_credit_hold));
      setItems(loadedItems.filter((row) => row.is_active !== false && row.is_sales_enabled !== false));
      setUoms(loadedUoms.filter((row) => row.is_active !== false));
      setTaxCodes(loadedTax.filter((row) => row.is_active !== false && row.tax_type !== "INPUT"));
    }).catch(() => { if (!controller.signal.aborted) setLookupError(true); })
      .finally(() => { if (!controller.signal.aborted) setLoading(false); });
    return () => controller.abort();
  }, [entity, revision]);

  const updateLine = useCallback((key: string, values: Partial<InvoiceLineInput>) => setLines((current) => current.map((line) => line.key === key ? { ...line, ...values } : line)), []);
  const chooseItem = (line: InvoiceLineInput, id: string) => {
    const item = items.find((row) => row.id === id);
    if (!item) return updateLine(line.key, { item_id: "", uom: "", description: "" });
    const uom = uoms.find((row) => row.id === (item.sales_uom_id || item.base_uom_id));
    const tax = taxCodes.find((row) => row.id === item.default_tax_code_id || row.code === item.tax_code);
    updateLine(line.key, { item_id: id, description: displayName(item, locale), uom: String(uom?.code ?? ""), unit_price: String(item.list_price ?? "0"), tax_code: String(tax?.code ?? "") });
  };
  const missingOptions = !customers.length || !items.length || !uoms.length || !taxCodes.length;
  const submit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault(); setError("");
    const customer = customers.find((row) => row.id === customerID);
    if (!customer || lines.some((line) => !line.item_id || !line.description.trim() || !line.uom || !line.tax_code)) return setError(copy.required);
    if (dueDate < issueDate) return setError(copy.invalidDates);
    if (lines.some((line) => !validLineAmounts(line)) || !/^(?:0|[1-9]\d{0,11})(?:\.\d{1,12})?$/.test(exchangeRate) || !/[1-9]/.test(exchangeRate)) return setError(copy.invalidDecimal);
    if (invoiceType === "STANDARD" && !/^3\d{13}3$/.test(String(customer.tax_id ?? ""))) return setError(copy.invalidVAT);
    const abort = new AbortController(); requestController.current = abort;
    setSaving(true);
    try {
      const payload = { legal_entity_id: entity, customer_id: customer.id, customer_name_ar: customer.name_ar, customer_name_en: customer.name_en,
        customer_tax_id: customer.tax_id ?? null, customer_cr_number: customer.cr_number ?? null, customer_address_json: JSON.stringify(customer.address ?? {}),
        payment_terms_id: customer.payment_terms_id ?? null, invoice_type: invoiceType, issue_date: issueDate, due_date: dueDate,
        currency_code: currency, exchange_rate: exchangeRate, notes: notes.trim() || null, lines: lines.map(invoiceLinePayload) };
      const result = unwrap(await apiFetch<Row>(scoped(invoicePath, entity), { method: "POST", body: JSON.stringify(payload), signal: abort.signal }));
      if (!abort.signal.aborted) onSaved(result, invoiceInputsRounded(lines, result));
    } catch (cause) { if (!abort.signal.aborted) setError(cause instanceof Error ? cause.message : copy.error); }
    finally { if (!abort.signal.aborted) setSaving(false); }
  };

  return <form onSubmit={(event) => void submit(event)} aria-label={copy.newInvoice} className="space-y-5 rounded-xl border bg-card p-5">
    <h2 className="text-lg font-bold">{copy.newInvoice}</h2>
    {loading && <p role="status">{copy.loading}</p>}
    {lookupError && <div role="alert" className="space-y-3 text-destructive"><p>{copy.lookupError}</p><button type="button" className={buttonClass} onClick={() => setRevision((value) => value + 1)}>{copy.retry}</button></div>}
    {!loading && !lookupError && missingOptions && <p role="alert">{copy.noOptions}</p>}
    {error && <p role="alert" className="text-sm text-destructive">{error}</p>}
    <fieldset disabled={loading || lookupError || missingOptions || saving} className="space-y-5">
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <label className="space-y-1 text-sm"><span>{copy.customer}</span><select required className={inputClass} value={customerID} onChange={(event) => { setCustomerID(event.target.value); const customer = customers.find((row) => row.id === event.target.value); setCurrency(String(customer?.currency_code || "SAR")); }}><option value="">{copy.select}</option>{customers.map((row) => <option key={String(row.id)} value={String(row.id)}>{displayName(row, locale)}</option>)}</select></label>
        <label className="space-y-1 text-sm"><span>{copy.type}</span><select className={inputClass} value={invoiceType} onChange={(event) => setInvoiceType(event.target.value)}><option value="STANDARD">{copy.standard}</option><option value="SIMPLIFIED">{copy.simplified}</option></select></label>
        <label className="space-y-1 text-sm"><span>{copy.issueDate}</span><input type="date" required className={inputClass} value={issueDate} onChange={(event) => setIssueDate(event.target.value)} /></label>
        <label className="space-y-1 text-sm"><span>{copy.dueDate}</span><input type="date" required min={issueDate} className={inputClass} value={dueDate} onChange={(event) => setDueDate(event.target.value)} /></label>
        <label className="space-y-1 text-sm"><span>{copy.currency}</span><input required pattern="[A-Z]{3}" maxLength={3} dir="ltr" className={inputClass} value={currency} onChange={(event) => setCurrency(event.target.value.toUpperCase())} /></label>
        <label className="space-y-1 text-sm"><span>{copy.exchangeRate}</span><input required inputMode="decimal" dir="ltr" className={inputClass} value={exchangeRate} onChange={(event) => setExchangeRate(event.target.value)} /></label>
      </div>
      <div className="space-y-4"><h3 className="font-semibold">{copy.lines}</h3>{lines.map((line, index) => <fieldset key={line.key} className="space-y-4 rounded-lg border p-4"><legend className="px-2 text-sm font-semibold">{copy.line} {index + 1}</legend>
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <label className="space-y-1 text-sm sm:col-span-2"><span>{copy.item}</span><select required className={inputClass} value={line.item_id} onChange={(event) => chooseItem(line, event.target.value)}><option value="">{copy.select}</option>{items.map((row) => <option key={String(row.id)} value={String(row.id)}>{String(row.code)} · {displayName(row, locale)}</option>)}</select></label>
          <label className="space-y-1 text-sm"><span>{copy.uom}</span><select required className={inputClass} value={line.uom} onChange={(event) => updateLine(line.key, { uom: event.target.value })}><option value="">{copy.select}</option>{uoms.map((row) => <option key={String(row.id)} value={String(row.code)}>{displayName(row, locale)}</option>)}</select></label>
          <label className="space-y-1 text-sm"><span>{copy.taxCode}</span><select required className={inputClass} value={line.tax_code} onChange={(event) => updateLine(line.key, { tax_code: event.target.value })}><option value="">{copy.select}</option>{taxCodes.map((row) => <option key={String(row.id)} value={String(row.code)}>{String(row.code)} · {displayName(row, locale)}</option>)}</select></label>
          <label className="space-y-1 text-sm sm:col-span-2"><span>{copy.lineDescription}</span><input required maxLength={500} className={inputClass} value={line.description} onChange={(event) => updateLine(line.key, { description: event.target.value })} /></label>
          {([ ["quantity", copy.quantity], ["unit_price", copy.unitPrice], ["discount_percent", copy.discount] ] as const).map(([key, label]) => <label key={key} className="space-y-1 text-sm"><span>{label}</span><input required inputMode="decimal" dir="ltr" className={inputClass} value={line[key]} onChange={(event) => updateLine(line.key, { [key]: event.target.value })} /></label>)}
          <label className="space-y-1 text-sm sm:col-span-2"><span>{copy.exemptionReason}</span><input maxLength={500} className={inputClass} value={line.exemption_reason} onChange={(event) => updateLine(line.key, { exemption_reason: event.target.value })} /></label>
        </div>
        <button type="button" className={buttonClass} disabled={lines.length === 1} onClick={() => setLines((current) => current.filter((row) => row.key !== line.key))}>{copy.removeLine}</button>
      </fieldset>)}<button type="button" className={buttonClass} onClick={() => setLines((current) => [...current, emptyLine()])}><Plus className="h-4 w-4" />{copy.addLine}</button></div>
      <label className="block space-y-1 text-sm"><span>{copy.notes}</span><textarea className="w-full rounded-lg border bg-background p-3" maxLength={2000} value={notes} onChange={(event) => setNotes(event.target.value)} /></label>
    </fieldset>
    <div className="flex flex-wrap gap-2"><button type="submit" className={primaryClass} disabled={loading || lookupError || missingOptions || saving}>{saving ? copy.saving : copy.save}</button><button type="button" className={buttonClass} disabled={saving} onClick={onCancel}>{copy.cancel}</button></div>
  </form>;
}
