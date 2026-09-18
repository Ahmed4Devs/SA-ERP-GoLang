import { ModuleAPIView } from "@/components/modules/module-api-view";
import { moduleResources } from "@/lib/module-resources";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "إدارة الإجازات والأرصدة النظامية | منصة تخطيط الموارد للمنشآت السعودية",
  description: "أنواع الإجازات التسع وفق نظام العمل السعودي، تسوية الأرصدة، مسارات الموافقات وسجل حركات الأرصدة غير القابل للتعديل",
};

export default function LeavePage() {
  return <ModuleAPIView titleAr="الإجازات" titleEn="Leave Management" descriptionAr="أنواع وسياسات وطلبات الإجازات المسجلة فعلياً." descriptionEn="Live leave types, policies, and requests." resources={moduleResources.leave} />;
}
