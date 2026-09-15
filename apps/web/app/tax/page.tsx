import { TaxView } from "@/features/tax/tax-view";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "إدارة الضرائب وضريبة القيمة المضافة (VAT & ZATCA Engine) | منصة تخطيط الموارد للمنشآت السعودية",
  description: "محرك ضريبة القيمة المضافة السعودي المؤرخ بالسريان الزمني، حاسبة الضريبة الدقيقة، وإقرار الـ 16 حقلاً المعتمد لدى هيئة الزكاة والضريبة والجمارك",
};

export default function TaxPage() {
  return <TaxView />;
}
