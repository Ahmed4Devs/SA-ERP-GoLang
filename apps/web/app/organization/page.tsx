import { ModuleAPIView } from "@/components/modules/module-api-view";
import { moduleResources } from "@/lib/module-resources";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "الهيكل التنظيمي والترقيم | منصة تخطيط الموارد للمنشآت السعودية",
  description: "إدارة الكيانات القانونية، الفروع، الأقسام، مراكز التكلفة، ومحرك الترقيم المتزامن",
};

export default function OrganizationPage() {
  return <ModuleAPIView titleAr="الهيكل التنظيمي" titleEn="Organization" descriptionAr="المنظمات ونطاقات الكيانات الممنوحة للمستخدم." descriptionEn="Live organizations and the current user's entity scopes." resources={moduleResources.organization} />;
}
