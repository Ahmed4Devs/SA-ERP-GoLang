import { PayrollView } from "@/features/payroll/payroll-view";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "محرك الرواتب والأجور ونظام حماية الأجور (WPS) | منصة تخطيط الموارد للمنشآت السعودية",
  description: "محرك رواتب محكم وغير قابل للتعديل بعد الترحيل، متوافق 100% مع نظام حماية الأجور السعودي (WPS / MHRSD / Mudad)، والتأمينات الاجتماعية GOSI ومكافأة نهاية الخدمة",
};

export default function PayrollPage() {
  return <PayrollView />;
}
