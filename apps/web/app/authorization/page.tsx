import { ModuleAPIView } from "@/components/modules/module-api-view";
import { moduleResources } from "@/lib/module-resources";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "إدارة الصلاحيات والأدوار | منصة تخطيط الموارد للمنشآت السعودية",
  description: "إدارة الأدوار، الصلاحيات الدقيقة، وتقييد الوصول بنطاق الكيان القانوني وكاش السياسات",
};

export default function AuthorizationPage() {
  return <ModuleAPIView titleAr="الصلاحيات والأدوار" titleEn="Authorization" descriptionAr="الأدوار والصلاحيات والتكليفات التي يفرضها الخادم." descriptionEn="Server-enforced roles, permissions, and assignments." resources={moduleResources.authorization} />;
}
