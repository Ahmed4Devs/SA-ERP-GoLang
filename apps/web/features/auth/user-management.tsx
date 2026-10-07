"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useLocale } from "@/components/layout/shell";
import { apiFetch } from "@/lib/api-client";
import { useAuth } from "@/lib/auth-context";

interface ManagedUser {
  id: string;
  username?: string;
  email?: string;
  display_name: string;
  status: "ACTIVE" | "DISABLED" | "SUSPENDED";
  version: number;
}
interface Role { id: string; name_ar: string; name_en: string; code: string }
interface UserRoleAssignment {
  id: string;
  role_name_ar: string;
  role_name_en: string;
  role_code: string;
  legal_entity_id?: string;
  legal_entity_name_ar?: string;
  legal_entity_name_en?: string;
  branch_name_ar?: string;
  branch_name_en?: string;
}

const createSchema = z.object({
  username: z.string().trim().min(3).max(100).regex(/^[\p{L}\p{N}._-]+$/u),
  display_name: z.string().trim().min(1).max(200),
  email: z.union([z.email(), z.literal("")]),
  password: z.string().min(8).max(1024),
  preferred_locale: z.enum(["ar-SA", "en"]),
  preferred_timezone: z.string().min(1),
});
type CreateValues = z.infer<typeof createSchema>;
const control = "min-h-10 w-full rounded-lg border bg-background px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-ring disabled:opacity-60";
const button = "rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground disabled:opacity-60";

