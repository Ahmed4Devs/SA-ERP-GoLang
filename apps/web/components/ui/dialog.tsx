"use client";

import * as React from "react";
import { X } from "lucide-react";
import { cn } from "@/lib/utils";

interface DialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  children: React.ReactNode;
}

const Dialog: React.FC<DialogProps> = ({ open, onOpenChange, children }) => {
  React.useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape" && open) {
        onOpenChange(false);
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [open, onOpenChange]);

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-background/80 backdrop-blur-sm transition-opacity"
        onClick={() => onOpenChange(false)}
      />
      {/* Modal Box */}
      <div className="relative z-50 w-full max-w-lg overflow-hidden rounded-2xl border bg-card p-6 shadow-xl transition-all">
        <button
          onClick={() => onOpenChange(false)}
          className="absolute top-4 inset-inline-end-4 p-1.5 rounded-lg text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
          aria-label="Close"
        >
          <X className="h-4 w-4" />
        </button>
        {children}
      </div>
    </div>
  );
};

const DialogHeader: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  className,
  ...props
}) => (
  <div className={cn("flex flex-col space-y-1.5 text-start pb-4 border-b", className)} {...props} />
);

const DialogTitle: React.FC<React.HTMLAttributes<HTMLHeadingElement>> = ({
  className,
  ...props
}) => (
  <h2
    className={cn("text-base font-bold text-foreground leading-none tracking-tight", className)}
    {...props}
  />
);

const DialogDescription: React.FC<React.HTMLAttributes<HTMLParagraphElement>> = ({
  className,
  ...props
}) => (
  <p className={cn("text-xs text-muted-foreground leading-relaxed", className)} {...props} />
);

const DialogFooter: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  className,
  ...props
}) => (
  <div
    className={cn(
      "flex flex-col-reverse sm:flex-row sm:justify-end sm:gap-2 pt-4 border-t mt-4",
      className
    )}
    {...props}
  />
);

export { Dialog, DialogHeader, DialogTitle, DialogDescription, DialogFooter };
