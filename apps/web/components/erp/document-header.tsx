import React from "react";
import { Building2, Calendar, User, FileText } from "lucide-react";
import { DocumentStatus, ERPDocumentStatus } from "./document-status";
import { cn } from "@/lib/utils";

export interface DocumentHeaderProps {
  documentType: string;
  documentNumber: string;
  title: string;
  status: ERPDocumentStatus | string;
  legalEntityName?: string;
  branchName?: string;
  documentDate?: string;
  dueDate?: string;
  createdByName?: string;
  isArabic?: boolean;
  className?: string;
  children?: React.ReactNode;
}

export const DocumentHeader: React.FC<DocumentHeaderProps> = ({
  documentType,
  documentNumber,
  title,
  status,
  legalEntityName = "شركة الأفق المتقدمة المحدودة",
  branchName,
  documentDate,
  dueDate,
  createdByName,
  isArabic = true,
  className,
  children,
}) => {
  return (
    <div
      className={cn(
        "rounded-2xl border bg-card p-5 sm:p-6 shadow-xs space-y-4 transition-all",
        className
      )}
    >
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        {/* Document identity */}
        <div className="space-y-1.5">
          <div className="flex items-center gap-2">
            <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-md text-[11px] font-bold bg-primary/10 text-primary uppercase tracking-wide">
              <FileText className="h-3 w-3" />
              {documentType}
            </span>
            <span className="font-mono text-xs font-semibold text-muted-foreground">
              #{documentNumber}
            </span>
          </div>

          <h1 className="text-xl sm:text-2xl font-bold tracking-tight text-foreground">
            {title}
          </h1>
        </div>

        {/* Status & Optional Action Slot */}
        <div className="flex items-center gap-3">
          <DocumentStatus status={status} isArabic={isArabic} />
          {children}
        </div>
      </div>

      {/* Metadata bar */}
      <div className="pt-3 border-t flex flex-wrap items-center gap-4 sm:gap-6 text-xs text-muted-foreground font-medium">
        <div className="flex items-center gap-1.5">
          <Building2 className="h-3.5 w-3.5 text-primary" />
          <span>{legalEntityName}</span>
          {branchName && (
            <>
              <span className="text-border">•</span>
              <span>{branchName}</span>
            </>
          )}
        </div>

        {documentDate && (
          <div className="flex items-center gap-1.5">
            <Calendar className="h-3.5 w-3.5" />
            <span>
              {isArabic ? "تاريخ السند:" : "Doc Date:"} {documentDate}
            </span>
          </div>
        )}

        {dueDate && (
          <div className="flex items-center gap-1.5">
            <Calendar className="h-3.5 w-3.5 text-amber-500" />
            <span>
              {isArabic ? "تاريخ الاستحقاق:" : "Due Date:"} {dueDate}
            </span>
          </div>
        )}

        {createdByName && (
          <div className="flex items-center gap-1.5">
            <User className="h-3.5 w-3.5" />
            <span>
              {isArabic ? "بواسطة:" : "By:"} {createdByName}
            </span>
          </div>
        )}
      </div>
    </div>
  );
};
