import { ModuleAPIView } from "@/components/modules/module-api-view";
import { moduleResources } from "@/lib/module-resources";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "المخازن والمستودعات وسجل الحركات غير القابل للتعديل | منصة تخطيط الموارد للمنشآت السعودية",
  description: "إدارة المستودعات، أرصدة المخزون اللحظية، سجل حركات المخزون المقفول ضد التعديل، أوامر النقل الداخلي، وحجوزات الأصناف",
};

export default function InventoryPage() {
  return <ModuleAPIView titleAr="المخزون والمستودعات" titleEn="Inventory" descriptionAr="الأرصدة والمستودعات والحركات والجرد الفعلي." descriptionEn="Live balances, warehouses, movements, transfers, and counts." resources={moduleResources.inventory} />;
}
