import { ModuleAPIView } from "@/components/modules/module-api-view";
import { moduleResources } from "@/lib/module-resources";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "المبيعات ودورة التحصيل وإدارة الطلبات والفواتير | منصة تخطيط الموارد للمنشآت السعودية",
  description: "إدارة دورة المبيعات المتكاملة (Order-to-Cash): عروض الأسعار، أوامر البيع والحد الائتماني، إرساليات البضائع ومخازن الصرف، مسارات الفواتير القياسية والمبسطة وتكامل زاتكا، وإشعارات الخصم الدائنة",
};

export default function SalesPage() {
  return <ModuleAPIView titleAr="المبيعات والفوترة" titleEn="Sales" descriptionAr="الفواتير والأوامر والعروض والتسليمات الفعلية." descriptionEn="Live invoices, orders, quotations, deliveries, and credit notes." resources={moduleResources.sales} />;
}
