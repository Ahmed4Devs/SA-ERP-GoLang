import { PurchasingView } from "@/features/purchasing/purchasing-view";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "المشتريات ودورة التعميد والتوريد (P2P) | منصة تخطيط الموارد للمنشآت السعودية",
  description: "إدارة طلبات الشراء الداخلية، استدراج العروض، أوامر التعميد، استلام وفحص البضائع بالمستودع، والمطابقة الثلاثية",
};

export default function PurchasingPage() {
  return <PurchasingView />;
}
