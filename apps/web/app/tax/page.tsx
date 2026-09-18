import { ModuleAPIView } from "@/components/modules/module-api-view";
import { moduleResources } from "@/lib/module-resources";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "إدارة الضرائب وضريبة القيمة المضافة (VAT & ZATCA Engine) | منصة تخطيط الموارد للمنشآت السعودية",
  description: "محرك ضريبة القيمة المضافة السعودي المؤرخ بالسريان الزمني، حاسبة الضريبة الدقيقة، وإقرار الـ 16 حقلاً المعتمد لدى هيئة الزكاة والضريبة والجمارك",
};

export default function TaxPage() {
  return <ModuleAPIView titleAr="الضرائب وضريبة القيمة المضافة" titleEn="Tax & VAT" descriptionAr="الفئات والأكواد والحركات والفترات والإقرارات الضريبية الفعلية." descriptionEn="Live tax categories, codes, transactions, periods, and returns." resources={moduleResources.tax} />;
}
