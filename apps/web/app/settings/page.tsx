import { SettingsView } from "@/features/settings/settings-view";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "إعدادات النظام والمنشأة (System Settings) | منصة تخطيط الموارد للمنشآت السعودية",
  description: "إدارة بيانات الكيان التجاري، تكوين الربط مع هيئة الزكاة والضريبة والجمارك (زاتكا)، السياسات المحاسبية والأمان",
};

export default function SettingsPage() {
  return <SettingsView />;
}
