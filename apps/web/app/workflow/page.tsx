import { ApprovalWorkspace } from "@/features/workflow/approval-workspace";
import { ModuleAPIView } from "@/components/modules/module-api-view";
import { moduleResources } from "@/lib/module-resources";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "محرك مسارات الموافقة والفصل بين المهام | منصة تخطيط الموارد للمنشآت السعودية",
  description: "محرك سير العمل ومسارات الاعتماد متعددة المستويات مع الفصل الصارم بين المهام والتفويض",
};

export default function WorkflowPage() {
  return <div className="space-y-6"><ApprovalWorkspace /><ModuleAPIView titleAr="مسارات الموافقة" titleEn="Workflow" descriptionAr="صندوق الموافقات وتعريفات المسارات والتفويضات الفعلية." descriptionEn="Live approval inbox, definitions, and delegations." resources={moduleResources.workflow.filter(resource => resource.key !== "inbox")} /></div>;
}
