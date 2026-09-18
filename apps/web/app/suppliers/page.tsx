import { ModuleAPIView } from "@/components/modules/module-api-view";
import { moduleResources } from "@/lib/module-resources";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "سجل الموردين وبيانات الاتصال والائتمان | منصة تخطيط الموارد للمنشآت السعودية",
  description: "إدارة الموردين المحليين والدوليين، التحقق من السجل التجاري والرقم الضريبي، الآيبان وشروط الدفع وحسابات الذمم الدائنة",
};

export default function SuppliersPage() {
  return <ModuleAPIView titleAr="الموردون" titleEn="Suppliers" descriptionAr="الموردون وشروط الدفع الفعلية للكيان المختار." descriptionEn="Live suppliers and payment terms for the selected entity." resources={moduleResources.suppliers} />;
}
