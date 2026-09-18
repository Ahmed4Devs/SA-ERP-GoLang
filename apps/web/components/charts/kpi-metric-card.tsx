import React from "react";
import { ArrowUpRight, ArrowDownRight, Minus } from "lucide-react";
import { cn } from "@/lib/utils";

export interface KPIMetricCardProps {
  title: string;
  value: string | number;
  change?: string;
  changeType?: "positive" | "negative" | "neutral";
  changeNote?: string;
  icon: React.ElementType;
  progressPercent?: number;
  accentColor?: "emerald" | "blue" | "indigo" | "amber" | "rose" | "teal";
  className?: string;
}

const colorMap = {
  emerald: {
    bg: "bg-emerald-500/10",
    text: "text-emerald-600 dark:text-emerald-400",
    bar: "bg-emerald-500",
  },
  blue: {
    bg: "bg-blue-500/10",
    text: "text-blue-600 dark:text-blue-400",
    bar: "bg-blue-500",
  },
  indigo: {
    bg: "bg-indigo-500/10",
    text: "text-indigo-600 dark:text-indigo-400",
    bar: "bg-indigo-500",
  },
  amber: {
    bg: "bg-amber-500/10",
    text: "text-amber-600 dark:text-amber-400",
    bar: "bg-amber-500",
  },
  rose: {
    bg: "bg-rose-500/10",
    text: "text-rose-600 dark:text-rose-400",
    bar: "bg-rose-500",
  },
  teal: {
    bg: "bg-teal-500/10",
    text: "text-teal-600 dark:text-teal-400",
    bar: "bg-teal-500",
  },
};

export const KPIMetricCard: React.FC<KPIMetricCardProps> = ({
  title,
  value,
  change,
  changeType = "neutral",
  changeNote,
  icon: Icon,
  progressPercent,
  accentColor = "blue",
  className,
}) => {
  const c = colorMap[accentColor] || colorMap.blue;

  return (
    <div
      className={cn(
        "rounded-xl border bg-card p-5 space-y-3 shadow-xs hover:shadow-md transition-all group",
        className
      )}
    >
      <div className="flex items-center justify-between">
        <span className="text-xs font-semibold text-muted-foreground">{title}</span>
        <div
          className={cn(
            "h-9 w-9 rounded-lg flex items-center justify-center group-hover:scale-105 transition-transform",
            c.bg,
            c.text
          )}
        >
          <Icon className="h-5 w-5" />
        </div>
      </div>

      <div className="space-y-1">
        <div className="text-2xl font-bold tracking-tight text-foreground">{value}</div>
        {change && (
          <div className="flex items-center gap-1 text-xs font-medium">
            {changeType === "positive" ? (
              <ArrowUpRight className="h-3.5 w-3.5 text-emerald-600" />
            ) : changeType === "negative" ? (
              <ArrowDownRight className="h-3.5 w-3.5 text-rose-600" />
            ) : (
              <Minus className="h-3.5 w-3.5 text-muted-foreground" />
            )}
            <span
              className={
                changeType === "positive"
                  ? "text-emerald-600 dark:text-emerald-400"
                  : changeType === "negative"
                  ? "text-rose-600 dark:text-rose-400"
                  : "text-muted-foreground"
              }
            >
              {change}
            </span>
            {changeNote && (
              <span className="text-muted-foreground font-normal ps-1">{changeNote}</span>
            )}
          </div>
        )}
      </div>

      {progressPercent !== undefined && (
        <div className="h-1.5 w-full bg-muted rounded-full overflow-hidden">
          <div
            className={cn("h-full rounded-full transition-all duration-500", c.bar)}
            style={{ width: `${Math.min(100, Math.max(0, progressPercent))}%` }}
          />
        </div>
      )}
    </div>
  );
};
