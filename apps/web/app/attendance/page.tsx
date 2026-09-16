import { AttendanceView } from "@/features/attendance/attendance-view";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "إدارة الحضور والانصراف والورديات | منصة تخطيط الموارد للمنشآت السعودية",
  description: "تسجيل الحضور والانصراف، المناوبات والورديات، احتساب التأخير والانصراف المبكر، وموافقات العمل الإضافي وفق المادة 107 من نظام العمل السعودي",
};

export default function AttendancePage() {
  return <AttendanceView />;
}
