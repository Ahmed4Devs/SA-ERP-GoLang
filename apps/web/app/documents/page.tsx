import { DocumentLibrary } from "@/features/documents/document-library";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "إدارة وتخزين الوثائق والمستندات | منصة تخطيط الموارد للمنشآت السعودية",
  description: "المنصة المركزية لتخزين وإدارة وثائق ومستندات المنشأة، أرشفة الإصدارات بالبصمة الرقمية SHA-256، والربط الشامل بمختلف العمليات المحاسبية والتجارية",
};

export default function DocumentsPage() {
  return <DocumentLibrary />;
}
