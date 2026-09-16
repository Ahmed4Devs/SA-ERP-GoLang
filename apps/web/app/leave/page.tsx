import { LeaveView } from "@/features/leave/leave-view";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "إدارة الإجازات والأرصدة النظامية | منصة تخطيط الموارد للمنشآت السعودية",
  description: "أنواع الإجازات التسع وفق نظام العمل السعودي، تسوية الأرصدة، مسارات الموافقات وسجل حركات الأرصدة غير القابل للتعديل",
};

export default function LeavePage() {
  return <LeaveView />;
}
