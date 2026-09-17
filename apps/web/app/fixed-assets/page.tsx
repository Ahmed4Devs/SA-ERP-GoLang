import { FixedAssetsView } from "@/features/fixed-assets/fixed-assets-view";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "دورة حياة الأصول الثابتة وإهلاكها (Fixed Assets Lifecycle) | منصة تخطيط الموارد للمنشآت السعودية",
  description: "نظام إدارة دورة حياة الأصول الثابتة، احتساب القسط الثابت آلياً، الترحيل المالي لدفتر الأستاذ العام، المناقلات والمواقع والاستبعاد",
};

export default function FixedAssetsPage() {
  return <FixedAssetsView />;
}
