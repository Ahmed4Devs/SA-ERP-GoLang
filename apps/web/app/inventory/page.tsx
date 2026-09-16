import { InventoryView } from "@/features/inventory/inventory-view";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "المخازن والمستودعات وسجل الحركات غير القابل للتعديل | منصة تخطيط الموارد للمنشآت السعودية",
  description: "إدارة المستودعات، أرصدة المخزون اللحظية، سجل حركات المخزون المقفول ضد التعديل، أوامر النقل الداخلي، وحجوزات الأصناف",
};

export default function InventoryPage() {
  return <InventoryView />;
}
