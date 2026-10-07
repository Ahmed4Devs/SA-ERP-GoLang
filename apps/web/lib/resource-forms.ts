import { z } from "zod";
import en from "@/messages/en.json";

export type FieldLabel = keyof typeof en.resourceForm.fields | (string & {});
export type OptionLabel = keyof typeof en.resourceForm.options | (string & {});
export interface FormField {
  key: string;
  label: FieldLabel;
  type?: "text" | "email" | "decimal" | "integer" | "select" | "textarea" | "date";
  dateTime?: boolean;
  decimalPlaces?: number;
  showWhen?: { key: string; value: string };
  min?: number;
  max?: number;
  required?: boolean;
  initial?: string;
  pattern?: string;
  options?: { value: string; label: OptionLabel }[];
  source?: string;
  sourceTenant?: boolean;
}

export interface ResourceFormDefinition {
  fields: FormField[];
  defaults?: Record<string, unknown>;
  createPath?: string;
  notice?: string;
  submitLabel?: string;
  lines?: { fields: FormField[]; key: string; optional?: boolean };
}

const identity: FormField[] = [
  { key: "code", label: "code", required: true },
  { key: "name_ar", label: "name_ar", required: true },
  { key: "name_en", label: "name_en", required: true },
];
const names = identity.filter((field) => field.key !== "code");
const address: FormField[] = [
  { key: "city", label: "city", required: true },
  { key: "building_no", label: "building_no" },
  { key: "street_name", label: "street_name" },
  { key: "district", label: "district" },
  { key: "postal_code", label: "postal_code" },
  { key: "additional_no", label: "additional_no" },
  { key: "country_code", label: "country_code", required: true, initial: "SA", pattern: "^[A-Z]{2}$" },
];
const contact: FormField[] = [
  { key: "email", label: "email", type: "email" },
  { key: "phone", label: "phone" },
  { key: "contact_person", label: "contact_person" },
  { key: "tax_id", label: "tax_id", pattern: "^3[0-9]{13}3$" },
  { key: "cr_number", label: "cr_number", pattern: "^[0-9]{10}$" },
  { key: "currency_code", label: "currency_code", initial: "SAR", required: true, pattern: "^[A-Z]{3}$" },
];
const paymentTerms: ResourceFormDefinition = {
  fields: [...identity,
    { key: "due_days", label: "due_days", type: "integer", initial: "0", required: true },
    { key: "discount_days", label: "discount_days", type: "integer", initial: "0", required: true },
    { key: "discount_percent", label: "discount_percent", type: "decimal", initial: "0", required: true },
  ],
  defaults: { is_active: true },
};

