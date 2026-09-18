"use client";

import * as React from "react";
import { cn } from "@/lib/utils";

interface TabsContextValue {
  activeTab: string;
  setActiveTab: (tab: string) => void;
}

const TabsContext = React.createContext<TabsContextValue | null>(null);

export interface TabsProps extends React.HTMLAttributes<HTMLDivElement> {
  defaultValue: string;
  value?: string;
  onValueChange?: (value: string) => void;
}

const Tabs: React.FC<TabsProps> = ({
  defaultValue,
  value,
  onValueChange,
  className,
  children,
  ...props
}) => {
  const [activeTabState, setActiveTabState] = React.useState(defaultValue);
  const activeTab = value !== undefined ? value : activeTabState;

  const setActiveTab = (val: string) => {
    if (value === undefined) {
      setActiveTabState(val);
    }
    onValueChange?.(val);
  };

  return (
    <TabsContext.Provider value={{ activeTab, setActiveTab }}>
      <div className={cn("w-full space-y-4", className)} {...props}>
        {children}
      </div>
    </TabsContext.Provider>
  );
};

const TabsList = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => (
  <div
    ref={ref}
    className={cn(
      "inline-flex h-10 items-center justify-center rounded-xl bg-muted/70 p-1 text-muted-foreground border",
      className
    )}
    {...props}
  />
));
TabsList.displayName = "TabsList";

export interface TabsTriggerProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  value: string;
}

const TabsTrigger = React.forwardRef<HTMLButtonElement, TabsTriggerProps>(
  ({ className, value, children, ...props }, ref) => {
    const ctx = React.useContext(TabsContext);
    const isActive = ctx?.activeTab === value;

    return (
      <button
        ref={ref}
        type="button"
        onClick={() => ctx?.setActiveTab(value)}
        className={cn(
          "inline-flex items-center justify-center whitespace-nowrap rounded-lg px-3.5 py-1.5 text-xs font-semibold ring-offset-background transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50",
          isActive
            ? "bg-card text-foreground shadow-sm font-bold"
            : "hover:bg-background/50 hover:text-foreground",
          className
        )}
        {...props}
      >
        {children}
      </button>
    );
  }
);
TabsTrigger.displayName = "TabsTrigger";

export interface TabsContentProps
  extends React.HTMLAttributes<HTMLDivElement> {
  value: string;
}

const TabsContent = React.forwardRef<HTMLDivElement, TabsContentProps>(
  ({ className, value, children, ...props }, ref) => {
    const ctx = React.useContext(TabsContext);
    if (ctx?.activeTab !== value) return null;

    return (
      <div
        ref={ref}
        className={cn("mt-2 ring-offset-background focus-visible:outline-none", className)}
        {...props}
      >
        {children}
      </div>
    );
  }
);
TabsContent.displayName = "TabsContent";

export { Tabs, TabsList, TabsTrigger, TabsContent };
