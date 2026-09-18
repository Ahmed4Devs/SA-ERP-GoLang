"use client";

import React, { useState, useEffect } from "react";
import { cn } from "@/lib/utils";

interface MoneyInputProps {
  value: number;
  onChange?: (val: number) => void;
  currency?: string;
  disabled?: boolean;
  min?: number;
  max?: number;
  placeholder?: string;
  className?: string;
  isArabic?: boolean;
}

export const MoneyInput: React.FC<MoneyInputProps> = ({
  value,
  onChange,
  currency = "SAR",
  disabled = false,
  min = 0,
  max,
  placeholder = "0.00",
  className,
  isArabic = true,
}) => {
  const [displayValue, setDisplayValue] = useState(value !== undefined ? value.toString() : "");

  useEffect(() => {
    setDisplayValue(value !== undefined ? value.toString() : "");
  }, [value]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const raw = e.target.value;
    // Allow empty, digits, and a single decimal point
    if (raw === "" || /^\d*\.?\d{0,4}$/.test(raw)) {
      setDisplayValue(raw);
      const parsed = parseFloat(raw);
      if (!isNaN(parsed)) {
        onChange?.(parsed);
      } else {
        onChange?.(0);
      }
    }
  };

  return (
    <div className={cn("relative flex items-center w-full", className)}>
      <input
        type="text"
        inputMode="decimal"
        value={displayValue}
        onChange={handleChange}
        disabled={disabled}
        placeholder={placeholder}
        className={cn(
          "w-full h-9 rounded-lg border border-input bg-background ps-3 pe-12 text-xs font-mono font-medium text-foreground shadow-xs transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50 text-start",
          className
        )}
      />
      <div className="absolute inset-inline-end-1 top-1/2 -translate-y-1/2 flex items-center pointer-events-none">
        <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-muted text-muted-foreground border">
          {currency}
        </span>
      </div>
    </div>
  );
};
