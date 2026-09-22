"use client";

import React from "react";
import { X, ShieldCheck, History, User, Hash, Clock } from "lucide-react";
import { cn } from "@/lib/utils";

export interface AuditRecord {
  id: string;
  action: string;
  actorName: string;
  occurredAt: string;
  previousHash?: string;
  recordHash: string;
  changes?: Record<string, { before: any; after: any }>;
}

interface AuditDrawerProps {
  open: boolean;
  onClose: () => void;
  documentTitle: string;
  records: AuditRecord[];
  isArabic?: boolean;
}

export const AuditDrawer: React.FC<AuditDrawerProps> = ({
  open,
  onClose,
  documentTitle,
  records,
  isArabic = true,
}) => {
  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 overflow-hidden">
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-background/80 backdrop-blur-sm transition-opacity"
        onClick={onClose}
      />

      {/* Slide-over panel */}
      <div className="fixed inset-y-0 inset-inline-end-0 flex max-w-full">
        <div className="w-screen max-w-md bg-card border-s shadow-2xl flex flex-col justify-between">
          {/* Header */}
          <div className="p-5 border-b flex items-center justify-between">
            <div className="space-y-1">
              <div className="inline-flex items-center gap-1.5 text-xs font-bold text-emerald-600 dark:text-emerald-400">
                <ShieldCheck className="h-4 w-4" />
                <span>{isArabic ? "سلسلة التدقيق المشفرة (SHA-256)" : "Cryptographically Chained Trail"}</span>
              </div>
              <h3 className="font-bold text-sm text-foreground">{documentTitle}</h3>
            </div>
            <button
              onClick={onClose}
              className="p-1.5 rounded-lg text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
            >
              <X className="h-4 w-4" />
            </button>
          </div>

          {/* Audit events timeline */}
          <div className="flex-1 overflow-y-auto p-5 space-y-4">
            {records.length === 0 ? (
              <div className="text-center py-12 text-xs text-muted-foreground">
                <History className="h-8 w-8 text-muted-foreground/40 mx-auto mb-2" />
                <span>{isArabic ? "لا توجد سجلات تدقيق مسجلة بعد" : "No audit entries recorded yet"}</span>
              </div>
            ) : (
              records.map((rec, i) => (
                <div
                  key={rec.id || i}
                  className="p-3.5 rounded-xl border bg-muted/20 space-y-2 text-xs hover:border-primary/40 transition-colors"
                >
                  <div className="flex items-center justify-between">
                    <span className="px-2 py-0.5 rounded font-mono font-bold text-[10px] bg-primary/10 text-primary">
                      {rec.action}
                    </span>
                    <span className="text-[11px] text-muted-foreground flex items-center gap-1 font-mono">
                      <Clock className="h-3 w-3" />
                      {rec.occurredAt}
                    </span>
                  </div>

                  <div className="flex items-center gap-1.5 text-muted-foreground">
                    <User className="h-3.5 w-3.5" />
                    <span className="font-medium text-foreground">{rec.actorName}</span>
                  </div>

                  {/* Hash verification badge */}
                  <div className="pt-2 border-t flex items-center gap-1 font-mono text-[10px] text-muted-foreground">
                    <Hash className="h-3 w-3 text-emerald-500" />
                    <span className="truncate">{rec.recordHash}</span>
                  </div>
                </div>
              ))
            )}
          </div>

          {/* Footer note */}
          <div className="p-4 border-t bg-muted/30 text-[11px] text-muted-foreground text-center">
            {isArabic
              ? "سجل تدقيق مترابط ببصمات تشفيرية وقابل لكشف العبث"
              : "Append-only audit trail with cryptographic tamper evidence"}
          </div>
        </div>
      </div>
    </div>
  );
};
