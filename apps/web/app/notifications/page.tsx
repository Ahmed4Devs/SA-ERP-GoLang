import { ModuleAPIView } from "@/components/modules/module-api-view";
import { moduleResources } from "@/lib/module-resources";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "مركز التنبيهات والإشعارات | منصة تخطيط الموارد للمنشآت السعودية",
  description: "المنظومة المركزية للتنبيهات الفورية وإشعارات مسارات العمل، موافقات الفوترة الإلكترونية زاتكا، وتحذيرات الميزانيات والمخزون والتراسل البريدي",
};

export default function NotificationsPage() {
  return <ModuleAPIView titleAr="الإشعارات" titleEn="Notifications" descriptionAr="الإشعارات وملخصها وعدّاد غير المقروء من الخدمة الفعلية." descriptionEn="Live notifications, summary, and unread count." resources={moduleResources.notifications} />;
}
