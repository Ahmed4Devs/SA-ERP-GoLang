import { ModuleAPIView } from "@/components/modules/module-api-view";
import { moduleResources } from "@/lib/module-resources";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "إدارة الحضور والانصراف والورديات | منصة تخطيط الموارد للمنشآت السعودية",
  description: "تسجيل الحضور والانصراف، المناوبات والورديات، احتساب التأخير والانصراف المبكر، وموافقات العمل الإضافي وفق المادة 107 من نظام العمل السعودي",
};

export default function AttendancePage() {
  return <ModuleAPIView titleAr="الحضور والانصراف" titleEn="Attendance" descriptionAr="بيانات الورديات والحضور والعمل الإضافي المسجلة في النظام." descriptionEn="Live shifts, attendance records, and overtime data." resources={moduleResources.attendance} />;
}
