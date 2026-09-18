import { ModuleAPIView } from "@/components/modules/module-api-view";
import { moduleResources } from "@/lib/module-resources";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "إدارة وتخزين الوثائق والمستندات | منصة تخطيط الموارد للمنشآت السعودية",
  description: "المنصة المركزية لتخزين وإدارة وثائق ومستندات المنشأة، أرشفة الإصدارات بالبصمة الرقمية SHA-256، والربط الشامل بمختلف العمليات المحاسبية والتجارية",
};

export default function DocumentsPage() {
  return <ModuleAPIView titleAr="المستندات والوثائق" titleEn="Documents" descriptionAr="الوثائق الفعلية وملخص التخزين للكيان المختار." descriptionEn="Live documents and storage summary for the selected entity." resources={moduleResources.documents} />;
}
