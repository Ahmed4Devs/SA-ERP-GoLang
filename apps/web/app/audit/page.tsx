import { ModuleAPIView } from "@/components/modules/module-api-view";
import { moduleResources } from "@/lib/module-resources";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "سجل التدقيق المالي والأمني | منصة تخطيط الموارد للمنشآت السعودية",
  description: "سجل تاريخي ثابت غير قابل للتعديل ومحمي بسلسلة تجزئة تشفيرية SHA-256",
};

export default function AuditPage() {
  return <ModuleAPIView titleAr="سجل التدقيق" titleEn="Audit Log" descriptionAr="أحداث التدقيق والأمن وسلامة سلسلة التجزئة من المصدر الفعلي." descriptionEn="Live audit, security events, and integrity status." resources={moduleResources.audit} />;
}
