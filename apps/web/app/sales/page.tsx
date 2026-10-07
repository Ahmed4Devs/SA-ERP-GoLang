import { ModuleAPIView } from "@/components/modules/module-api-view";
import { moduleResources } from "@/lib/module-resources";
import type { Metadata } from "next";
import { InvoiceWorkspace } from "@/features/sales/invoice-workspace";

export const metadata: Metadata = {
  title: "المبيعات ودورة التحصيل وإدارة الطلبات والفواتير | منصة تخطيط الموارد للمنشآت السعودية",
  description: "إدارة دورة المبيعات المتكاملة (Order-to-Cash): عروض الأسعار، أوامر البيع والحد الائتماني، إرساليات البضائع ومخازن الصرف، مسارات الفواتير القياسية والمبسطة وتكامل زاتكا، وإشعارات الخصم الدائنة",
};

export default function SalesPage() {
  return <div className="space-y-8"><InvoiceWorkspace /><ModuleAPIView titleAr="مستندات المبيعات" titleEn="Sales documents" descriptionAr="الأوامر والعروض والتسليمات والإشعارات الدائنة." descriptionEn="Orders, quotations, deliveries, and credit notes." resources={moduleResources.sales.filter((resource) => resource.key !== "invoices")} /></div>;
}
