import { AuditView } from "@/features/audit/audit-view";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "سجل التدقيق المالي والأمني | منصة تخطيط الموارد للمنشآت السعودية",
  description: "سجل تاريخي ثابت غير قابل للتعديل ومحمي بسلسلة تجزئة تشفيرية SHA-256",
};

export default function AuditPage() {
  return <AuditView />;
}
