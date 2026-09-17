import { Metadata } from "next";
import { ReportsView } from "@/features/reports/reports-view";

export const metadata: Metadata = {
  title: "التقارير والذكاء المالي | ERP-GO",
  description: "منظومة التقارير الشاملة والقوائم المالية وإقرارات الزكاة والضريبة والتحليلات التشغيلية",
};

export default function ReportsPage() {
  return <ReportsView />;
}
