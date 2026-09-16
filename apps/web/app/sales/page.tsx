import { SalesView } from "@/features/sales/sales-view";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "المبيعات ودورة التحصيل وإدارة الطلبات والفواتير | منصة تخطيط الموارد للمنشآت السعودية",
  description: "إدارة دورة المبيعات المتكاملة (Order-to-Cash): عروض الأسعار، أوامر البيع والحد الائتماني، إرساليات البضائع ومخازن الصرف، الفواتير القياسية والمبسطة المتوافقة مع زاتكا، وإشعارات الخصم الدائنة",
};

export default function SalesPage() {
  return <SalesView />;
}
