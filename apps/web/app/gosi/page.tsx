import { GosiView } from "@/features/gosi/gosi-view";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "محرك التأمينات الاجتماعية (GOSI Engine) | منصة تخطيط الموارد للمنشآت السعودية",
  description: "محرك التأمينات الاجتماعية السعودي لاحتساب المعاشات التقاعدية والأخطار المهنية وساند بدقة وفق القواعد الإصدارية وبوابة التأمينات",
};

export default function GosiPage() {
  return <GosiView />;
}
