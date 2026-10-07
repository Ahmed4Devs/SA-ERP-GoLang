import { describe, expect, it } from "vitest";
import {
  appendQueryParams,
  appendTenant,
  buildMutationRequest,
  buildResourcePath,
  normalizeRows,
} from "@/components/modules/module-api-view";
import { moduleResources } from "@/lib/module-resources";

describe("module API resources", () => {
  it("uses mounted API paths and never uses the former zero tenant id", () => {
    const paths = Object.values(moduleResources).flat().map((resource) => resource.path);
    expect(paths.length).toBeGreaterThan(50);
    expect(paths.every((path) => path.startsWith("/api/v1/"))).toBe(true);
    expect(paths.some((path) => path.includes("00000000-0000-0000-0000-000000000000"))).toBe(false);
  });

  it("adds the selected tenant without corrupting existing query strings", () => {
    expect(appendTenant("/api/v1/items", "entity-1")).toBe("/api/v1/items?legal_entity_id=entity-1");
    expect(appendTenant("/api/v1/items?limit=20", "entity-1")).toBe("/api/v1/items?limit=20&legal_entity_id=entity-1");
    expect(appendTenant("/api/v1/auth/me", "entity-1", false)).toBe("/api/v1/auth/me");
    expect(appendTenant("/api/v1/items?legal_entity_id=old", "entity-1")).toBe("/api/v1/items?legal_entity_id=entity-1");
  });

  it("builds required dependent-filter paths without sending incomplete requests", () => {
    const statements = moduleResources.banking.find((resource) => resource.key === "statements");
    const payslips = moduleResources.payroll.find((resource) => resource.key === "payslips");
    expect(statements?.filters?.[0]?.optionsPath).toBe("/api/v1/banking/bank-accounts");
    expect(payslips?.filters?.[0]?.optionsPath).toBe("/api/v1/payroll/periods");
    expect(buildResourcePath(statements!, "entity-1", { bank_account_id: "bank-1" })).toBe(
      "/api/v1/banking/statements?legal_entity_id=entity-1&bank_account_id=bank-1",
    );
    expect(buildResourcePath(payslips!, "entity-1", { period_id: "period-1" })).toBe(
      "/api/v1/payroll/payslips?legal_entity_id=entity-1&period_id=period-1",
    );
    expect(appendQueryParams("/api/v1/items?limit=20", { status: "ACTIVE" })).toBe(
      "/api/v1/items?limit=20&status=ACTIVE",
    );
  });

  it("normalizes common API envelopes", () => {
    expect(normalizeRows({ data: [{ id: "1" }] })).toEqual([{ id: "1" }]);
    expect(normalizeRows([{ id: "2" }])).toEqual([{ id: "2" }]);
    expect(normalizeRows({ total: 4 })).toEqual([{ total: 4 }]);
    expect(normalizeRows({ data: null })).toEqual([]);
    expect(normalizeRows(null)).toEqual([]);
    expect(normalizeRows({ data: { id: "3" } })).toEqual([{ id: "3" }]);
  });

  it("builds tenant-scoped mutation requests and rejects external endpoints", () => {
    const request = buildMutationRequest(
      "/api/v1/products/categories",
      "POST",
      JSON.stringify({ code: "TEST" }),
      "entity-1",
    );
    expect(request.path).toBe("/api/v1/products/categories?legal_entity_id=entity-1");
    expect(JSON.parse(String(request.init.body))).toEqual({ code: "TEST", legal_entity_id: "entity-1" });
    expect(() => buildMutationRequest("https://example.com/api/v1/items", "POST", "{}", "entity-1")).toThrow(
      /local \/api\/v1\//,
    );
    expect(() => buildMutationRequest("/api/v1/items", "PATCH", "not-json", "entity-1")).toThrow(/valid JSON/);
    expect(() => buildMutationRequest("/api/v1/items", "POST", '{"legal_entity_id":"entity-2"}', "entity-1")).toThrow(/match the selected legal entity/);
  });
});
