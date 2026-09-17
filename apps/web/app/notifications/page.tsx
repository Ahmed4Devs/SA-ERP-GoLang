import { NotificationsView } from "@/features/notifications/notifications-view";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "مركز التنبيهات والإشعارات | منصة تخطيط الموارد للمنشآت السعودية",
  description: "المنظومة المركزية للتنبيهات الفورية وإشعارات مسارات العمل، موافقات الفوترة الإلكترونية زاتكا، وتحذيرات الميزانيات والمخزون والتراسل البريدي",
};

export default function NotificationsPage() {
  return <NotificationsView />;
}
