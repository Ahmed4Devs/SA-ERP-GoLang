import { ModuleAPIView } from "@/components/modules/module-api-view";
import { moduleResources } from "@/lib/module-resources";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "إدارة البنوك والمطابقة المصرفية | منصة تخطيط الموارد للمنشآت السعودية",
  description: "إدارة الحسابات المصرفية الرسمية، استيراد كشوف الحسابات الإلكترونية، ومحرك المطابقة والتسوية الآلية",
};

export default function BankingPage() {
  return <ModuleAPIView titleAr="البنوك والمطابقة" titleEn="Banking" descriptionAr="الحسابات البنكية والكشوف والمطابقات الفعلية." descriptionEn="Live bank accounts, statements, and reconciliations." resources={moduleResources.banking} />;
}
