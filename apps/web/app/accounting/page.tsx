import { ModuleAPIView } from "@/components/modules/module-api-view";
import { moduleResources } from "@/lib/module-resources";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "المحاسبة العامة ودفاتر الأستاذ (General Ledger) | منصة تخطيط الموارد للمنشآت السعودية",
  description: "دليل الحسابات الشجري خماسي المستويات (SOCPA/IFRS)، قيود اليومية المتوازنة وغير القابلة للتعديل، وميزان المراجعة الحسابي",
};

export default function AccountingPage() {
  return <ModuleAPIView titleAr="المحاسبة العامة" titleEn="General Ledger" descriptionAr="عرض الحسابات والقيود والفترات من دفتر الأستاذ الفعلي." descriptionEn="Live accounts, journals, entries, and fiscal periods from the ledger API." resources={moduleResources.accounting} />;
}
