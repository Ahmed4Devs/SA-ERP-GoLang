import { ModuleAPIView } from "@/components/modules/module-api-view";
import { moduleResources } from "@/lib/module-resources";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "دليل الأصناف والمنتجات الشامل (ERP Product Master) | منصة تخطيط الموارد للمنشآت السعودية",
  description: "بطاقات الأصناف، وحدات القياس والتحويلات، الفئات الهرمية، المتغيرات، أكواد SKU، ماسح الباركود الفوري، وملفات الضريبة المعتمدة زاتكا",
};

export default function ProductsPage() {
  return <ModuleAPIView titleAr="الأصناف والمنتجات" titleEn="Products" descriptionAr="الأصناف والفئات ووحدات القياس المسجلة في النظام." descriptionEn="Live items, categories, and units of measure." resources={moduleResources.products} />;
}
