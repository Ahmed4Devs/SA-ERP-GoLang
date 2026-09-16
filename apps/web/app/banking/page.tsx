import { BankingView } from "@/features/banking/banking-view";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "إدارة البنوك والمطابقة المصرفية | منصة تخطيط الموارد للمنشآت السعودية",
  description: "إدارة الحسابات المصرفية الرسمية، استيراد كشوف الحسابات الإلكترونية، ومحرك المطابقة والتسوية الآلية",
};

export default function BankingPage() {
  return <BankingView />;
}
