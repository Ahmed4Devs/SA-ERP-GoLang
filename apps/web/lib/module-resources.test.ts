import { describe, expect, it } from "vitest";
import { appendTenant, normalizeRows } from "@/components/modules/module-api-view";
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
  });

  it("normalizes common API envelopes", () => {
    expect(normalizeRows({ data: [{ id: "1" }] })).toEqual([{ id: "1" }]);
    expect(normalizeRows([{ id: "2" }])).toEqual([{ id: "2" }]);
    expect(normalizeRows({ total: 4 })).toEqual([{ total: 4 }]);
  });
});
