"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { Download, FileUp, RefreshCw } from "lucide-react";
import { useLocale } from "@/components/layout/shell";
import { useAuth } from "@/lib/auth-context";
import { APIError, apiFetch, apiURL } from "@/lib/api-client";

interface DocumentRecord {
  id: string;
  title: string;
  category: string;
  current_version?: { version_number: number; original_filename: string; byte_size: number };
}

export function DocumentLibrary() {
  const { t } = useLocale();
  const copy = t.documentLibrary;
  const { legalEntityId } = useAuth();
  const [documents, setDocuments] = useState<DocumentRecord[]>([]);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [search, setSearch] = useState("");
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);
  const [uploadOpen, setUploadOpen] = useState(false);
  const [versionOf, setVersionOf] = useState<DocumentRecord | null>(null);
  const [file, setFile] = useState<File | null>(null);
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [category, setCategory] = useState("GENERAL");
  const request = useRef<AbortController | null>(null);
  const mutation = useRef<AbortController | null>(null);
  const input = useRef<HTMLInputElement | null>(null);

  const load = useCallback(async () => {
    request.current?.abort();
    const controller = new AbortController();
    request.current = controller;
    setDocuments([]); setError(null); setTotal(0);
    if (!legalEntityId) return;
    setLoading(true);
    try {
      const params = new URLSearchParams({ legal_entity_id: legalEntityId, limit: "50", offset: String(offset), search: query });
      const result = await apiFetch<{ data: DocumentRecord[] | null; total: number }>(`/api/v1/documents/?${params}`, { signal: controller.signal });
      if (!controller.signal.aborted) { setDocuments(result.data ?? []); setTotal(result.total); }
    } catch (failure) { if (!controller.signal.aborted) setError(failure instanceof Error ? failure.message : copy.error); }
    finally { if (!controller.signal.aborted) setLoading(false); }
  }, [copy.error, legalEntityId, offset, query]);
  useEffect(() => { void load(); return () => request.current?.abort(); }, [load]);
  useEffect(() => {
    mutation.current?.abort();
    setUploadOpen(false); setFile(null); setTitle(""); setDescription(""); setCategory("GENERAL"); setVersionOf(null); setSuccess(false); setBusy(false); setOffset(0);
    return () => mutation.current?.abort();
  }, [legalEntityId]);

  async function upload(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!legalEntityId) return;
    setError(null); setSuccess(false);
    if (!file) { setError(copy.required); return; }
    if (file.size > 50 * 1024 * 1024) { setError(copy.sizeError); return; }
    const controller = new AbortController();
    mutation.current = controller;
    setBusy(true);
    const body = new FormData();
    body.set("file", file); body.set("legal_entity_id", legalEntityId);
    body.set("title", title.trim() || file.name); body.set("description", description.trim()); body.set("category", category);
    const path = versionOf ? `/api/v1/documents/${encodeURIComponent(versionOf.id)}/versions` : "/api/v1/documents/upload";
    try {
      await apiFetch(`${path}?legal_entity_id=${encodeURIComponent(legalEntityId)}`, { method: "POST", body, signal: controller.signal });
      if (controller.signal.aborted) return;
      setUploadOpen(false); setFile(null); setSuccess(true); await load();
    } catch (failure) { if (!controller.signal.aborted) setError(failure instanceof Error ? failure.message : copy.error); }
    finally { if (!controller.signal.aborted) setBusy(false); }
  }

  async function download(document: DocumentRecord) {
    if (!legalEntityId) return;
    const controller = new AbortController();
    mutation.current = controller;
    setBusy(true); setError(null);
    try {
      const response = await fetch(apiURL(`/api/v1/documents/${encodeURIComponent(document.id)}/download?legal_entity_id=${encodeURIComponent(legalEntityId)}`), { credentials: "include", cache: "no-store", signal: controller.signal });
      if (!response.ok) {
        if (response.status === 401) window.dispatchEvent(new CustomEvent("erp:session-expired"));
        throw new APIError(copy.error, response.status);
      }
      const blob = await response.blob();
      if (controller.signal.aborted) return;
      const url = URL.createObjectURL(blob);
      const link = window.document.createElement("a");
      link.href = url; link.download = document.current_version?.original_filename || document.title;
      window.document.body.appendChild(link); link.click(); link.remove();
      window.setTimeout(() => URL.revokeObjectURL(url), 1000);
    } catch (failure) { if (!controller.signal.aborted) setError(failure instanceof Error ? failure.message : copy.error); }
    finally { if (!controller.signal.aborted) setBusy(false); }
  }

  function openUpload(document: DocumentRecord | null) {
    setVersionOf(document); setTitle(document?.title ?? ""); setDescription(""); setCategory("GENERAL"); setFile(null); setUploadOpen(true); setError(null); setSuccess(false);
    if (input.current) input.current.value = "";
  }
  const control = "min-h-10 w-full rounded-lg border bg-background px-3 py-2 text-sm focus:ring-2 focus:ring-ring";
  return (
    <div className="space-y-5">
      <header className="flex flex-wrap items-center justify-between gap-4"><h1 className="text-2xl font-bold">{copy.title}</h1><div className="flex gap-2"><button type="button" onClick={() => openUpload(null)} disabled={busy || !legalEntityId} className="inline-flex items-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground disabled:opacity-50"><FileUp className="h-4 w-4" />{copy.upload}</button><button type="button" onClick={() => void load()} disabled={loading || busy} className="inline-flex items-center gap-2 rounded-lg border px-3 py-2 text-sm"><RefreshCw className="h-4 w-4" />{copy.refresh}</button></div></header>
      {!legalEntityId && <p role="alert">{copy.entityRequired}</p>}
      {error && <p role="alert" className="rounded-lg border border-destructive/40 bg-destructive/10 p-3 text-sm text-destructive">{error}</p>}
      {success && <p role="status" className="rounded-lg bg-emerald-500/10 p-3 text-sm text-emerald-700 dark:text-emerald-300">{copy.success}</p>}
      {uploadOpen && <form onSubmit={(event) => void upload(event)} className="space-y-4 rounded-xl border bg-card p-5">
        <h2 className="font-semibold">{versionOf ? `${copy.version}: ${versionOf.title}` : copy.upload}</h2>
        <label className="block space-y-2 text-sm font-medium">{copy.file}<input ref={input} aria-label={copy.file} type="file" required disabled={busy} onChange={(event) => setFile(event.target.files?.[0] ?? null)} className={control} /></label>
        {!versionOf && <div className="grid gap-4 sm:grid-cols-2"><label className="space-y-2 text-sm font-medium">{copy.name}<input aria-label={copy.name} value={title} onChange={(event) => setTitle(event.target.value)} disabled={busy} className={control} /></label><label className="space-y-2 text-sm font-medium">{copy.category}<select aria-label={copy.category} value={category} onChange={(event) => setCategory(event.target.value)} disabled={busy} className={control}>{Object.entries(copy.categories).map(([value, label]) => <option key={value} value={value}>{label}</option>)}</select></label><label className="space-y-2 text-sm font-medium sm:col-span-2">{copy.description}<textarea value={description} onChange={(event) => setDescription(event.target.value)} disabled={busy} className={control} rows={3} /></label></div>}
        <div className="flex gap-3"><button type="submit" disabled={busy || !file} className="rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground disabled:opacity-50">{busy ? copy.uploading : copy.save}</button><button type="button" disabled={busy} onClick={() => { setUploadOpen(false); setFile(null); }} className="rounded-lg border px-4 py-2 text-sm">{copy.cancel}</button></div>
      </form>}
      <form onSubmit={(event) => { event.preventDefault(); setQuery(search); setOffset(0); }} className="flex gap-2"><input aria-label={copy.search} placeholder={copy.search} value={search} onChange={(event) => setSearch(event.target.value)} className={`${control} max-w-md`} /><button type="submit" className="rounded-lg border px-4 py-2 text-sm">{copy.search}</button></form>
      <div className="overflow-x-auto rounded-xl border bg-card"><table className="w-full text-sm"><thead className="bg-muted"><tr><th className="p-3 text-start">{copy.name}</th><th className="p-3 text-start">{copy.category}</th><th className="p-3 text-start">{copy.currentVersion}</th><th className="p-3 text-start">{copy.file}</th></tr></thead><tbody className="divide-y">{loading || documents.length === 0 ? <tr><td colSpan={4} className="p-8 text-center text-muted-foreground">{loading ? copy.loading : copy.empty}</td></tr> : documents.map((document) => <tr key={document.id}><td className="p-3">{document.title}</td><td className="p-3">{copy.categories[document.category as keyof typeof copy.categories] ?? document.category}</td><td className="p-3">{document.current_version?.version_number ?? "—"}</td><td className="p-3"><div className="flex flex-wrap gap-2"><button type="button" disabled={busy} onClick={() => void download(document)} className="inline-flex items-center gap-1 rounded border px-3 py-1.5 hover:bg-muted"><Download className="h-4 w-4" />{copy.download}</button><button type="button" disabled={busy} onClick={() => openUpload(document)} className="rounded border px-3 py-1.5 hover:bg-muted">{copy.version}</button></div></td></tr>)}</tbody></table></div>
      <div className="flex justify-end gap-3"><button type="button" disabled={loading || offset === 0} onClick={() => setOffset((value) => Math.max(0, value - 50))} className="rounded-lg border px-3 py-2 text-sm disabled:opacity-50">{copy.previous}</button><button type="button" disabled={loading || offset + 50 >= total} onClick={() => setOffset((value) => value + 50)} className="rounded-lg border px-3 py-2 text-sm disabled:opacity-50">{copy.next}</button></div>
    </div>
  );
}
