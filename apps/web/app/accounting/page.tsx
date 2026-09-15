import { AccountingView } from "@/features/accounting/accounting-view";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "المحاسبة العامة ودفاتر الأستاذ (General Ledger) | منصة تخطيط الموارد للمنشآت السعودية",
  description: "دليل الحسابات الشجري خماسي المستويات (SOCPA/IFRS)، قيود اليومية المتوازنة وغير القابلة للتعديل، وميزان المراجعة الحسابي",
};

export default function AccountingPage() {
  return <AccountingView />;
}
