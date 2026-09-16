import { PayablesView } from "@/features/payables/payables-view";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "حسابات الدائنين وفواتير المشتريات وسندات الصرف | منصة تخطيط الموارد للمنشآت السعودية",
  description: "إدارة فواتير الموردين، القيود اليومية للأستاذ العام، سندات الصرف، ومحرك أعمار الديون (AP Aging)",
};

export default function PayablesPage() {
  return <PayablesView />;
}
