"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { AlertCircle, CheckCircle2, Database, Play, RefreshCw, Search, Settings2, X } from "lucide-react";
import { useLocale } from "@/components/layout/shell";
import { apiFetch } from "@/lib/api-client";
import { useAuth } from "@/lib/auth-context";

export interface ResourceDefinition {
  key: string;
  labelAr: string;
  labelEn: string;
  path: string;
  tenant?: boolean;
}

interface ModuleAPIViewProps {
  titleAr: string;
  titleEn: string;
  descriptionAr: string;
  descriptionEn: string;
  resources: ResourceDefinition[];
}

type RecordRow = Record<string, unknown>;
type MutationMethod = "POST" | "PUT" | "PATCH" | "DELETE";

export interface MutationRequest {
  path: string;
  init: RequestInit;
}

const preferredColumns = [
  "id",
  "code",
  "number",
  "name_ar",
  "name_en",
  "name",
  "title",
  "status",
  "type",
  "amount",
  "total_amount",
  "created_at",
  "updated_at",
];

export function appendTenant(path: string, legalEntityId: string | null, tenant = true): string {
  if (!tenant || !legalEntityId) return path;
  const separator = path.includes("?") ? "&" : "?";
  return `${path}${separator}legal_entity_id=${encodeURIComponent(legalEntityId)}`;
}

export function normalizeRows(payload: unknown): RecordRow[] {
  if (Array.isArray(payload)) return payload.map(toRecord);
  if (!payload || typeof payload !== "object") return [{ value: payload }];

  const envelope = payload as RecordRow;
  for (const key of ["data", "items", "results", "records"]) {
    if (Array.isArray(envelope[key])) return (envelope[key] as unknown[]).map(toRecord);
  }
  return [envelope];
}

export function buildMutationRequest(
  rawPath: string,
  method: MutationMethod,
  rawBody: string,
  legalEntityId: string | null,
  tenant = true,
): MutationRequest {
  const path = rawPath.trim();
  if (!path.startsWith("/api/v1/") || path.startsWith("//") || /[\r\n]/.test(path)) {
    throw new Error("The operation endpoint must be a local /api/v1/ path.");
  }

  const init: RequestInit = { method };
  if (method !== "DELETE") {
    let payload: unknown;
    try {
      payload = JSON.parse(rawBody || "{}");
    } catch {
      throw new Error("The request body must be valid JSON.");
    }
    if (!payload || typeof payload !== "object" || Array.isArray(payload)) {
      throw new Error("The request body must be a JSON object.");
    }
    if (tenant && legalEntityId && !("legal_entity_id" in payload)) {
      payload = { ...payload, legal_entity_id: legalEntityId };
    }
    init.body = JSON.stringify(payload);
  }

  return {
    path: appendTenant(path, legalEntityId, tenant),
    init,
  };
}

function toRecord(value: unknown): RecordRow {
  return value && typeof value === "object" && !Array.isArray(value)
    ? (value as RecordRow)
    : { value };
}

function displayValue(value: unknown): string {
  if (value === null || value === undefined || value === "") return "—";
  if (typeof value === "boolean") return value ? "✓" : "✕";
  if (typeof value === "object") {
    const serialized = JSON.stringify(value);
    return serialized.length > 160 ? `${serialized.slice(0, 157)}…` : serialized;
  }
  return String(value);
}

