import { WorkflowView } from "@/features/workflow/workflow-view";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "محرك مسارات الموافقة والفصل بين المهام | منصة تخطيط الموارد للمنشآت السعودية",
  description: "محرك سير العمل ومسارات الاعتماد متعددة المستويات مع الفصل الصارم بين المهام والتفويض",
};

export default function WorkflowPage() {
  return <WorkflowView />;
}
