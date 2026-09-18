import { ModuleAPIView } from "@/components/modules/module-api-view";
import { moduleResources } from "@/lib/module-resources";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "محرك الرواتب والأجور ونظام حماية الأجور (WPS) | منصة تخطيط الموارد للمنشآت السعودية",
  description: "محرك رواتب وفترات ومسيرات وقسائم مع دعم إنشاء ملف WPS؛ يتطلب التحقق القانوني والتشغيلي قبل الاستخدام الفعلي",
};

export default function PayrollPage() {
  return <ModuleAPIView titleAr="الرواتب والأجور" titleEn="Payroll" descriptionAr="مسيرات وفترات ومكونات وقسائم الرواتب الفعلية." descriptionEn="Live payroll runs, periods, components, structures, payslips, and loans." resources={moduleResources.payroll} />;
}
