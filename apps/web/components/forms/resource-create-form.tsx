"use client";

import { useEffect, useMemo, useState } from "react";
import { useForm, useWatch } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useLocale } from "@/components/layout/shell";
import { apiFetch } from "@/lib/api-client";
import { createResourceSchema, resourceFormPayload, type FormField, type ResourceFormDefinition } from "@/lib/resource-forms";
import { fieldLabel } from "@/lib/record-presentation";
import { loadReferenceRows, queryPath, type RecordRow } from "@/lib/resource-data";

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
const initialFields = (fields: FormField[]) => Object.fromEntries(fields.map(field => [field.key, field.initial ?? ""]));

export function ResourceCreateForm({ definition, path, title, legalEntityId, tenant, initialValues, onCancel, onSaved }: ResourceCreateFormProps) {
  const { locale, t } = useLocale();
  const copy = t.resourceForm;
  const [options, setOptions] = useState<Record<string, RecordRow[]>>({});
  const [loadingOptions, setLoadingOptions] = useState(true);
  const [optionError, setOptionError] = useState(false);
  const [retry, setRetry] = useState(0);
  const [saveError, setSaveError] = useState<string | null>(null);
  const [lines, setLines] = useState<Record<string, string>[]>(() => definition.lines && !definition.lines.optional ? [initialFields(definition.lines.fields)] : []);
  const [lineErrors, setLineErrors] = useState<Record<string, string>[]>([]);
  const fields = useMemo(() => [...definition.fields, ...(definition.lines?.fields ?? [])], [definition]);
  const schema = useMemo(() => createResourceSchema(definition, copy), [definition, copy]);
  const { control, register, handleSubmit, formState: { errors, isSubmitting } } = useForm<Record<string, string>>({
    resolver: zodResolver(schema),
    defaultValues: { ...initialFields(definition.fields), ...initialValues },
  });

  const headerValues = useWatch({ control });
  const sourceScope = JSON.stringify(fields.filter(field => field.source).map(field => ({
    source: field.source!,
    path: field.showWhen && headerValues[field.showWhen.key] !== field.showWhen.value ? "" : field.source!.replace(/\{([^}]+)\}/g, (_, key: string) => headerValues[key] ? encodeURIComponent(headerValues[key]!) : "{missing}"),
    tenant: field.sourceTenant !== false,
  })));

  useEffect(() => {
    const controller = new AbortController();
    setLoadingOptions(true); setOptionError(false); setOptions({});
    async function loadOptions() {
      try {
        const sources = JSON.parse(sourceScope) as { source: string; path: string; tenant: boolean }[];
        const references = sources.filter((reference, index) => reference.path && !reference.path.includes("{missing}") && sources.findIndex(other => other.source === reference.source && other.path) === index);
        const loaded = await Promise.all(references.map(async reference => {
          let rows = await loadReferenceRows(reference.path, legalEntityId, controller.signal, reference.tenant);
          if (reference.source === "/api/v1/accounting/accounts") rows = rows.filter(row => row.is_group !== true && row.allow_posting !== false);
          return [reference.source, rows] as const;
        }));
        if (!controller.signal.aborted) setOptions(Object.fromEntries(loaded));
      } catch {
        if (!controller.signal.aborted) setOptionError(true);
      } finally { if (!controller.signal.aborted) setLoadingOptions(false); }
    }
    void loadOptions();
    return () => controller.abort();
  }, [sourceScope, legalEntityId, retry]);

  const submit = handleSubmit(async values => {
    setSaveError(null);
    const referencesValid = (fields: FormField[], row: Record<string, string>) => fields.every(field => !field.source || !row[field.key] || (field.showWhen && row[field.showWhen.key] !== field.showWhen.value) || options[field.source]?.some(option => String(option.id) === row[field.key]));
    if (!referencesValid(definition.fields, values) || (definition.lines && lines.some(line => !referencesValid(definition.lines!.fields, line)))) {
      setSaveError(copy.invalid); return;
    }
    const payload = resourceFormPayload(definition, values);
    if (definition.lines) {
      if (!lines.length && !definition.lines.optional) { setSaveError(t.business.requiredLines); return; }
      const lineDefinition = { fields: definition.lines.fields };
      const lineSchema = createResourceSchema(lineDefinition, copy);
      const validations = lines.map(line => lineSchema.safeParse(line));
      setLineErrors(validations.map(result => result.success ? {} : Object.fromEntries(result.error.issues.map(issue => [String(issue.path[0]), issue.message]))));
      if (validations.some(result => !result.success)) return;
      payload[definition.lines.key] = lines.map(line => resourceFormPayload(lineDefinition, line));
    }
    try {
      if (tenant) {
        if (!legalEntityId) return;
        payload.legal_entity_id = legalEntityId;
      }
      await apiFetch(queryPath(definition.createPath ?? path, tenant && legalEntityId ? { legal_entity_id: legalEntityId } : {}), { method: "POST", body: JSON.stringify(payload) });
      await onSaved();
    } catch (error) { setSaveError(error instanceof Error ? error.message : copy.saveError); }
  });
  const missingRequiredOptions = [...definition.fields, ...(lines.length ? definition.lines?.fields ?? [] : [])].some(field => field.required && field.source && (!field.showWhen || headerValues[field.showWhen.key] === field.showWhen.value) && !options[field.source]?.length);
  const controlClass = "min-h-10 w-full rounded-lg border bg-background px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-ring disabled:opacity-60";

  function renderField(field: FormField, lineIndex?: number) {
    if (field.showWhen && headerValues[field.showWhen.key] !== field.showWhen.value) return null;
    const id = lineIndex === undefined ? `create-${field.key}` : `line-${lineIndex}-${field.key}`;
    const error = lineIndex === undefined ? errors[field.key]?.message : lineErrors[lineIndex]?.[field.key];
    const referenceOptions = options[field.source ?? ""] ?? [];
    const props = {
      id,
      ...(lineIndex === undefined ? register(field.key) : { value: lines[lineIndex][field.key] ?? "", onChange: (event: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => setLines(current => current.map((line, index) => index === lineIndex ? { ...line, [field.key]: event.target.value } : line)) }),
      "aria-invalid": Boolean(error),
      "aria-describedby": error ? `${id}-error` : field.source && !referenceOptions.length ? `${id}-help` : undefined,
      "aria-required": Boolean(field.required),
      disabled: isSubmitting,
      className: controlClass,
    };
    const optionLabels = copy.options as Record<string, string>;
    return <div key={id} className="min-w-0 space-y-1.5">
      <label htmlFor={id} className="text-sm font-medium">{fieldLabel(field.label, locale)}{field.required ? " *" : ` (${copy.optional})`}</label>
      {field.type === "select" ? <select {...props} value={lineIndex === undefined ? headerValues[field.key] ?? "" : lines[lineIndex][field.key] ?? ""} disabled={isSubmitting || (Boolean(field.source) && loadingOptions)}>
        <option value="">{loadingOptions && field.source ? copy.loading : copy.choose}</option>
        {field.options?.map(option => <option key={option.value} value={option.value}>{optionLabels[option.label] ?? fieldLabel(option.label, locale)}</option>)}
        {referenceOptions.map(row => <option key={String(row.id)} value={String(row.id)}>{[row.code ?? row.account_code ?? row.period_year, (locale === "ar" ? row.name_ar ?? row.period_name_ar : row.name_en ?? row.period_name_en) || row.name_en || row.name_ar || row.period_name_en || row.bank_name_ar || row.bank_name_en || row.document_number || row.number || row.period_month || row.id].filter(Boolean).join(" · ")}</option>)}
      </select> : field.type === "textarea" ? <textarea {...props} rows={3} /> : <input {...props} type={field.type === "email" ? "email" : field.type === "date" ? "date" : "text"}
        dir={field.key.endsWith("_ar") ? "rtl" : field.key.endsWith("_en") || ["decimal", "integer", "email", "date"].includes(field.type ?? "") ? "ltr" : undefined}
        inputMode={field.type === "decimal" ? "decimal" : field.type === "integer" ? "numeric" : undefined} />}
      {error && <p id={`${id}-error`} role="alert" className="text-xs text-destructive">{String(error)}</p>}
      {field.source && !loadingOptions && !referenceOptions.length && <p id={`${id}-help`} className="text-xs text-muted-foreground">{copy.emptyOptions}</p>}
    </div>;
  }

  return <section className="rounded-xl border bg-card p-5 shadow-sm" aria-labelledby="resource-form-title">
    <h2 id="resource-form-title" className="mb-5 text-lg font-semibold">{copy.title}: {title}</h2>
    {definition.notice && <p className="mb-4 rounded-lg border bg-muted/30 p-3 text-sm">{(t.business as Record<string, string>)[definition.notice]}</p>}
    <form onSubmit={submit} noValidate className="space-y-5">
      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">{definition.fields.map(field => renderField(field))}</div>
      {definition.lines && <fieldset className="space-y-4 border-t pt-4"><legend className="px-2 font-medium">{definition.lines.key === "lines" ? t.business.lines : fieldLabel(definition.lines.key, locale)}</legend>
        {lines.map((_, index) => <div key={index} className="space-y-3 rounded-lg border p-4"><div className="flex items-center justify-between"><span className="text-sm font-medium">{index + 1}</span><button type="button" disabled={isSubmitting} onClick={() => { setLines(current => current.filter((_, i) => i !== index)); setLineErrors([]); }} className="rounded-lg border px-3 py-1 text-sm">{t.business.removeLine}</button></div><div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">{definition.lines!.fields.map(field => renderField(field, index))}</div></div>)}
        <button type="button" disabled={isSubmitting} onClick={() => setLines(current => [...current, initialFields(definition.lines!.fields)])} className="rounded-lg border px-3 py-2 text-sm">{t.business.addLine}</button>
      </fieldset>}
      {optionError && <div role="alert" className="text-sm text-destructive">{copy.loadError} <button type="button" onClick={() => setRetry(value => value + 1)} className="underline">{copy.retry}</button></div>}
      {saveError && <div role="alert" className="rounded-lg border border-destructive/40 bg-destructive/10 p-3 text-sm text-destructive">{saveError}</div>}
      <div className="flex flex-wrap gap-3 border-t pt-4">
        <button type="submit" disabled={isSubmitting || loadingOptions || optionError || missingRequiredOptions || (tenant && !legalEntityId)} className="rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground disabled:opacity-60">{isSubmitting ? copy.saving : definition.submitLabel ? (t.business as Record<string, string>)[definition.submitLabel] : copy.save}</button>
        <button type="button" disabled={isSubmitting} onClick={onCancel} className="rounded-lg border px-4 py-2 text-sm font-semibold hover:bg-muted disabled:opacity-60">{copy.cancel}</button>
      </div>
    </form>
  </section>;
}
