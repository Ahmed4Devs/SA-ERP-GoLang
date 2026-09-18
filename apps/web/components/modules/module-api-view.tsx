"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { AlertCircle, Database, RefreshCw, Search } from "lucide-react";
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
          <button
            type="button"
            onClick={() => void load()}
            disabled={loading}
            className="inline-flex items-center justify-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground disabled:opacity-60"
          >
            <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />
            {isAr ? "تحديث" : "Refresh"}
          </button>
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
                  <tr key={String(row.id ?? `${activeKey}-${rowIndex}`)} className="hover:bg-muted/30">
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
          ? "هذه الشاشة لا تنشئ نتائج محلية وهمية. أي عملية كتابة غير متاحة هنا يجب تنفيذها من مسار عمل موصول بالـAPI ومخوّل بالصلاحيات المناسبة."
          : "This screen never fabricates local results. Writes must use an API-backed workflow with the required permission."}
      </p>
    </div>
  );
}
