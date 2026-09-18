import { ModuleAPIView } from "@/components/modules/module-api-view";
import { moduleResources } from "@/lib/module-resources";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "محرك التأمينات الاجتماعية (GOSI Engine) | منصة تخطيط الموارد للمنشآت السعودية",
  description: "محرك التأمينات الاجتماعية السعودي لاحتساب المعاشات التقاعدية والأخطار المهنية وساند بدقة وفق القواعد الإصدارية وبوابة التأمينات",
};

export default function GosiPage() {
  return <ModuleAPIView titleAr="التأمينات الاجتماعية" titleEn="GOSI" descriptionAr="قواعد التأمينات والمشتركون والاحتساب الفعلي." descriptionEn="Live GOSI rules, profiles, and calculations." resources={moduleResources.gosi} />;
}
