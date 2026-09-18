import React from "react";
import {
  FileEdit,
  Clock,
  CheckCircle2,
  Lock,
  XCircle,
  AlertTriangle,
  Ban,
  ShieldCheck,
} from "lucide-react";
import { cn } from "@/lib/utils";

export type ERPDocumentStatus =
  | "DRAFT"
  | "SUBMITTED"
  | "PENDING_APPROVAL"
  | "APPROVED"
  | "POSTED"
  | "CLEARED"
  | "REJECTED"
  | "CANCELLED";

interface DocumentStatusProps {
  status: ERPDocumentStatus | string;
  className?: string;
  isArabic?: boolean;
}

const statusConfig: Record<
  string,
  {
    labelAr: string;
    labelEn: string;
    icon: React.ElementType;
    badgeClass: string;
  }
> = {
  DRAFT: {
    labelAr: "مسودة",
    labelEn: "Draft",
    icon: FileEdit,
    badgeClass: "bg-muted text-muted-foreground border-border/80",
  },
  SUBMITTED: {
    labelAr: "مقدم للاعتماد",
    labelEn: "Submitted",
    icon: Clock,
    badgeClass: "bg-blue-500/10 text-blue-600 dark:text-blue-400 border-blue-500/20",
  },
  PENDING_APPROVAL: {
    labelAr: "قيد الموافقة",
    labelEn: "Pending Approval",
    icon: Clock,
    badgeClass: "bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/20",
  },
  APPROVED: {
    labelAr: "معتمد إدارياً",
    labelEn: "Approved",
    icon: CheckCircle2,
    badgeClass: "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20",
  },
  POSTED: {
    labelAr: "مرحل للدفتر العام",
    labelEn: "Posted & Locked",
    icon: Lock,
    badgeClass: "bg-indigo-500/10 text-indigo-600 dark:text-indigo-400 border-indigo-500/20 font-bold",
  },
  CLEARED: {
    labelAr: "مخلص زاتكا",
    labelEn: "ZATCA Cleared",
    icon: ShieldCheck,
    badgeClass: "bg-teal-500/10 text-teal-600 dark:text-teal-400 border-teal-500/20 font-bold",
  },
  REJECTED: {
    labelAr: "مرفوض",
    labelEn: "Rejected",
    icon: XCircle,
    badgeClass: "bg-rose-500/10 text-rose-600 dark:text-rose-400 border-rose-500/20",
  },
  CANCELLED: {
    labelAr: "ملغي",
    labelEn: "Cancelled",
    icon: Ban,
    badgeClass: "bg-zinc-500/10 text-zinc-500 border-zinc-500/20 line-through",
  },
};

export const DocumentStatus: React.FC<DocumentStatusProps> = ({
  status,
  className,
  isArabic = true,
}) => {
  const normStatus = (status || "DRAFT").toUpperCase();
  const cfg = statusConfig[normStatus] || {
    labelAr: normStatus,
    labelEn: normStatus,
    icon: AlertTriangle,
    badgeClass: "bg-muted text-muted-foreground border-border",
  };

  const Icon = cfg.icon;

  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold border shadow-xs transition-colors",
        cfg.badgeClass,
        className
      )}
    >
      <Icon className="h-3.5 w-3.5 shrink-0" />
      <span>{isArabic ? cfg.labelAr : cfg.labelEn}</span>
    </span>
  );
};
