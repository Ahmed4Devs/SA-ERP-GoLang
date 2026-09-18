import { ModuleAPIView } from "@/components/modules/module-api-view";
import { moduleResources } from "@/lib/module-resources";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "إدارة الموارد البشرية والعقود | منصة تخطيط الموارد للمنشآت السعودية",
  description: "ملف الموظفين الموحد، عقود قوى وتفكيك الرواتب، حماية الخصوصية وحجب الهويات والآيبان، ومحرك التوطين ونطاقات",
};

export default function HRPage() {
  return <ModuleAPIView titleAr="الموارد البشرية" titleEn="Human Resources" descriptionAr="الموظفون والوظائف والعقود وبيانات التوطين الفعلية." descriptionEn="Live employees, jobs, contracts, and Saudization metrics." resources={moduleResources.hr} />;
}
