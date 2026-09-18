import { ModuleAPIView } from "@/components/modules/module-api-view";
import { moduleResources } from "@/lib/module-resources";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "دورة حياة الأصول الثابتة وإهلاكها (Fixed Assets Lifecycle) | منصة تخطيط الموارد للمنشآت السعودية",
  description: "نظام إدارة دورة حياة الأصول الثابتة، احتساب القسط الثابت آلياً، الترحيل المالي لدفتر الأستاذ العام، المناقلات والمواقع والاستبعاد",
};

export default function FixedAssetsPage() {
  return <ModuleAPIView titleAr="الأصول الثابتة" titleEn="Fixed Assets" descriptionAr="الأصول والفئات والمواقع والإهلاك والمناقلات الفعلية." descriptionEn="Live assets, categories, locations, depreciation, and transfers." resources={moduleResources.fixedassets} />;
}
