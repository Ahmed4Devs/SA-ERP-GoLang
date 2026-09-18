import { ModuleAPIView } from "@/components/modules/module-api-view";
import { moduleResources } from "@/lib/module-resources";
import { MFASettings } from "@/features/auth/mfa-settings";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "إعدادات النظام والمنشأة (System Settings) | منصة تخطيط الموارد للمنشآت السعودية",
  description: "إدارة بيانات الكيان التجاري، تكوين الربط مع هيئة الزكاة والضريبة والجمارك (زاتكا)، السياسات المحاسبية والأمان",
};

export default function SettingsPage() {
  return (
    <div className="space-y-6">
      <MFASettings />
      <ModuleAPIView titleAr="إعدادات الحساب والوصول" titleEn="Account & Access Settings" descriptionAr="بيانات الجلسة والصلاحيات والأدوار الفعلية. إعدادات المنصة الحساسة تُدار من بيئة الخادم." descriptionEn="Live session, permissions, and assignments. Sensitive platform configuration is managed server-side." resources={moduleResources.settings} />
    </div>
  );
}
