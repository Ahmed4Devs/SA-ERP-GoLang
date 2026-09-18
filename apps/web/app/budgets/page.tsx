import { ModuleAPIView } from "@/components/modules/module-api-view";
import { moduleResources } from "@/lib/module-resources";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "الميزانيات التقديرية والرقابة المالية (Budgeting & Control) | منصة تخطيط الموارد للمنشآت السعودية",
  description: "نظام إدارة الميزانيات التقديرية متعدد الأبعاد، الرقابة على المصروفات، المناقلات وإعادة التخصيص، ومراقبة الانحرافات اللحظية",
};

export default function BudgetsPage() {
  return <ModuleAPIView titleAr="الميزانيات التقديرية" titleEn="Budgeting" descriptionAr="الميزانيات والمناقلات وملخص الرقابة من قاعدة البيانات." descriptionEn="Live budgets, transfers, and control summary." resources={moduleResources.budgeting} />;
}
