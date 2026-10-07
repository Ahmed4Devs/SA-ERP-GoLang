import ar from "@/messages/ar.json";
import en from "@/messages/en.json";

export function fieldKey(key: string): string {
  return key.replace(/([A-Z]+)([A-Z][a-z])/g, "$1_$2").replace(/([a-z0-9])([A-Z])/g, "$1_$2").toLowerCase();
}
export function fieldLabel(key: string, locale: string): string {
  const normalized = fieldKey(key);
  const messages = locale === "ar" ? ar : en;
  const dictionary = messages.resourceForm.fields as Record<string, string>;
  const words = messages.record.words as Record<string, string>;
  if (dictionary[normalized]) return dictionary[normalized];
  if (words[normalized]) return words[normalized];
  if (locale !== "ar") return normalized.replaceAll("_", " ").replace(/^./, value => value.toUpperCase());
  return normalized.split("_").map(word => words[word] ?? word).filter(Boolean).join(" ");
}
export function stateLabel(value: unknown, locale: string): string {
  const text = String(value ?? "—");
  const states = (locale === "ar" ? ar : en).record.states as Record<string, string>;
  return states[text] ?? text.replaceAll("_", " ");
}
