import { CustomersView } from "@/features/customers/customers-view";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "العملاء والمقبوضات وحسابات المدينين | منصة تخطيط الموارد للمنشآت السعودية",
  description: "سجل العملاء المتوافق مع هيئة الزكاة والضريبة والجمارك، سندات القبض وتوزيعها، متابعة الأرصدة المفتوحة، شروط الدفع والائتمان، وتحليل أعمار الديون",
};

export default function CustomersPage() {
  return <CustomersView />;
}
