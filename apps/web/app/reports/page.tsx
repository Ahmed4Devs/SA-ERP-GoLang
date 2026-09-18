import { Metadata } from "next";
import { ModuleAPIView } from "@/components/modules/module-api-view";
import { moduleResources } from "@/lib/module-resources";

export const metadata: Metadata = {
  title: "التقارير والذكاء المالي | ERP-GO",
  description: "منظومة التقارير الشاملة والقوائم المالية وإقرارات الزكاة والضريبة والتحليلات التشغيلية",
};

export default function ReportsPage() {
  return <ModuleAPIView titleAr="التقارير المالية" titleEn="Financial Reports" descriptionAr="تقارير مباشرة من محرك التقارير؛ النتائج الفارغة تبقى فارغة ولا تُستبدل ببيانات تجريبية." descriptionEn="Live reporting output with no fabricated fallback values." resources={moduleResources.reports} />;
}
