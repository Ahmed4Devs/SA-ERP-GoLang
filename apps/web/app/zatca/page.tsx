import { ModuleAPIView } from "@/components/modules/module-api-view";
import { moduleResources } from "@/lib/module-resources";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "الفوترة الإلكترونية مرحلة الربط والتكامل (ZATCA FATOORA Phase 2) | منصة تخطيط الموارد للمنشآت السعودية",
  description: "تكامل الفوترة الإلكترونية وإدارة وحدات EGS وشهادات CSID والإرسال؛ يحتاج اعتماداً رسمياً قبل الاستخدام الإنتاجي",
};

export default function ZATCAPage() {
  return <ModuleAPIView titleAr="الفوترة الإلكترونية" titleEn="Electronic Invoicing" descriptionAr="مستندات وإرسالات ووحدات EGS من تكامل النظام الفعلي؛ لا تُعرض شهادات أو نتائج قبول وهمية." descriptionEn="Live documents, submissions, and EGS units; no fabricated certificates or acceptance results." resources={moduleResources.zatca} />;
}
