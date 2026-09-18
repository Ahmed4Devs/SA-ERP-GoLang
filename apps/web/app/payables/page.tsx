import { ModuleAPIView } from "@/components/modules/module-api-view";
import { moduleResources } from "@/lib/module-resources";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "حسابات الدائنين وفواتير المشتريات وسندات الصرف | منصة تخطيط الموارد للمنشآت السعودية",
  description: "إدارة فواتير الموردين، القيود اليومية للأستاذ العام، سندات الصرف، ومحرك أعمار الديون (AP Aging)",
};

export default function PayablesPage() {
  return <ModuleAPIView titleAr="حسابات الدائنين" titleEn="Payables" descriptionAr="فواتير الموردين والمدفوعات والإشعارات الدائنة الفعلية." descriptionEn="Live supplier invoices, payments, and credit notes." resources={moduleResources.payables} />;
}
