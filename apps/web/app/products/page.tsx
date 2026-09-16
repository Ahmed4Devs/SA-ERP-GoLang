import { ProductView } from "@/features/products/product-view";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "دليل الأصناف والمنتجات الشامل (ERP Product Master) | منصة تخطيط الموارد للمنشآت السعودية",
  description: "بطاقات الأصناف، وحدات القياس والتحويلات، الفئات الهرمية، المتغيرات، أكواد SKU، ماسح الباركود الفوري، وملفات الضريبة المعتمدة زاتكا",
};

export default function ProductsPage() {
  return <ProductView />;
}
