import { ModuleAPIView } from "@/components/modules/module-api-view";
import { moduleResources } from "@/lib/module-resources";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "المشتريات ودورة التعميد والتوريد (P2P) | منصة تخطيط الموارد للمنشآت السعودية",
  description: "إدارة طلبات الشراء الداخلية، استدراج العروض، أوامر التعميد، استلام وفحص البضائع بالمستودع، والمطابقة الثلاثية",
};

export default function PurchasingPage() {
  return <ModuleAPIView titleAr="المشتريات" titleEn="Purchasing" descriptionAr="طلبات وعروض وأوامر واستلامات الشراء الفعلية." descriptionEn="Live purchase requests, RFQs, quotes, orders, and receipts." resources={moduleResources.purchasing} />;
}
