import { OrganizationView } from "@/features/organization/organization-view";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "الهيكل التنظيمي والترقيم | منصة تخطيط الموارد للمنشآت السعودية",
  description: "إدارة الكيانات القانونية، الفروع، الأقسام، مراكز التكلفة، ومحرك الترقيم المتزامن",
};

export default function OrganizationPage() {
  return <OrganizationView />;
}
