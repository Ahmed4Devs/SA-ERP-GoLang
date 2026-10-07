"use client";

import { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";
import { usePathname, useRouter } from "next/navigation";
import { APIError, apiFetch } from "@/lib/api-client";

export interface SessionUser {
  id: string;
  username?: string;
  email?: string;
  display_name: string;
}

export interface RoleAssignment {
  id: string;
  role_code: string;
  legal_entity_id?: string;
  legal_entity_name_ar?: string;
  legal_entity_name_en?: string;
  branch_id?: string;
  branch_name_ar?: string;
  branch_name_en?: string;
}

interface Organization {
  id: string;
}

interface LegalEntity {
  id: string;
  name_ar?: string;
  name_en?: string;
}

export interface Branch { id: string; name_ar?: string; name_en?: string; code?: string }

interface AuthContextValue {
  user: SessionUser | null;
  assignments: RoleAssignment[];
  legalEntityId: string | null;
  branchId: string | null;
  branches: Branch[];
  branchesLoading: boolean;
  branchError: string | null;
  canSelectAllBranches: boolean;
  selectBranch: (id: string) => void;
  loading: boolean;
  authError: string | null;
  scopeError: string | null;
  refresh: () => Promise<void>;
  selectLegalEntity: (id: string) => void;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const pathname = usePathname();
  const [user, setUser] = useState<SessionUser | null>(null);
  const [assignments, setAssignments] = useState<RoleAssignment[]>([]);
  const [legalEntityId, setLegalEntityId] = useState<string | null>(null);
  const [branchId, setBranchId] = useState<string | null>(null);
  const [branches, setBranches] = useState<Branch[]>([]);
  const [branchesLoading, setBranchesLoading] = useState(false);
  const [branchError, setBranchError] = useState<string | null>(null);
  const canSelectAllBranches = assignments.some(assignment => !assignment.branch_id && (!assignment.legal_entity_id || assignment.legal_entity_id === legalEntityId));
  const [loading, setLoading] = useState(true);
  const [authError, setAuthError] = useState<string | null>(null);
  const [scopeError, setScopeError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    setLoading(true);
    setAuthError(null);
    setScopeError(null);
    try {
      const session = await apiFetch<{ user: SessionUser }>("/api/v1/auth/me");
      setUser(session.user);
      try {
        const assignmentResponse = await apiFetch<{ data?: RoleAssignment[] }>("/api/v1/authz/assignments/me");
        let availableAssignments = Array.isArray(assignmentResponse.data) ? assignmentResponse.data : [];
        let scopedAssignments = availableAssignments.filter((assignment) => assignment.legal_entity_id);

        // Global administrators have an intentionally unscoped role. Resolve
        // real entities from the API instead of inventing a tenant in-browser.
        if (scopedAssignments.length === 0 && availableAssignments.some((assignment) => !assignment.legal_entity_id)) {
          const organizations = await apiFetch<{ data?: Organization[] }>("/api/v1/org/organizations");
          const discovered = (
            await Promise.all(
              (organizations.data ?? []).map((organization) =>
                apiFetch<{ data?: LegalEntity[] }>(`/api/v1/org/legal-entities?organization_id=${organization.id}`),
              ),
            )
          ).flatMap((response) => response.data ?? []);
          scopedAssignments = discovered.map((entity) => ({
            id: `entity-${entity.id}`,
            role_code: "GLOBAL",
            legal_entity_id: entity.id,
            legal_entity_name_ar: entity.name_ar,
            legal_entity_name_en: entity.name_en,
          }));
          availableAssignments = [...availableAssignments, ...scopedAssignments];
        }

        const savedEntity = window.localStorage.getItem("erp_legal_entity_id");
        const selected = scopedAssignments.some((assignment) => assignment.legal_entity_id === savedEntity)
          ? savedEntity
          : scopedAssignments[0]?.legal_entity_id ?? null;

        setAssignments(availableAssignments);
        setLegalEntityId(selected);
        if (selected) window.localStorage.setItem("erp_legal_entity_id", selected);
        else window.localStorage.removeItem("erp_legal_entity_id");
      } catch (requestError) {
        setAssignments([]);
        setLegalEntityId(null);
        setScopeError(requestError instanceof Error ? requestError.message : "Unable to load account scopes");
      }
    } catch (requestError) {
      if (requestError instanceof APIError && requestError.status === 401) {
        setUser(null);
        setAssignments([]);
        setLegalEntityId(null);
        window.localStorage.removeItem("erp_legal_entity_id");
      } else {
        setAuthError(requestError instanceof Error ? requestError.message : "Unable to verify the session");
      }
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  useEffect(() => {
    const handleExpiredSession = () => {
      setUser(null);
      setAssignments([]);
      setLegalEntityId(null);
      setAuthError(null);
      setScopeError(null);
      window.localStorage.removeItem("erp_legal_entity_id");
      if (pathname !== "/login") {
        router.replace(`/login?next=${encodeURIComponent(pathname)}`);
      }
    };
    window.addEventListener("erp:session-expired", handleExpiredSession);
    return () => window.removeEventListener("erp:session-expired", handleExpiredSession);
  }, [pathname, router]);

  useEffect(() => {
    if (!loading && !user && !authError && pathname !== "/login") {
      router.replace(`/login?next=${encodeURIComponent(pathname)}`);
    }
  }, [authError, loading, pathname, router, user]);

  useEffect(() => {
    const controller = new AbortController();
    setBranches([]); setBranchId(null); setBranchError(null);
    if (!legalEntityId || !user) { setBranchesLoading(false); return; }
    const entity = legalEntityId;
    setBranchesLoading(true);
    async function loadBranches() {
      try {
        const scoped = assignments.filter(item => item.legal_entity_id === entity && item.branch_id);
        const available = canSelectAllBranches
          ? (await apiFetch<{ data?: Branch[] }>(`/api/v1/org/branches?legal_entity_id=${encodeURIComponent(entity)}`, { signal: controller.signal })).data ?? []
          : scoped.map(item => ({ id: item.branch_id!, name_ar: item.branch_name_ar, name_en: item.branch_name_en }));
        if (controller.signal.aborted) return;
        const unique = available.filter((branch, index, all) => all.findIndex(other => other.id === branch.id) === index);
        setBranches(unique);
        const saved = window.localStorage.getItem(`erp_branch_${entity}`);
        setBranchId(unique.some(branch => branch.id === saved) ? saved : canSelectAllBranches ? null : unique[0]?.id ?? null);
      } catch (failure) {
        if (!controller.signal.aborted) setBranchError(failure instanceof Error ? failure.message : "Unable to load branches");
      } finally { if (!controller.signal.aborted) setBranchesLoading(false); }
    }
    void loadBranches();
    return () => controller.abort();
  }, [assignments, canSelectAllBranches, legalEntityId, user]);

  const selectBranch = useCallback((id: string) => {
    if (!legalEntityId || (id ? !branches.some(branch => branch.id === id) : !canSelectAllBranches)) return;
    setBranchId(id || null);
    window.localStorage.setItem(`erp_branch_${legalEntityId}`, id);
  }, [branches, canSelectAllBranches, legalEntityId]);

  const selectLegalEntity = useCallback((id: string) => {
    if (!assignments.some(assignment => assignment.legal_entity_id === id)) return;
    setBranchId(null); setBranches([]);
    setLegalEntityId(id);
    window.localStorage.setItem("erp_legal_entity_id", id);
  }, [assignments]);

  const logout = useCallback(async () => {
    try {
      await apiFetch("/api/v1/auth/logout", { method: "POST" });
    } finally {
      setUser(null);
      setAssignments([]);
      setLegalEntityId(null);
      setAuthError(null);
      setScopeError(null);
      window.localStorage.removeItem("erp_legal_entity_id");
      router.replace("/login");
    }
  }, [router]);

  const value = useMemo(
    () => ({ user, assignments, legalEntityId, branchId, branches, branchesLoading, branchError, canSelectAllBranches, selectBranch, loading, authError, scopeError, refresh, selectLegalEntity, logout }),
    [user, assignments, legalEntityId, branchId, branches, branchesLoading, branchError, canSelectAllBranches, selectBranch, loading, authError, scopeError, refresh, selectLegalEntity, logout],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) throw new Error("useAuth must be used inside AuthProvider");
  return context;
}
