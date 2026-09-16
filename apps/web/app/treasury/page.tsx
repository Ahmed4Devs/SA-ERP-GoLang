import { TreasuryView } from "@/features/treasury/treasury-view";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "الخزينة وإدارة السيولة النقدية | منصة تخطيط الموارد للمنشآت السعودية",
  description: "إدارة صناديق النقدية، العهد النثرية، سندات القبض والصرف، والتحويلات المالية الداخلية",
};

export default function TreasuryPage() {
  return <TreasuryView />;
}
