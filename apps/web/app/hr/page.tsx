import { HRView } from "@/features/hr/hr-view";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "إدارة الموارد البشرية والعقود | منصة تخطيط الموارد للمنشآت السعودية",
  description: "ملف الموظفين الموحد، عقود قوى وتفكيك الرواتب، حماية الخصوصية وحجب الهويات والآيبان، ومحرك التوطين ونطاقات",
};

export default function HRPage() {
  return <HRView />;
}
