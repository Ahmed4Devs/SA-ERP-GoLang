import { ModuleAPIView } from "@/components/modules/module-api-view";
import { moduleResources } from "@/lib/module-resources";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "العملاء والمقبوضات وحسابات المدينين | منصة تخطيط الموارد للمنشآت السعودية",
  description: "سجل العملاء المتوافق مع هيئة الزكاة والضريبة والجمارك، سندات القبض وتوزيعها، متابعة الأرصدة المفتوحة، شروط الدفع والائتمان، وتحليل أعمار الديون",
};

export default function CustomersPage() {
  return <ModuleAPIView titleAr="العملاء والمدينون" titleEn="Customers & Receivables" descriptionAr="العملاء وشروط الدفع والمقبوضات الفعلية." descriptionEn="Live customers, payment terms, and receipts." resources={moduleResources.customers} />;
}