export function ModuleAPIView({
  titleAr,
  titleEn,
  descriptionAr,
  descriptionEn,
  resources,
}: ModuleAPIViewProps) {
  const { locale, dir } = useLocale();
  const { legalEntityId } = useAuth();
  const isAr = locale === "ar";
  const [activeKey, setActiveKey] = useState(resources[0]?.key ?? "");
  const [rows, setRows] = useState<RecordRow[]>([]);
  const [search, setSearch] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);
  const [operationOpen, setOperationOpen] = useState(false);
  const [operationMethod, setOperationMethod] = useState<MutationMethod>("POST");
  const [operationPath, setOperationPath] = useState(resources[0]?.path ?? "/api/v1/");
  const [operationBody, setOperationBody] = useState("{\n  \n}");
  const [operationLoading, setOperationLoading] = useState(false);
  const [operationError, setOperationError] = useState<string | null>(null);
  const [operationResult, setOperationResult] = useState<unknown>(null);
  const [selectedRow, setSelectedRow] = useState<RecordRow | null>(null);

  const activeResource = resources.find((resource) => resource.key === activeKey) ?? resources[0];

  const load = useCallback(async () => {
    if (!activeResource) return;
    if (activeResource.tenant !== false && !legalEntityId) {
      setRows([]);
      setError(isAr ? "اختر كياناً قانونياً لعرض البيانات." : "Select a legal entity to load data.");
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const payload = await apiFetch<unknown>(appendTenant(activeResource.path, legalEntityId, activeResource.tenant !== false));
      setRows(normalizeRows(payload));
      setLastUpdated(new Date());
    } catch (requestError) {
      setRows([]);
      setError(requestError instanceof Error ? requestError.message : isAr ? "تعذر تحميل البيانات." : "Unable to load data.");
    } finally {
      setLoading(false);
    }
  }, [activeResource, isAr, legalEntityId]);

  useEffect(() => {
    void load();
  }, [load]);

  useEffect(() => {
    if (!activeResource) return;
    setOperationPath(activeResource.path);
    setOperationError(null);
    setOperationResult(null);
    setSelectedRow(null);
  }, [activeResource]);

  const runOperation = async () => {
    if (!activeResource) return;
    setOperationLoading(true);
    setOperationError(null);
    setOperationResult(null);
    try {
      const request = buildMutationRequest(
        operationPath,
        operationMethod,
        operationBody,
        legalEntityId,
        activeResource.tenant !== false,
      );
      const result = await apiFetch<unknown>(request.path, request.init);
      setOperationResult(result ?? { ok: true });
      await load();
    } catch (requestError) {
      setOperationError(requestError instanceof Error ? requestError.message : isAr ? "تعذر تنفيذ العملية." : "Operation failed.");
    } finally {
      setOperationLoading(false);
    }
  };

  const useSelectedRecord = () => {
    const id = selectedRow?.id;
    if (!id || !activeResource) return;
    setOperationPath(`${activeResource.path.replace(/\/$/, "")}/${encodeURIComponent(String(id))}`);
  };

  const filteredRows = useMemo(() => {
    const query = search.trim().toLocaleLowerCase();
    if (!query) return rows;
    return rows.filter((row) => JSON.stringify(row).toLocaleLowerCase().includes(query));
  }, [rows, search]);

  const columns = useMemo(() => {
    const keys = Array.from(new Set(filteredRows.flatMap((row) => Object.keys(row))));
    return keys
      .sort((left, right) => {
        const leftIndex = preferredColumns.indexOf(left);
        const rightIndex = preferredColumns.indexOf(right);
        if (leftIndex === -1 && rightIndex === -1) return left.localeCompare(right);
        if (leftIndex === -1) return 1;
        if (rightIndex === -1) return -1;
        return leftIndex - rightIndex;
      })
      .slice(0, 12);
  }, [filteredRows]);

  return (
    <div className="space-y-6 pb-12" dir={dir}>
      <div className="rounded-2xl border bg-card p-6 shadow-sm">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div>
            <div className="mb-2 inline-flex items-center gap-2 rounded-full border bg-primary/10 px-3 py-1 text-xs font-semibold text-primary">
              <Database className="h-3.5 w-3.5" />
              {isAr ? "بيانات مباشرة من النظام" : "Live system data"}
            </div>
            <h1 className="text-2xl font-bold">{isAr ? titleAr : titleEn}</h1>
            <p className="mt-2 max-w-3xl text-sm text-muted-foreground">{isAr ? descriptionAr : descriptionEn}</p>
          </div>
          <div className="flex flex-wrap gap-2">
            <button
              type="button"
              onClick={() => void load()}
              disabled={loading}
              className="inline-flex items-center justify-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground disabled:opacity-60"
            >
              <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />
              {isAr ? "تحديث" : "Refresh"}
            </button>
            <button
              type="button"
              onClick={() => setOperationOpen((open) => !open)}
              className="inline-flex items-center justify-center gap-2 rounded-lg border bg-background px-4 py-2 text-sm font-semibold hover:bg-muted"
              aria-expanded={operationOpen}
              aria-controls="module-operation-panel"
            >
              {operationOpen ? <X className="h-4 w-4" /> : <Settings2 className="h-4 w-4" />}
              {operationOpen ? (isAr ? "إغلاق العملية" : "Close operation") : (isAr ? "عملية جديدة" : "New operation")}
            </button>
          </div>
        </div>
      </div>

      <div className="flex gap-2 overflow-x-auto border-b pb-2">
        {resources.map((resource) => (
          <button
            key={resource.key}
            type="button"
            onClick={() => {
              setActiveKey(resource.key);
              setSearch("");
            }}
            className={`whitespace-nowrap rounded-lg px-3 py-2 text-xs font-semibold ${
              activeResource?.key === resource.key ? "bg-primary text-primary-foreground" : "bg-muted text-muted-foreground hover:text-foreground"
            }`}
          >
            {isAr ? resource.labelAr : resource.labelEn}
          </button>
        ))}
      </div>

      {operationOpen && activeResource && (
        <section id="module-operation-panel" className="rounded-xl border bg-card p-5 shadow-sm" aria-label={isAr ? "تنفيذ عملية API" : "Run API operation"}>
          <div className="mb-4">
            <h2 className="font-bold">{isAr ? "تنفيذ عملية فعلية" : "Run a live operation"}</h2>
            <p className="mt-1 text-xs text-muted-foreground">
              {isAr
                ? "تُرسل العملية مباشرة إلى الخادم وتخضع لصلاحيات حسابك وعزل الكيان. استخدم مسار العملية الموثق في عقد OpenAPI."
                : "The operation is sent directly to the server and remains subject to your permissions and tenant boundary. Use an endpoint documented by the OpenAPI contract."}
            </p>
          </div>

          <div className="grid gap-4 lg:grid-cols-[9rem_1fr]">
            <label className="space-y-1.5 text-xs font-semibold">
              <span>{isAr ? "الطريقة" : "Method"}</span>
              <select
                id="operation-method"
                value={operationMethod}
                onChange={(event) => setOperationMethod(event.target.value as MutationMethod)}
                className="h-10 w-full rounded-lg border bg-background px-3 font-mono"
              >
                <option>POST</option>
                <option>PUT</option>
                <option>PATCH</option>
                <option>DELETE</option>
              </select>
            </label>
            <label className="space-y-1.5 text-xs font-semibold">
              <span>{isAr ? "مسار العملية" : "Operation endpoint"}</span>
              <input
                id="operation-path"
                value={operationPath}
                onChange={(event) => setOperationPath(event.target.value)}
                dir="ltr"
                className="h-10 w-full rounded-lg border bg-background px-3 font-mono text-xs"
              />
            </label>
          </div>

          {selectedRow?.id !== undefined && (
            <button type="button" onClick={useSelectedRecord} className="mt-3 text-xs font-semibold text-primary hover:underline">
              {isAr ? `استخدام السجل المحدد: ${String(selectedRow.id)}` : `Use selected record: ${String(selectedRow.id)}`}
            </button>
          )}

          {operationMethod !== "DELETE" && (
            <label className="mt-4 block space-y-1.5 text-xs font-semibold">
              <span>{isAr ? "بيانات الطلب (JSON)" : "Request body (JSON)"}</span>
              <textarea
                id="operation-body"
                value={operationBody}
                onChange={(event) => setOperationBody(event.target.value)}
                dir="ltr"
                spellCheck={false}
                rows={8}
                className="w-full rounded-lg border bg-background p-3 font-mono text-xs outline-none focus:ring-1 focus:ring-ring"
              />
              <span className="block font-normal text-muted-foreground">
                {isAr ? "يُضاف legal_entity_id المختار تلقائيًا إذا لم ترسله." : "The selected legal_entity_id is added automatically when omitted."}
              </span>
            </label>
          )}

          {operationError && (
            <div role="alert" className="mt-4 flex items-start gap-2 rounded-lg border border-destructive/40 bg-destructive/10 p-3 text-xs text-destructive">
              <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
              <span>{operationError}</span>
            </div>
          )}
          {operationResult !== null && (
            <div data-testid="operation-success" className="mt-4 rounded-lg border border-emerald-500/40 bg-emerald-500/10 p-3 text-xs text-emerald-700 dark:text-emerald-300">
              <div className="mb-2 flex items-center gap-2 font-semibold">
                <CheckCircle2 className="h-4 w-4" />
                {isAr ? "نُفذت العملية وحُدثت النتائج" : "Operation completed and results refreshed"}
              </div>
              <pre className="max-h-48 overflow-auto whitespace-pre-wrap break-all font-mono">{JSON.stringify(operationResult, null, 2)}</pre>
            </div>
          )}

          <button
            id="operation-submit"
            type="button"
            onClick={() => void runOperation()}
            disabled={operationLoading || (activeResource.tenant !== false && !legalEntityId)}
            className="mt-4 inline-flex items-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground disabled:opacity-60"
          >
            <Play className="h-4 w-4" />
            {operationLoading ? (isAr ? "جاري التنفيذ…" : "Running…") : (isAr ? "تنفيذ" : "Run")}
          </button>
        </section>
      )}

      {error && (
        <div role="alert" className="flex items-start gap-3 rounded-xl border border-destructive/40 bg-destructive/10 p-4 text-sm text-destructive">
          <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
          <span>{error}</span>
        </div>
      )}

      <div className="rounded-xl border bg-card shadow-sm">
        <div className="flex flex-col gap-3 border-b p-4 sm:flex-row sm:items-center sm:justify-between">
          <div className="relative w-full max-w-md">
            <Search className="absolute start-3 top-2.5 h-4 w-4 text-muted-foreground" />
            <input
              value={search}
              onChange={(event) => setSearch(event.target.value)}
              placeholder={isAr ? "بحث في النتائج الفعلية…" : "Search live results…"}
              className="h-9 w-full rounded-lg border bg-background ps-9 pe-3 text-sm outline-none focus:ring-1 focus:ring-ring"
            />
          </div>
          <div className="text-xs text-muted-foreground">
            {filteredRows.length} {isAr ? "سجل" : "records"}
            {lastUpdated ? ` · ${lastUpdated.toLocaleTimeString(isAr ? "ar-SA" : "en-US")}` : ""}
          </div>
        </div>

        <div className="overflow-x-auto">
          {loading ? (
            <div className="p-10 text-center text-sm text-muted-foreground">{isAr ? "جاري تحميل البيانات…" : "Loading data…"}</div>
          ) : filteredRows.length === 0 ? (
            <div className="p-10 text-center text-sm text-muted-foreground">
              {isAr ? "لا توجد سجلات فعلية مطابقة." : "No matching live records."}
            </div>
          ) : (
            <table className="w-full min-w-[720px] text-xs">
              <thead className="border-b bg-muted/50 text-muted-foreground">
                <tr>
                  {columns.map((column) => <th key={column} className="px-4 py-3 text-start font-semibold">{column}</th>)}
                </tr>
              </thead>
              <tbody className="divide-y">
                {filteredRows.map((row, rowIndex) => (
                  <tr
                    key={String(row.id ?? `${activeKey}-${rowIndex}`)}
                    onClick={() => setSelectedRow(row)}
                    className={`cursor-pointer hover:bg-muted/30 ${selectedRow === row ? "bg-primary/10" : ""}`}
                  >
                    {columns.map((column) => (
                      <td key={column} title={displayValue(row[column])} className="max-w-72 truncate px-4 py-3 align-top">
                        {displayValue(row[column])}
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>

      <p className="text-xs text-muted-foreground">
        {isAr
          ? "كل النتائج والعمليات في هذه الشاشة متصلة بالخادم مباشرة؛ لا تُنشأ بيانات محلية وهمية."
          : "All results and operations on this screen use the live server; no local records are fabricated."}
      </p>
    </div>
  );
}
