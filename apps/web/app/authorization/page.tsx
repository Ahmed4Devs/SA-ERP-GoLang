import { RolesView } from "@/features/authorization/roles-view";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "إدارة الصلاحيات والأدوار | منصة تخطيط الموارد للمنشآت السعودية",
  description: "إدارة الأدوار، الصلاحيات الدقيقة، وتقييد الوصول بنطاق الكيان القانوني وكاش السياسات",
};

export default function AuthorizationPage() {
  return <RolesView />;
}