// Each form is tied to an existing domain command. Financial state transitions
// continue to be validated and executed by the Go application services.
export const resourceForms: Record<string, ResourceFormDefinition> = {
  "/api/v1/org/organizations": { fields: identity },
  "/api/v1/org/legal-entities": { fields: [
    { key: "organization_id", label: "organization_id", type: "select", source: "/api/v1/org/organizations", sourceTenant: false, required: true },
    ...identity,
    { key: "cr_number", label: "cr_number", required: true, pattern: "^[0-9]{10}$" },
    { key: "base_currency", label: "currency_code", required: true, initial: "SAR", pattern: "^[A-Z]{3}$" },
  ] },
  "/api/v1/org/branches": { fields: identity },
  "/api/v1/org/departments": { fields: identity },
  "/api/v1/org/cost-centers": { fields: identity },
  "/api/v1/receivables/customers": { fields: [
    ...names,
    { key: "customer_type", label: "customer_type", type: "select", required: true, initial: "B2B", options: [
      { value: "B2B", label: "business" }, { value: "B2C", label: "individual" }, { value: "GOVERNMENT", label: "government" },
    ] },
    ...address, ...contact,
    { key: "payment_terms_id", label: "payment_terms_id", type: "select", source: "/api/v1/receivables/payment-terms" },
    { key: "credit_limit", label: "credit_limit", type: "decimal", initial: "0", required: true },
    { key: "max_overdue_days", label: "max_overdue_days", type: "integer", initial: "0", required: true },
  ], defaults: { block_action: "WARN" } },
  "/api/v1/payables/suppliers": { fields: [
    ...names,
    { key: "supplier_type", label: "supplier_type", type: "select", required: true, initial: "LOCAL", options: [
      { value: "LOCAL", label: "local" }, { value: "INTERNATIONAL", label: "international" }, { value: "GOVERNMENT", label: "government" },
    ] },
    ...address, ...contact,
    { key: "payment_terms_id", label: "payment_terms_id", type: "select", source: "/api/v1/payables/payment-terms" },
  ] },
  "/api/v1/receivables/payment-terms": paymentTerms,
  "/api/v1/payables/payment-terms": paymentTerms,
  "/api/v1/products/categories": { fields: [...identity, { key: "description", label: "description", type: "textarea" }] },
  "/api/v1/products/uoms": { fields: [
    ...identity,
    { key: "uom_type", label: "uom_type", type: "select", required: true, initial: "COUNT", options: [
      { value: "COUNT", label: "count" }, { value: "WEIGHT", label: "weight" }, { value: "LENGTH", label: "length" },
      { value: "VOLUME", label: "volume" }, { value: "AREA", label: "area" }, { value: "TIME", label: "time" },
    ] },
    { key: "symbol_ar", label: "symbol_ar" }, { key: "symbol_en", label: "symbol_en" },
  ] },
  "/api/v1/products/items": { fields: [
    ...identity,
    { key: "item_type", label: "item_type", type: "select", required: true, initial: "STOCK", options: [
      { value: "STOCK", label: "stock" }, { value: "SERVICE", label: "service" },
      { value: "NON_STOCK", label: "non_stock" }, { value: "ASSET", label: "asset" },
    ] },
    { key: "tracking_type", label: "tracking_type", type: "select", required: true, initial: "NONE", options: [
      { value: "NONE", label: "none" }, { value: "LOT", label: "lot" }, { value: "SERIAL", label: "serial" },
    ] },
    { key: "base_uom_id", label: "base_uom_id", type: "select", source: "/api/v1/products/uoms", required: true },
    { key: "category_id", label: "category_id", type: "select", source: "/api/v1/products/categories" },
    { key: "cost_price", label: "cost_price", type: "decimal", required: true, initial: "0" },
    { key: "list_price", label: "list_price", type: "decimal", required: true, initial: "0" },
    { key: "default_barcode", label: "default_barcode" },
  ], defaults: { is_sales_enabled: true, is_purchase_enabled: true } },
  "/api/v1/inventory/warehouses": { fields: [
    ...identity,
    { key: "warehouse_type", label: "warehouse_type", type: "select", required: true, initial: "PHYSICAL", options: [
      { value: "PHYSICAL", label: "physical" }, { value: "VIRTUAL", label: "virtual" },
      { value: "TRANSIT", label: "transit" }, { value: "CONSIGNMENT", label: "consignment" },
    ] },
    { key: "city", label: "city" }, { key: "address_line1", label: "address_line1" },
    { key: "country_code", label: "country_code", required: true, initial: "SA", pattern: "^[A-Z]{2}$" },
  ] },
};

export function createResourceSchema(definition: ResourceFormDefinition, messages: { required: string; invalid: string }) {
  return z.record(z.string(), z.string()).superRefine((values, context) => {
    for (const field of definition.fields) {
      if (field.showWhen && values[field.showWhen.key] !== field.showWhen.value) continue;
      const value = (values[field.key] ?? "").trim();
      let message: string | undefined;
      if (!value && field.required) message = messages.required;
      else if (value) {
        if (field.pattern && !new RegExp(field.pattern).test(value)) message = messages.invalid;
        if (field.type === "email" && !z.email().safeParse(value).success) message = messages.invalid;
        if (field.type === "decimal" && !new RegExp(`^\\d{1,18}(\\.\\d{1,${field.decimalPlaces ?? 6}})?$`).test(value)) message = messages.invalid;
        if (field.type === "integer" && (!/^\d+$/.test(value) || Number(value) > 2147483647)) message = messages.invalid;
        if (field.type === "date" && (!/^\d{4}-\d{2}-\d{2}$/.test(value) || Number.isNaN(Date.parse(value)) || new Date(value).toISOString().slice(0, 10) !== value)) message = messages.invalid;
        if (field.min !== undefined && Number(value) < field.min) message = messages.invalid;
        if (field.max !== undefined && Number(value) > field.max) message = messages.invalid;
        if (field.options && !field.options.some((option) => option.value === value)) message = messages.invalid;
        if (field.source && !z.uuid().safeParse(value).success) message = messages.invalid;
      }
      if (message) context.addIssue({ code: "custom", path: [field.key], message });
    }
  });
}

export function resourceFormPayload(definition: ResourceFormDefinition, values: Record<string, string>): Record<string, unknown> {
  const payload: Record<string, unknown> = { ...definition.defaults };
  for (const field of definition.fields) {
    if (field.showWhen && values[field.showWhen.key] !== field.showWhen.value) continue;
    const value = (values[field.key] ?? "").trim();
    if (value) payload[field.key] = field.type === "integer" ? Number(value) : field.dateTime ? `${value}T00:00:00Z` : value;
  }
  return payload;
}
