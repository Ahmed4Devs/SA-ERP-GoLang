import { ZATCAView } from "@/features/zatca/zatca-view";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "الفوترة الإلكترونية مرحلة الربط والتكامل (ZATCA FATOORA Phase 2) | منصة تخطيط الموارد للمنشآت السعودية",
  description: "وحدة الامتثال للفوترة الإلكترونية لهيئة الزكاة والضريبة والجمارك (ZATCA)، إدارة أجهزة الحلول التقنية EGS، الشهادات الرقمية CSID، التشفير والتوقيع الرقمي والتحقق من قواعد BR-KSA",
};

export default function ZATCAPage() {
  return <ZATCAView />;
}
