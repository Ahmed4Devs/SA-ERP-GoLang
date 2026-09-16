import { SuppliersView } from "@/features/suppliers/suppliers-view";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "سجل الموردين وبيانات الاتصال والائتمان | منصة تخطيط الموارد للمنشآت السعودية",
  description: "إدارة الموردين المحليين والدوليين، التحقق من السجل التجاري والرقم الضريبي، الآيبان وشروط الدفع وحسابات الذمم الدائنة",
};

export default function SuppliersPage() {
  return <SuppliersView />;
}
