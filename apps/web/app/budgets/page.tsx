import { BudgetingView } from "@/features/budgeting/budgeting-view";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "الميزانيات التقديرية والرقابة المالية (Budgeting & Control) | منصة تخطيط الموارد للمنشآت السعودية",
  description: "نظام إدارة الميزانيات التقديرية متعدد الأبعاد، الرقابة على المصروفات، المناقلات وإعادة التخصيص، ومراقبة الانحرافات اللحظية",
};

export default function BudgetsPage() {
  return <BudgetingView />;
}
