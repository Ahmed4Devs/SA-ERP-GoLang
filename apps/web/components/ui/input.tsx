import * as React from "react";
import { cn } from "@/lib/utils";

export interface InputProps
  extends React.InputHTMLAttributes<HTMLInputElement> {
  startIcon?: React.ReactNode;
  endIcon?: React.ReactNode;
}

const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, type, startIcon, endIcon, ...props }, ref) => {
    return (
      <div className="relative w-full">
        {startIcon && (
          <div className="absolute inset-inline-start-3 top-1/2 -translate-y-1/2 text-muted-foreground pointer-events-none flex items-center">
            {startIcon}
          </div>
        )}
        <input
          type={type}
          className={cn(
            "flex h-9 w-full rounded-lg border border-input bg-background px-3 py-1.5 text-xs text-foreground shadow-sm transition-colors file:border-0 file:bg-transparent file:text-xs file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50",
            startIcon && "ps-9",
            endIcon && "pe-9",
            className
          )}
          ref={ref}
          {...props}
        />
        {endIcon && (
          <div className="absolute inset-inline-end-3 top-1/2 -translate-y-1/2 text-muted-foreground flex items-center">
            {endIcon}
          </div>
        )}
      </div>
    );
  }
);
Input.displayName = "Input";

export { Input };
