"use client";

import { useEffect, useMemo, useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useLocale } from "@/components/layout/shell";
import { apiFetch } from "@/lib/api-client";
import { createResourceSchema, resourceFormPayload, type ResourceFormDefinition } from "@/lib/resource-forms";

type OptionRow = Record<string, unknown>;
interface ResourceCreateFormProps {
  definition: ResourceFormDefinition;
  path: string;
  title: string;
  legalEntityId: string | null;
  tenant: boolean;
  initialValues?: Record<string, string>;
  onCancel: () => void;
  onSaved: () => Promise<void>;
}

export function ResourceCreateForm({ definition, path, title, legalEntityId, tenant, initialValues, onCancel, onSaved }: ResourceCreateFormProps) {
  const { locale, t } = useLocale();
  const copy = t.resourceForm;
  const [options, setOptions] = useState<Record<string, OptionRow[]>>({});
  const [loadingOptions, setLoadingOptions] = useState(true);
  const [optionError, setOptionError] = useState(false);
  const [retry, setRetry] = useState(0);
  const [saveError, setSaveError] = useState<string | null>(null);
  const schema = useMemo(() => createResourceSchema(definition, copy), [definition, copy]);
  const { register, handleSubmit, formState: { errors, isSubmitting } } = useForm<Record<string, string>>({
    resolver: zodResolver(schema),
    defaultValues: Object.fromEntries(definition.fields.map((field) => [field.key, initialValues?.[field.key] ?? field.initial ?? ""])),
  });

  useEffect(() => {
    const controller = new AbortController();
    setLoadingOptions(true);
    setOptionError(false);
    setOptions({});
    async function loadOptions() {
      try {
        const loaded = await Promise.all(definition.fields.filter((field) => field.source).map(async (field) => {
          const query = field.sourceTenant !== false && legalEntityId ? `?legal_entity_id=${encodeURIComponent(legalEntityId)}` : "";
          const result = await apiFetch<{ data?: OptionRow[] } | OptionRow[]>(`${field.source}${query}`, { signal: controller.signal });
          const rows = Array.isArray(result) ? result : Array.isArray(result.data) ? result.data : [];
          return [field.key, rows.filter((row) => row.id && row.is_active !== false)] as const;
        }));
        if (!controller.signal.aborted) setOptions(Object.fromEntries(loaded));
      } catch {
        if (!controller.signal.aborted) setOptionError(true);
      } finally {
        if (!controller.signal.aborted) setLoadingOptions(false);
      }
    }
    void loadOptions();
    return () => controller.abort();
  }, [definition, legalEntityId, retry]);

  const submit = handleSubmit(async (values) => {
    setSaveError(null);
    try {
      const payload = resourceFormPayload(definition, values);
      if (tenant) {
        if (!legalEntityId) return;
        payload.legal_entity_id = legalEntityId;
      }
      const query = tenant && legalEntityId ? `?legal_entity_id=${encodeURIComponent(legalEntityId)}` : "";
      await apiFetch(path + query, { method: "POST", body: JSON.stringify(payload) });
      await onSaved();
    } catch (error) {
      setSaveError(error instanceof Error ? error.message : copy.saveError);
    }
  });
  const missingRequiredOptions = definition.fields.some((field) => field.required && field.source && !options[field.key]?.length);
  const controlClass = "min-h-10 w-full rounded-lg border bg-background px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-ring disabled:opacity-60";

  return (
    <section className="rounded-xl border bg-card p-5 shadow-sm" aria-labelledby="resource-form-title">
      <h2 id="resource-form-title" className="mb-5 text-lg font-semibold">{copy.title}: {title}</h2>
      <form onSubmit={submit} noValidate className="space-y-5">
        <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {definition.fields.map((field) => {
            const id = `create-${field.key}`;
            const error = errors[field.key]?.message;
            const props = {
              id,
              ...register(field.key),
              "aria-invalid": Boolean(error),
              "aria-describedby": error ? `${id}-error` : field.source && !options[field.key]?.length ? `${id}-help` : undefined,
              "aria-required": Boolean(field.required),
              disabled: isSubmitting,
              className: controlClass,
            };
            return (
              <div key={field.key} className="min-w-0 space-y-1.5">
                <label htmlFor={id} className="text-sm font-medium">{copy.fields[field.label]}{field.required ? " *" : ` (${copy.optional})`}</label>
                {field.type === "select" ? (
                  <select {...props} disabled={isSubmitting || (Boolean(field.source) && loadingOptions)}>
                    <option value="">{loadingOptions && field.source ? copy.loading : copy.choose}</option>
                    {field.options?.map((option) => <option key={option.value} value={option.value}>{copy.options[option.label]}</option>)}
                    {(options[field.key] ?? []).map((row) => (
                      <option key={String(row.id)} value={String(row.id)}>
                        {String((locale === "ar" ? row.name_ar : row.name_en) || row.name_en || row.name_ar || row.code || row.id)}
                      </option>
                    ))}
                  </select>
                ) : field.type === "textarea" ? (
                  <textarea {...props} rows={3} />
                ) : (
                  <input {...props} type={field.type === "email" ? "email" : "text"}
                    dir={field.key.endsWith("_ar") ? "rtl" : field.key.endsWith("_en") || ["decimal", "integer", "email"].includes(field.type ?? "") ? "ltr" : undefined}
                    inputMode={field.type === "decimal" ? "decimal" : field.type === "integer" ? "numeric" : undefined}
                  />
                )}
                {error && <p id={`${id}-error`} role="alert" className="text-xs text-destructive">{String(error)}</p>}
                {field.source && !loadingOptions && !options[field.key]?.length && (
                  <p id={`${id}-help`} className="text-xs text-muted-foreground">{copy.emptyOptions}</p>
                )}
              </div>
            );
          })}
        </div>
        {optionError && <div role="alert" className="text-sm text-destructive">{copy.loadError} <button type="button" onClick={() => setRetry((value) => value + 1)} className="underline">{copy.retry}</button></div>}
        {saveError && <div role="alert" className="rounded-lg border border-destructive/40 bg-destructive/10 p-3 text-sm text-destructive">{saveError}</div>}
        <div className="flex flex-wrap gap-3 border-t pt-4">
          <button type="submit" disabled={isSubmitting || loadingOptions || optionError || missingRequiredOptions || (tenant && !legalEntityId)} className="rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground disabled:opacity-60">{isSubmitting ? copy.saving : copy.save}</button>
          <button type="button" disabled={isSubmitting} onClick={onCancel} className="rounded-lg border px-4 py-2 text-sm font-semibold hover:bg-muted disabled:opacity-60">{copy.cancel}</button>
        </div>
      </form>
    </section>
  );
}