export function UserManagement() {
  const { locale, t } = useLocale();
  const copy = t.userManagement;
  const { user: sessionUser, assignments, legalEntityId } = useAuth();
  const [users, setUsers] = useState<ManagedUser[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [search, setSearch] = useState("");
  const [query, setQuery] = useState("");
  const [offset, setOffset] = useState(0);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [roleError, setRoleError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [creating, setCreating] = useState(false);
  const [selected, setSelected] = useState<ManagedUser | null>(null);
  const [nextStatus, setNextStatus] = useState<ManagedUser["status"]>("ACTIVE");
  const [password, setPassword] = useState("");
  const [roleID, setRoleID] = useState("");
  const [entityID, setEntityID] = useState(legalEntityId ?? "");
  const [userAssignments, setUserAssignments] = useState<UserRoleAssignment[]>([]);
  const [assignmentsLoading, setAssignmentsLoading] = useState(false);
  const [assignmentsError, setAssignmentsError] = useState<string | null>(null);
  const [assignmentsRevision, setAssignmentsRevision] = useState(0);
  const request = useRef<AbortController | null>(null);
  const form = useForm<CreateValues>({ resolver: zodResolver(createSchema), defaultValues: {
    username: "", email: "", display_name: "", password: "", preferred_locale: "ar-SA", preferred_timezone: "Asia/Riyadh",
  } });
  const entityOptions = assignments.filter((item, index, all) => item.legal_entity_id && all.findIndex((other) => other.legal_entity_id === item.legal_entity_id) === index);
  const statusLabel = (value: ManagedUser["status"]) => value === "ACTIVE" ? copy.active : value === "DISABLED" ? copy.disabled : copy.suspended;

  const load = useCallback(async () => {
    request.current?.abort();
    const controller = new AbortController();
    request.current = controller;
    setLoading(true);
    setError(null);
    try {
      const result = await apiFetch<{ data: ManagedUser[] | null }>(`/api/v1/auth/users?${new URLSearchParams({ search: query, limit: "50", offset: String(offset) })}`, { signal: controller.signal });
      if (!controller.signal.aborted) {
        const rows = result.data ?? [];
        setUsers(rows);
        setSelected((current) => current ? rows.find((row) => row.id === current.id) ?? null : null);
      }
    } catch (failure) {
      if (!controller.signal.aborted) { setUsers([]); setError(failure instanceof Error ? failure.message : copy.unavailable); }
    } finally { if (!controller.signal.aborted) setLoading(false); }
  }, [copy.unavailable, offset, query]);

  useEffect(() => { void load(); return () => request.current?.abort(); }, [load]);
  useEffect(() => {
    const controller = new AbortController();
    apiFetch<{ data: Role[] | null }>("/api/v1/authz/roles", { signal: controller.signal })
      .then((result) => { if (!controller.signal.aborted) setRoles(result.data ?? []); })
      .catch((failure) => { if (!controller.signal.aborted) setRoleError(failure instanceof Error ? failure.message : copy.unavailable); });
    return () => controller.abort();
  }, [copy.unavailable]);
  useEffect(() => { setPassword(""); setRoleID(""); setEntityID(legalEntityId ?? ""); }, [selected?.id, legalEntityId]);
  useEffect(() => {
    setUserAssignments([]); setAssignmentsError(null);
    if (!selected?.id) return;
    const controller = new AbortController();
    setAssignmentsLoading(true);
    apiFetch<{ data: UserRoleAssignment[] | null }>(`/api/v1/authz/assignments?user_id=${encodeURIComponent(selected.id)}`, { signal: controller.signal })
      .then((result) => { if (!controller.signal.aborted) setUserAssignments(result.data ?? []); })
      .catch((failure) => { if (!controller.signal.aborted) setAssignmentsError(failure instanceof Error ? failure.message : copy.unavailable); })
      .finally(() => { if (!controller.signal.aborted) setAssignmentsLoading(false); });
    return () => controller.abort();
  }, [selected?.id, assignmentsRevision, copy.unavailable]);

  const create = async (values: CreateValues) => {
    setError(null); setSuccess(null);
    try {
      await apiFetch<ManagedUser>("/api/v1/auth/users", { method: "POST", body: JSON.stringify(values) });
      form.reset(); setCreating(false); setSuccess(copy.created); await load();
    } catch (failure) { setError(failure instanceof Error ? failure.message : copy.unavailable); }
  };

  async function update(path: string, payload: Record<string, unknown>) {
    if (!selected) return;
    setBusy(true); setError(null); setSuccess(null);
    try {
      await apiFetch(path, { method: "POST", body: JSON.stringify(payload) });
      if (path === "/api/v1/authz/assignments") setAssignmentsRevision((value) => value + 1);
      setPassword(""); setSuccess(copy.success); await load();
    } catch (failure) { setError(failure instanceof Error ? failure.message : copy.unavailable); }
    finally { setBusy(false); }
  }

  return (
    <section className="mb-8 space-y-5 rounded-xl border bg-card p-5" aria-labelledby="user-management-title">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div><h1 id="user-management-title" className="text-xl font-bold">{copy.title}</h1><p className="mt-1 text-sm text-muted-foreground">{copy.subtitle}</p></div>
        <button type="button" onClick={() => { setCreating(true); setSelected(null); setSuccess(null); }} className={button}>{copy.create}</button>
      </div>
      {error && <div role="alert" className="rounded-lg border border-destructive/40 bg-destructive/10 p-3 text-sm text-destructive">{error}</div>}
      {success && <div role="status" className="rounded-lg bg-emerald-500/10 p-3 text-sm text-emerald-700 dark:text-emerald-300">{success}</div>}
      {creating && <form onSubmit={(event) => void form.handleSubmit(create)(event)} noValidate className="space-y-4 rounded-lg border p-4">
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {([
            ["username", copy.username, "text"], ["display_name", copy.displayName, "text"], ["email", copy.email, "email"], ["password", copy.password, "password"], ["preferred_timezone", copy.timezone, "text"],
          ] as const).map(([key, label, type]) => <div key={key} className="space-y-1.5">
            <label htmlFor={`user-${key}`} className="text-sm font-medium">{label}</label>
            <input id={`user-${key}`} {...form.register(key)} type={type} autoComplete={type === "password" ? "new-password" : "off"} aria-invalid={Boolean(form.formState.errors[key])} aria-describedby={form.formState.errors[key] ? `user-${key}-error` : undefined} disabled={form.formState.isSubmitting} className={control} />
            {form.formState.errors[key] && <p id={`user-${key}-error`} role="alert" className="text-xs text-destructive">{copy.required}{type === "password" ? ` ${copy.passwordHint}` : ""}</p>}
          </div>)}
          <label className="space-y-1.5 text-sm font-medium">{copy.locale}<select {...form.register("preferred_locale")} className={control}><option value="ar-SA">{copy.arabic}</option><option value="en">{copy.english}</option></select></label>
        </div>
        <div className="flex gap-3"><button type="submit" disabled={form.formState.isSubmitting} className={button}>{form.formState.isSubmitting ? copy.saving : copy.save}</button><button type="button" onClick={() => { setCreating(false); form.reset(); }} disabled={form.formState.isSubmitting} className="rounded-lg border px-4 py-2 text-sm">{copy.cancel}</button></div>
      </form>}
      <form onSubmit={(event) => { event.preventDefault(); setQuery(search.trim()); setOffset(0); setSelected(null); }} className="flex flex-wrap gap-2">
        <input aria-label={copy.search} value={search} onChange={(event) => setSearch(event.target.value)} maxLength={200} className={`${control} max-w-sm`} placeholder={copy.search} />
        <button className={button} type="submit">{copy.searchButton}</button>
        <button type="button" onClick={() => void load()} disabled={loading} className="rounded-lg border px-4 py-2 text-sm">{copy.refresh}</button>
      </form>
      <div className="overflow-x-auto rounded-lg border">
        <table className="w-full text-start text-sm"><thead className="bg-muted"><tr>{[copy.displayName, copy.username, copy.email, copy.status, copy.manage].map((label) => <th key={label} className="p-3 text-start">{label}</th>)}</tr></thead>
          <tbody className="divide-y">{loading ? <tr><td colSpan={5} className="p-5 text-center">{copy.loading}</td></tr> : users.length === 0 ? <tr><td colSpan={5} className="p-5 text-center">{copy.empty}</td></tr> : users.map((row) => <tr key={row.id} className={selected?.id === row.id ? "bg-primary/5" : ""}>
            <td className="p-3">{row.display_name}</td><td className="p-3">{row.username}</td><td className="p-3">{row.email}</td><td className="p-3">{statusLabel(row.status)}</td><td className="p-3"><button type="button" disabled={busy} onClick={() => { setSelected(row); setNextStatus(row.status); setCreating(false); form.reset(); setSuccess(null); }} className="rounded border px-3 py-1.5 hover:bg-muted" aria-label={`${copy.manage}: ${row.display_name}`}>{copy.manage}</button></td>
          </tr>)}</tbody>
        </table>
      </div>
      <div className="flex justify-end gap-3"><button type="button" disabled={loading || offset === 0} onClick={() => setOffset((value) => Math.max(0, value - 50))} className="rounded border px-3 py-2 text-sm disabled:opacity-50">{copy.previous}</button><button type="button" disabled={loading || users.length < 50} onClick={() => setOffset((value) => value + 50)} className="rounded border px-3 py-2 text-sm disabled:opacity-50">{copy.next}</button></div>
      {selected && <div className="space-y-5 rounded-lg border p-4">
        <h2 className="font-semibold">{copy.manage}: {selected.display_name}</h2>
        <form onSubmit={(event) => { event.preventDefault(); void update(`/api/v1/auth/users/${selected.id}/status`, { status: nextStatus, expected_version: selected.version }); }} className="flex flex-wrap items-end gap-3">
          <label className="min-w-48 space-y-1.5 text-sm">{copy.status}<select aria-label={copy.status} value={nextStatus} onChange={(event) => setNextStatus(event.target.value as ManagedUser["status"])} disabled={busy} className={control}>{(["ACTIVE", "SUSPENDED", "DISABLED"] as const).map((status) => <option key={status} value={status} disabled={sessionUser?.id === selected.id && status !== "ACTIVE"}>{statusLabel(status)}</option>)}</select></label>
          <button type="submit" className={button} disabled={busy || nextStatus === selected.status}>{copy.changeStatus}</button>
        </form>
        <form onSubmit={(event) => { event.preventDefault(); void update(`/api/v1/auth/users/${selected.id}/password`, { password, expected_version: selected.version }); }} className="space-y-2 border-t pt-4">
          <div className="flex flex-wrap items-end gap-3"><label className="min-w-48 space-y-1.5 text-sm">{copy.newPassword}<input aria-label={copy.newPassword} type="password" autoComplete="new-password" value={password} onChange={(event) => setPassword(event.target.value)} minLength={8} maxLength={1024} required disabled={busy} className={control} /></label><button type="submit" disabled={busy || password.length < 8} className={button}>{copy.resetPassword}</button></div><p className="text-xs text-muted-foreground">{copy.passwordHint}</p>
        </form>
        <form onSubmit={(event) => { event.preventDefault(); void update("/api/v1/authz/assignments", { user_id: selected.id, role_id: roleID, ...(entityID ? { legal_entity_id: entityID } : {}) }); }} className="flex flex-wrap items-end gap-3 border-t pt-4">
          <label className="min-w-48 space-y-1.5 text-sm">{copy.role}<select aria-label={copy.role} value={roleID} onChange={(event) => setRoleID(event.target.value)} required disabled={busy || Boolean(roleError)} className={control}><option value="">{copy.chooseRole}</option>{roles.map((role) => <option key={role.id} value={role.id}>{(locale === "ar" ? role.name_ar : role.name_en) || role.code}</option>)}</select></label>
          <label className="min-w-48 space-y-1.5 text-sm">{copy.scope}<select aria-label={copy.scope} value={entityID} onChange={(event) => setEntityID(event.target.value)} disabled={busy} className={control}><option value="">{copy.globalScope}</option>{entityOptions.map((entity) => <option key={entity.id} value={entity.legal_entity_id}>{(locale === "ar" ? entity.legal_entity_name_ar : entity.legal_entity_name_en) || entity.legal_entity_id}</option>)}</select></label>
          <button type="submit" disabled={busy || !roleID || Boolean(roleError)} className={button}>{copy.assignRole}</button>
        </form>
        {roleError && <p role="alert" className="text-sm text-destructive">{roleError}</p>}
        <section aria-label={copy.viewAssignments} className="space-y-2 border-t pt-4">
          <h3 className="text-sm font-semibold">{copy.viewAssignments}</h3>
          {assignmentsLoading && <p className="text-sm text-muted-foreground">{copy.loading}</p>}
          {assignmentsError && <p role="alert" className="text-sm text-destructive">{assignmentsError}</p>}
          {userAssignments.map((assignment) => <div key={assignment.id} className="rounded-lg bg-muted p-3 text-sm">
            <span className="font-semibold">{(locale === "ar" ? assignment.role_name_ar : assignment.role_name_en) || assignment.role_code}</span>
            <span className="ms-3 text-muted-foreground">{assignment.legal_entity_id ? (locale === "ar" ? assignment.legal_entity_name_ar : assignment.legal_entity_name_en) || assignment.legal_entity_id : copy.globalScope}</span>
            {(assignment.branch_name_ar || assignment.branch_name_en) && <span className="ms-2">{locale === "ar" ? assignment.branch_name_ar : assignment.branch_name_en}</span>}
          </div>)}
        </section>
      </div>}
    </section>
  );
}
