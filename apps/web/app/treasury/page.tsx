import { ModuleAPIView } from "@/components/modules/module-api-view";
import { moduleResources } from "@/lib/module-resources";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "الخزينة وإدارة السيولة النقدية | منصة تخطيط الموارد للمنشآت السعودية",
  description: "إدارة صناديق النقدية، العهد النثرية، سندات القبض والصرف، والتحويلات المالية الداخلية",
};

export default function TreasuryPage() {
  return <ModuleAPIView titleAr="الخزينة والسيولة" titleEn="Treasury" descriptionAr="حسابات النقد والمقبوضات والمدفوعات والتحويلات الفعلية." descriptionEn="Live cash accounts, receipts, payments, and transfers." resources={moduleResources.treasury} />;
}
