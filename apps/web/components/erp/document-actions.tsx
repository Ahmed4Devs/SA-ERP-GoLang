import React from "react";
import {
  Send,
  CheckCircle,
  XCircle,
  Lock,
  Printer,
  Download,
  History,
  FileCheck,
  Save,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

export interface DocumentActionsProps {
  status: string;
  onSaveDraft?: () => void;
  onSubmit?: () => void;
  onApprove?: () => void;
  onReject?: () => void;
  onPost?: () => void;
  onZATCASubmit?: () => void;
  onPrint?: () => void;
  onExport?: () => void;
  onOpenAudit?: () => void;
  isArabic?: boolean;
  isLoading?: boolean;
  className?: string;
}

export const DocumentActions: React.FC<DocumentActionsProps> = ({
  status,
  onSaveDraft,
  onSubmit,
  onApprove,
  onReject,
  onPost,
  onZATCASubmit,
  onPrint,
  onExport,
  onOpenAudit,
  isArabic = true,
  isLoading = false,
  className,
}) => {
  const normStatus = (status || "DRAFT").toUpperCase();
  const isDraft = normStatus === "DRAFT";
  const isPending = normStatus === "SUBMITTED" || normStatus === "PENDING_APPROVAL";
  const isApproved = normStatus === "APPROVED";
  const isPosted = normStatus === "POSTED";

  return (
    <div
      className={cn(
        "flex flex-wrap items-center justify-between gap-3 p-4 rounded-xl border bg-card/80 backdrop-blur shadow-xs",
        className
      )}
    >
      {/* Left / Start: Document state progression buttons */}
      <div className="flex flex-wrap items-center gap-2">
        {isDraft && onSaveDraft && (
          <Button
            variant="outline"
            size="sm"
            onClick={onSaveDraft}
            isLoading={isLoading}
            className="gap-1.5"
          >
            <Save className="h-3.5 w-3.5" />
            <span>{isArabic ? "حفظ كمسودة" : "Save Draft"}</span>
          </Button>
        )}

        {isDraft && onSubmit && (
          <Button
            variant="default"
            size="sm"
            onClick={onSubmit}
            isLoading={isLoading}
            className="gap-1.5"
          >
            <Send className="h-3.5 w-3.5 rtl:rotate-180" />
            <span>{isArabic ? "تقديم للاعتماد" : "Submit for Approval"}</span>
          </Button>
        )}

        {isPending && onApprove && (
          <Button
            variant="success"
            size="sm"
            onClick={onApprove}
            isLoading={isLoading}
            className="gap-1.5"
          >
            <CheckCircle className="h-3.5 w-3.5" />
            <span>{isArabic ? "اعتماد المستند" : "Approve"}</span>
          </Button>
        )}

        {isPending && onReject && (
          <Button
            variant="destructive"
            size="sm"
            onClick={onReject}
            isLoading={isLoading}
            className="gap-1.5"
          >
            <XCircle className="h-3.5 w-3.5" />
            <span>{isArabic ? "رفض الطلب" : "Reject"}</span>
          </Button>
        )}

        {isApproved && onPost && (
          <Button
            variant="default"
            size="sm"
            onClick={onPost}
            isLoading={isLoading}
            className="gap-1.5 bg-indigo-600 hover:bg-indigo-700 text-white"
          >
            <Lock className="h-3.5 w-3.5" />
            <span>{isArabic ? "ترحيل للدفتر العام" : "Post to GL"}</span>
          </Button>
        )}

        {isPosted && onZATCASubmit && (
          <Button
            variant="default"
            size="sm"
            onClick={onZATCASubmit}
            isLoading={isLoading}
            className="gap-1.5 bg-teal-600 hover:bg-teal-700 text-white"
          >
            <FileCheck className="h-3.5 w-3.5" />
            <span>{isArabic ? "إرسال لمنصة زاتكا" : "Submit to ZATCA"}</span>
          </Button>
        )}
      </div>

      {/* Right / End: Auxiliary tools (Print, Export, Audit) */}
      <div className="flex items-center gap-2">
        {onPrint && (
          <Button
            variant="ghost"
            size="icon"
            onClick={onPrint}
            title={isArabic ? "طباعة" : "Print"}
          >
            <Printer className="h-4 w-4" />
          </Button>
        )}

        {onExport && (
          <Button
            variant="ghost"
            size="icon"
            onClick={onExport}
            title={isArabic ? "تصدير PDF" : "Export PDF"}
          >
            <Download className="h-4 w-4" />
          </Button>
        )}

        {onOpenAudit && (
          <Button
            variant="outline"
            size="sm"
            onClick={onOpenAudit}
            className="gap-1.5 text-xs text-muted-foreground hover:text-foreground"
          >
            <History className="h-3.5 w-3.5 text-primary" />
            <span>{isArabic ? "سجل التدقيق" : "Audit Trail"}</span>
          </Button>
        )}
      </div>
    </div>
  );
};
