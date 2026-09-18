import React from "react";
import { ShieldCheck, AlertTriangle, XCircle, Clock, QrCode } from "lucide-react";
import { cn } from "@/lib/utils";

export type ZATCASubmissionStatus =
  | "CLEARED"
  | "REPORTED"
  | "QUEUED"
  | "SUBMITTING"
  | "WARNING"
  | "REJECTED"
  | "NOT_GENERATED";

interface ZATCAStatusBadgeProps {
  status: ZATCASubmissionStatus | string;
  hasQRCode?: boolean;
  hasECDSA?: boolean;
  isArabic?: boolean;
  className?: string;
}

export const ZATCAStatusBadge: React.FC<ZATCAStatusBadgeProps> = ({
  status,
  hasQRCode = true,
  hasECDSA = true,
  isArabic = true,
  className,
}) => {
  const norm = (status || "QUEUED").toUpperCase();

  const isCleared = norm === "CLEARED" || norm === "REPORTED";
  const isWarning = norm === "WARNING";
  const isRejected = norm === "REJECTED" || norm === "FAILED";

  return (
    <div className={cn("inline-flex items-center gap-2", className)}>
      <span
        className={cn(
          "inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-bold border shadow-xs transition-colors",
          isCleared
            ? "bg-teal-500/10 text-teal-600 dark:text-teal-400 border-teal-500/20"
            : isWarning
            ? "bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/20"
            : isRejected
            ? "bg-rose-500/10 text-rose-600 dark:text-rose-400 border-rose-500/20"
            : "bg-blue-500/10 text-blue-600 dark:text-blue-400 border-blue-500/20"
        )}
      >
        {isCleared ? (
          <ShieldCheck className="h-3.5 w-3.5" />
        ) : isWarning ? (
          <AlertTriangle className="h-3.5 w-3.5" />
        ) : isRejected ? (
          <XCircle className="h-3.5 w-3.5" />
        ) : (
          <Clock className="h-3.5 w-3.5" />
        )}
        <span>
          {isCleared
            ? isArabic
              ? "معتمدة ومخلصة (زاتكا)"
              : "ZATCA Cleared"
            : isWarning
            ? isArabic
              ? "مقبولة مع تحذيرات"
              : "Accepted with Warnings"
            : isRejected
            ? isArabic
              ? "مرفوضة من الهيئة"
              : "ZATCA Rejected"
            : isArabic
            ? "قيد الإرسال للربط"
            : "Queued for Clearance"}
        </span>
      </span>

      {/* Security pills */}
      <div className="hidden sm:inline-flex items-center gap-1">
        {hasECDSA && (
          <span className="px-2 py-0.5 rounded text-[10px] font-mono font-semibold bg-muted text-muted-foreground border">
            ECDSA-SHA256
          </span>
        )}
        {hasQRCode && (
          <span className="p-1 rounded bg-muted text-muted-foreground border" title="ZATCA TLV Base64 QR Ready">
            <QrCode className="h-3 w-3" />
          </span>
        )}
      </div>
    </div>
  );
};
