"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { AlertCircle, Database, Plus, RefreshCw, Search, X } from "lucide-react";
import { useLocale } from "@/components/layout/shell";
import { apiFetch } from "@/lib/api-client";
import { useAuth } from "@/lib/auth-context";
import { ResourceCreateForm } from "@/components/forms/resource-create-form";
import { fieldKey, fieldLabel, stateLabel } from "@/lib/record-presentation";
import { ResourceActions } from "@/components/modules/resource-actions";
import { transactionForms } from "@/lib/transaction-forms";
import { resourceForms } from "@/lib/resource-forms";
import { hasServerPagination, hasServerSearch, normalizeRows, responseTotal, loadReferenceRows } from "@/lib/resource-data";
import { RecordDetails } from "@/components/modules/record-details";
export { normalizeRows } from "@/lib/resource-data";

export interface ResourceDefinition {
  key: string;
  labelAr: string;
  labelEn: string;
  path: string;
  tenant?: boolean;
  filters?: ResourceFilterDefinition[];
}

export interface ResourceFilterDefinition {
  key: string;
  labelAr: string;
  labelEn: string;
  optionsPath: string;
  tenant?: boolean;
  valueKey?: string;
  optionLabelKeysAr?: string[];
  optionLabelKeysEn?: string[];
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
const noResourceFilters: ResourceFilterDefinition[] = [];

export interface MutationRequest {
  path: string;
  init: RequestInit;
}

const preferredColumns = [
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
  return appendQueryParams(path, { legal_entity_id: legalEntityId });
}

export function appendQueryParams(path: string, params: Record<string, string>): string {
  const [pathname, rawQuery = ""] = path.split("?", 2);
  const query = new URLSearchParams(rawQuery);
  for (const [key, value] of Object.entries(params)) {
    if (value) query.set(key, value);
  }
  const serialized = query.toString();
  return serialized ? `${pathname}?${serialized}` : pathname;
}

export function buildResourcePath(
  resource: ResourceDefinition,
  legalEntityId: string | null,
  filterValues: Record<string, string> = {},
): string {
  return appendQueryParams(
    appendTenant(resource.path, legalEntityId, resource.tenant !== false),
    Object.fromEntries((resource.filters ?? []).map((filter) => [filter.key, filterValues[filter.key] ?? ""])),
  );
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
    if (tenant && legalEntityId) {
      if ("legal_entity_id" in payload && payload.legal_entity_id !== legalEntityId) {
        throw new Error("The request legal entity must match the selected legal entity.");
      }
      payload = { ...payload, legal_entity_id: legalEntityId };
    }
    init.body = JSON.stringify(payload);
  }

  return {
    path: appendTenant(path, legalEntityId, tenant),
    init,
  };
}

function displayValue(value: unknown): string {
  if (value === null || value === undefined || value === "") return "—";
  if (typeof value === "boolean") return value ? "✓" : "✕";
  if (typeof value === "object") {
    return Array.isArray(value) ? String(value.length) : "…";
  }
  return String(value);
}

function optionValue(row: RecordRow, filter: ResourceFilterDefinition): string {
  const value = row[fieldKey(filter.valueKey ?? "id")];
  return value === null || value === undefined ? "" : String(value);
}

function optionLabel(row: RecordRow, filter: ResourceFilterDefinition, isAr: boolean): string {
  const keys = isAr ? filter.optionLabelKeysAr : filter.optionLabelKeysEn;
  const parts = (keys ?? [])
    .map((key) => row[fieldKey(key)])
    .filter((value) => value !== null && value !== undefined && value !== "")
    .map(String);
  return parts.length > 0 ? parts.join(" · ") : optionValue(row, filter);
}

