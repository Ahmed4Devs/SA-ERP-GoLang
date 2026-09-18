import React from "react";
import { CheckCircle2, Clock, XCircle, ArrowRight, User } from "lucide-react";
import { cn } from "@/lib/utils";

export interface ApprovalStep {
  id: string;
  stepNumber: number;
  stepNameAr: string;
  stepNameEn: string;
  approverName: string;
  approverRole?: string;
  status: "PENDING" | "APPROVED" | "REJECTED" | "SKIPPED";
  decidedAt?: string;
  notes?: string;
}

interface ApprovalTimelineProps {
  steps: ApprovalStep[];
  isArabic?: boolean;
  className?: string;
}

export const ApprovalTimeline: React.FC<ApprovalTimelineProps> = ({
  steps,
  isArabic = true,
  className,
}) => {
  return (
    <div className={cn("space-y-4 rounded-xl border bg-card p-5 shadow-xs", className)}>
      <h4 className="text-xs font-bold uppercase tracking-wider text-muted-foreground">
        {isArabic ? "مسار الاعتماد وسلسلة الموافقات" : "Approval Workflow Progression"}
      </h4>

      <div className="relative">
        <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
          {steps.map((step, idx) => {
            const isLast = idx === steps.length - 1;
            const isApproved = step.status === "APPROVED";
            const isPending = step.status === "PENDING";
            const isRejected = step.status === "REJECTED";

            return (
              <React.Fragment key={step.id}>
                {/* Step Item */}
                <div className="flex items-center sm:flex-col sm:items-center text-start sm:text-center gap-3 sm:gap-2 flex-1 min-w-[130px]">
                  {/* Status Indicator Icon */}
                  <div
                    className={cn(
                      "h-10 w-10 rounded-full flex items-center justify-center border-2 transition-all shadow-xs shrink-0",
                      isApproved
                        ? "bg-emerald-500/10 border-emerald-500 text-emerald-600 dark:text-emerald-400"
                        : isPending
                        ? "bg-amber-500/10 border-amber-500 text-amber-600 dark:text-amber-400 animate-pulse"
                        : isRejected
                        ? "bg-rose-500/10 border-rose-500 text-rose-600 dark:text-rose-400"
                        : "bg-muted border-muted-foreground/30 text-muted-foreground"
                    )}
                  >
                    {isApproved ? (
                      <CheckCircle2 className="h-5 w-5" />
                    ) : isPending ? (
                      <Clock className="h-5 w-5" />
                    ) : isRejected ? (
                      <XCircle className="h-5 w-5" />
                    ) : (
                      <span className="text-xs font-bold">{step.stepNumber}</span>
                    )}
                  </div>

                  {/* Step Info */}
                  <div className="space-y-0.5">
                    <p className="text-xs font-bold text-foreground">
                      {isArabic ? step.stepNameAr : step.stepNameEn}
                    </p>
                    <div className="flex items-center gap-1 text-[11px] text-muted-foreground justify-start sm:justify-center">
                      <User className="h-3 w-3" />
                      <span>{step.approverName}</span>
                    </div>
                    {step.decidedAt && (
                      <p className="text-[10px] text-muted-foreground font-mono">
                        {step.decidedAt}
                      </p>
                    )}
                    {step.notes && (
                      <p className="text-[11px] text-foreground bg-muted/60 px-2 py-0.5 rounded mt-1 italic">
                        &ldquo;{step.notes}&rdquo;
                      </p>
                    )}
                  </div>
                </div>

                {/* Connecting Arrow for desktop */}
                {!isLast && (
                  <div className="hidden sm:flex items-center justify-center text-muted-foreground/40 shrink-0 px-2">
                    <ArrowRight className="h-4 w-4 rtl:rotate-180" />
                  </div>
                )}
              </React.Fragment>
            );
          })}
        </div>
      </div>
    </div>
  );
};
