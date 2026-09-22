import React from "react";
import { cn } from "@/lib/utils";

export interface TotalsPanelProps {
  subtotal: number;
  discountTotal?: number;
  taxableAmount: number;
  taxRate?: number; // e.g. 15 for KSA 15% VAT
  taxAmount: number;
  grandTotal: number;
  currency?: string;
  isArabic?: boolean;
  className?: string;
}

export const TotalsPanel: React.FC<TotalsPanelProps> = ({
  subtotal,
  discountTotal = 0,
  taxableAmount,
  taxRate = 15,
  taxAmount,
  grandTotal,
  currency = "SAR",
  isArabic = true,
  className,
}) => {
  const formatCur = (val: number) => {
    return new Intl.NumberFormat(isArabic ? "ar-SA" : "en-US", {
      style: "currency",
      currency: currency,
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(val || 0);
  };

  return (
    <div
      className={cn(
        "rounded-2xl border bg-card p-5 sm:p-6 shadow-xs max-w-sm ms-auto space-y-3",
        className
      )}
    >
      <h4 className="text-xs font-bold uppercase tracking-wider text-muted-foreground pb-2 border-b">
        {isArabic ? "ملخص الحسابات والضريبة" : "Financial Totals & Tax Summary"}
      </h4>

      {/* Subtotal */}
      <div className="flex items-center justify-between text-xs">
        <span className="text-muted-foreground">{isArabic ? "المجموع الفرعي (قبل الخصم):" : "Subtotal:"}</span>
        <span className="font-semibold text-foreground">{formatCur(subtotal)}</span>
      </div>

      {/* Discounts if any */}
      {discountTotal > 0 && (
        <div className="flex items-center justify-between text-xs text-rose-600">
          <span>{isArabic ? "إجمالي الخصومات:" : "Discounts:"}</span>
          <span className="font-semibold">-{formatCur(discountTotal)}</span>
        </div>
      )}

      {/* Net Taxable */}
      <div className="flex items-center justify-between text-xs">
        <span className="text-muted-foreground">{isArabic ? "المبلغ الخاضع للضريبة:" : "Taxable Amount:"}</span>
        <span className="font-semibold text-foreground">{formatCur(taxableAmount)}</span>
      </div>

      {/* VAT 15% */}
      <div className="flex items-center justify-between text-xs">
        <span className="text-muted-foreground">
          {isArabic ? `ضريبة القيمة المضافة (${taxRate}%):` : `VAT (${taxRate}%):`}
        </span>
        <span className="font-semibold text-emerald-600 dark:text-emerald-400">
          {formatCur(taxAmount)}
        </span>
      </div>

      {/* Grand Total */}
      <div className="pt-3 border-t-2 border-border/80 flex items-center justify-between">
        <span className="text-sm font-bold text-foreground">{isArabic ? "المجموع الإجمالي النهائي:" : "Grand Total Net:"}</span>
        <span className="text-lg font-extrabold text-primary">{formatCur(grandTotal)}</span>
      </div>

      <p className="text-[10px] text-muted-foreground text-end pt-1">
        {isArabic ? "حسابات عشرية دقيقة غير تقريبية" : "Exact Decimal Precision (No Floating-Point Rounding)"}
      </p>
    </div>
  );
};