export function ModuleAPIView({
  titleAr,
  titleEn,
  descriptionAr,
  descriptionEn,
  resources,
}: ModuleAPIViewProps) {
  const { locale, dir, t } = useLocale();
  const { legalEntityId, branchId, refresh } = useAuth();
  const isAr = locale === "ar";
  const [activeKey, setActiveKey] = useState(resources[0]?.key ?? "");
  const [total, setTotal] = useState<number | null>(null);
  const [paging, setPaging] = useState({ scope: "", index: 0, size: 25 });
  const [rows, setRows] = useState<RecordRow[]>([]);
  const [filterOptions, setFilterOptions] = useState<Record<string, RecordRow[]>>({});
  const [filterValues, setFilterValues] = useState<Record<string, string>>({});
  const [filtersLoading, setFiltersLoading] = useState(false);
  const [loadedFilterScope, setLoadedFilterScope] = useState<string | null>(null);
  const [filterError, setFilterError] = useState<string | null>(null);
  const [search, setSearch] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);
  const [selectedRow, setSelectedRow] = useState<RecordRow | null>(null);
  const [createOpen, setCreateOpen] = useState(false);
  const [createSuccess, setCreateSuccess] = useState(false);
  const loadController = useRef<AbortController | null>(null);

  const activeResource = resources.find((resource) => resource.key === activeKey) ?? resources[0];
  const activeFilters = activeResource?.filters ?? noResourceFilters;
  const createDefinition = activeResource ? (resourceForms[activeResource.path] ?? transactionForms[activeResource.path]) : undefined;
  const filterScope = `${activeResource?.path}:${legalEntityId}:${branchId}`;
  const serverPagination = hasServerPagination(activeResource?.path ?? "");
  const serverSearch = hasServerSearch(activeResource?.path ?? "");
  const pageScope = `${filterScope}:${JSON.stringify(filterValues)}:${search}`;
  const pageIndex = paging.scope === pageScope ? paging.index : 0;
  const pageSize = paging.size;
  const filtersReady = activeFilters.length === 0 || (loadedFilterScope === filterScope && activeFilters.every((filter) => Boolean(filterValues[filter.key])));

  useEffect(() => {
    let cancelled = false;

    const loadFilterOptions = async () => {
      setLoadedFilterScope(null);
      if (!activeResource || activeFilters.length === 0) {
        setFilterOptions({});
        setFilterValues({});
        setFilterError(null);
        setFiltersLoading(false);
        return;
      }
      if (activeResource.tenant !== false && !legalEntityId) {
        setFilterOptions({});
        setFilterValues({});
        setFilterError(null);
        setFiltersLoading(false);
        return;
      }

      setFiltersLoading(true);
      setFilterError(null);
      try {
        const loaded = await Promise.all(
          activeFilters.map(async (filter) => {
            const options = await loadReferenceRows(filter.optionsPath, legalEntityId, new AbortController().signal, filter.tenant !== false);
            return [filter.key, options] as const;
          }),
        );
        if (cancelled) return;

        const nextOptions = Object.fromEntries(loaded);
        setLoadedFilterScope(filterScope);
        setFilterOptions(nextOptions);
        setFilterValues((current) =>
          Object.fromEntries(
            activeFilters.map((filter) => {
              const options = nextOptions[filter.key] ?? [];
              const currentValue = current[filter.key] ?? "";
              const currentStillExists = options.some((row) => optionValue(row, filter) === currentValue);
              return [filter.key, currentStillExists ? currentValue : optionValue(options[0] ?? {}, filter)];
            }),
          ),
        );
      } catch (requestError) {
        if (cancelled) return;
        setFilterOptions({});
        setFilterValues({});
        setFilterError(
          requestError instanceof Error
            ? requestError.message
            : isAr
              ? "تعذر تحميل خيارات التصفية."
              : "Unable to load filter options.",
        );
      } finally {
        if (!cancelled) setFiltersLoading(false);
      }
    };

    void loadFilterOptions();
    return () => {
      cancelled = true;
    };
  }, [activeResource, activeFilters, filterScope, isAr, legalEntityId]);

  const load = useCallback(async () => {
    loadController.current?.abort();
    const controller = new AbortController();
    loadController.current = controller;
    setRows([]);
    setSelectedRow(null);
    setLastUpdated(null);
    setLoading(false);
    if (!activeResource) return;
    if (activeResource.tenant !== false && !legalEntityId) {
      setRows([]);
      setError(isAr ? "اختر كياناً قانونياً لعرض البيانات." : "Select a legal entity to load data.");
      return;
    }
    if (!filtersReady) {
      setRows([]);
      setLastUpdated(null);
      setError(null);
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const requestPath = appendQueryParams(buildResourcePath(activeResource, legalEntityId, filterValues), {
        ...(serverPagination ? { limit: String(pageSize), offset: String(pageIndex * pageSize) } : {}),
        ...(serverSearch && search.trim() ? { search: search.trim() } : {}),
      });
      const payload = await apiFetch<unknown>(requestPath, { signal: controller.signal });
      if (controller.signal.aborted) return;
      setRows(normalizeRows(payload));
      setTotal(responseTotal(payload));
      setLastUpdated(new Date());
    } catch (requestError) {
      if (controller.signal.aborted) return;
      setRows([]);
      setError(requestError instanceof Error ? requestError.message : isAr ? "تعذر تحميل البيانات." : "Unable to load data.");
    } finally {
      if (!controller.signal.aborted) setLoading(false);
    }
  }, [activeResource, filterValues, filtersReady, isAr, legalEntityId, serverPagination, serverSearch, pageSize, pageIndex, search]);

  useEffect(() => {
    void load();
    return () => loadController.current?.abort();
  }, [load]);

  useEffect(() => {
    if (!activeResource) return;
    setSelectedRow(null);
    setCreateOpen(false);
    setCreateSuccess(false);
  }, [activeResource, legalEntityId, branchId]);

  const filteredRows = useMemo(() => {
    const query = search.trim().toLocaleLowerCase();
    if (!query || serverSearch) return rows;
    return rows.filter((row) => JSON.stringify(row).toLocaleLowerCase().includes(query));
  }, [rows, search, serverSearch]);
  const visibleRows = serverPagination ? filteredRows : filteredRows.slice(pageIndex * pageSize, (pageIndex + 1) * pageSize);
  const recordCount = serverPagination ? total : filteredRows.length;
  const hasNext = serverPagination ? total === null ? rows.length === pageSize : (pageIndex + 1) * pageSize < total : (pageIndex + 1) * pageSize < filteredRows.length;

  const columns = useMemo(() => {
    const keys = Array.from(new Set(filteredRows.flatMap((row) => Object.keys(row)))).filter(key => !["legal_entity_id", "created_by", "created_by_user_id", "updated_by", "journal_entry_id", "version"].includes(key));
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
            {createDefinition && <button type="button" onClick={() => { setCreateOpen(true); setCreateSuccess(false); }} disabled={activeResource.tenant !== false && !legalEntityId} className="inline-flex items-center justify-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground disabled:opacity-60"><Plus className="h-4 w-4" />{t.resourceForm.create}</button>}
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

      {createSuccess && <p role="status" className="rounded-lg border border-emerald-500/40 bg-emerald-500/10 p-3 text-sm text-emerald-700 dark:text-emerald-300">{t.resourceForm.success}</p>}
      {createOpen && createDefinition && activeResource && (
        <ResourceCreateForm key={`${activeResource.key}-${legalEntityId}-${branchId}`} definition={createDefinition} path={activeResource.path} title={isAr ? activeResource.labelAr : activeResource.labelEn} legalEntityId={legalEntityId} tenant={activeResource.tenant !== false} initialValues={{ ...filterValues, ...(branchId ? { branch_id: branchId } : {}) }}
          onCancel={() => setCreateOpen(false)}
          onSaved={async () => {
            setCreateOpen(false);
            setCreateSuccess(true);
            await load();
            if (activeResource.path === "/api/v1/org/legal-entities") await refresh();
          }}
        />
      )}

      {activeFilters.length > 0 && (
        <section className="rounded-xl border bg-card p-4 shadow-sm" aria-label={isAr ? "مرشحات البيانات" : "Data filters"}>
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {activeFilters.map((filter) => {
              const options = filterOptions[filter.key] ?? [];
              return (
                <label key={filter.key} className="space-y-1.5 text-xs font-semibold">
                  <span>{isAr ? filter.labelAr : filter.labelEn}</span>
                  <select
                    aria-label={isAr ? filter.labelAr : filter.labelEn}
                    value={filterValues[filter.key] ?? ""}
                    onChange={(event) => {
                      setFilterValues((current) => ({ ...current, [filter.key]: event.target.value }));
                      setSelectedRow(null);
                    }}
                    disabled={filtersLoading || options.length === 0}
                    className="h-10 w-full rounded-lg border bg-background px-3 disabled:opacity-60"
                  >
                    {options.length === 0 && (
                      <option value="">
                        {filtersLoading
                          ? isAr ? "جاري تحميل الخيارات…" : "Loading options…"
                          : isAr ? "لا توجد خيارات متاحة" : "No options available"}
                      </option>
                    )}
                    {options.map((row) => {
                      const value = optionValue(row, filter);
                      return <option key={value} value={value}>{optionLabel(row, filter, isAr)}</option>;
                    })}
                  </select>
                </label>
              );
            })}
          </div>
        </section>
      )}

      {(filterError || error) && (
        <div role="alert" className="flex items-start gap-3 rounded-xl border border-destructive/40 bg-destructive/10 p-4 text-sm text-destructive">
          <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
          <span>{filterError || error}</span>
        </div>
      )}

      <div className="rounded-xl border bg-card shadow-sm">
        <div className="flex flex-col gap-3 border-b p-4 sm:flex-row sm:items-center sm:justify-between">
          <div className="relative w-full max-w-md">
            <Search className="absolute start-3 top-2.5 h-4 w-4 text-muted-foreground" />
            <input
              value={search}
              onChange={(event) => setSearch(event.target.value)}
              aria-label={serverPagination && !serverSearch ? t.business.searchPage : t.business.searchAll}
              placeholder={serverPagination && !serverSearch ? t.business.searchPage : t.business.searchAll}
              className="h-9 w-full rounded-lg border bg-background ps-9 pe-3 text-sm outline-none focus:ring-1 focus:ring-ring"
            />
          </div>
          <div className="text-xs text-muted-foreground">
            {recordCount ?? filteredRows.length} {t.business.records}
            {lastUpdated ? ` · ${lastUpdated.toLocaleTimeString(isAr ? "ar-SA" : "en-US")}` : ""}
          </div>
        </div>

        <div className="overflow-x-auto">
          {loading || filtersLoading ? (
            <div className="p-10 text-center text-sm text-muted-foreground">{isAr ? "جاري تحميل البيانات…" : "Loading data…"}</div>
          ) : filteredRows.length === 0 ? (
            <div className="p-10 text-center text-sm text-muted-foreground">
              {isAr ? "لا توجد سجلات فعلية مطابقة." : "No matching live records."}
            </div>
          ) : (
            <table className="w-full min-w-[720px] text-xs">
              <thead className="border-b bg-muted/50 text-muted-foreground">
                <tr>
                  {columns.map((column) => <th key={column} className="px-4 py-3 text-start font-semibold">{fieldLabel(column,locale)}</th>)}
                </tr>
              </thead>
              <tbody className="divide-y">
                {visibleRows.map((row, rowIndex) => (
                  <tr
                    key={String(row.id ?? `${activeKey}-${rowIndex}`)}
                    onClick={() => setSelectedRow(row)}
                    onKeyDown={(event) => { if (event.key === "Enter" || event.key === " ") { event.preventDefault(); setSelectedRow(row); } }}
                    tabIndex={0}
                    aria-selected={selectedRow === row}
                    className={`cursor-pointer hover:bg-muted/30 ${selectedRow === row ? "bg-primary/10" : ""}`}
                  >
                    {columns.map((column) => (
                      <td key={column} title={column === "status" || column.endsWith("_status") ? stateLabel(row[column],locale) : displayValue(row[column])} className="max-w-72 truncate px-4 py-3 align-top">
                        {column === "status" || column.endsWith("_status") ? stateLabel(row[column],locale) : displayValue(row[column])}
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>

      <nav aria-label={t.business.page} className="flex flex-wrap items-center justify-end gap-3 text-sm">
        <label className="flex items-center gap-2">{t.business.pageSize}<select value={pageSize} onChange={event => { setPaging({ scope: pageScope, index: 0, size: Number(event.target.value) }); setSelectedRow(null); }} className="rounded-lg border bg-background px-2 py-2">{[25, 50, 100].map(size => <option key={size} value={size}>{size}</option>)}</select></label>
        <span>{t.business.page} {pageIndex + 1}{recordCount !== null ? ` / ${Math.max(1, Math.ceil(recordCount / pageSize))}` : ""}</span>
        <button type="button" disabled={loading || pageIndex === 0} onClick={() => { setPaging({ scope: pageScope, index: pageIndex - 1, size: pageSize }); setSelectedRow(null); }} className="rounded-lg border px-3 py-2 disabled:opacity-50">{t.business.previous}</button>
        <button type="button" disabled={loading || !hasNext} onClick={() => { setPaging({ scope: pageScope, index: pageIndex + 1, size: pageSize }); setSelectedRow(null); }} className="rounded-lg border px-3 py-2 disabled:opacity-50">{t.business.next}</button>
      </nav>

      {selectedRow && (
        <section aria-label={t.resourceForm.details} className="rounded-xl border bg-card p-5">
          <div className="mb-4 flex items-center justify-between gap-4"><h2 className="font-semibold">{t.resourceForm.details}</h2><button type="button" onClick={() => setSelectedRow(null)} aria-label={t.resourceForm.close} className="rounded-lg p-2 hover:bg-muted"><X className="h-4 w-4" /></button></div>
          <ResourceActions key={`actions-${activeResource.path}-${selectedRow.id}`} path={activeResource.path} row={selectedRow} legalEntityId={legalEntityId} onSaved={load} />
          <RecordDetails key={`details-${activeResource.path}-${selectedRow.id}`} row={selectedRow} path={activeResource.path} legalEntityId={legalEntityId} />
        </section>
      )}


    </div>
  );
}
