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
}

interface Organization {
  id: string;
}

interface LegalEntity {
  id: string;
  name_ar?: string;
  name_en?: string;
}

interface AuthContextValue {
  user: SessionUser | null;
  assignments: RoleAssignment[];
  legalEntityId: string | null;
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
    if (!loading && !user && !authError && pathname !== "/login") {
      router.replace(`/login?next=${encodeURIComponent(pathname)}`);
    }
  }, [authError, loading, pathname, router, user]);

  const selectLegalEntity = useCallback((id: string) => {
    setLegalEntityId(id);
    window.localStorage.setItem("erp_legal_entity_id", id);
  }, []);

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
    () => ({ user, assignments, legalEntityId, loading, authError, scopeError, refresh, selectLegalEntity, logout }),
    [user, assignments, legalEntityId, loading, authError, scopeError, refresh, selectLegalEntity, logout],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) throw new Error("useAuth must be used inside AuthProvider");
  return context;
}
