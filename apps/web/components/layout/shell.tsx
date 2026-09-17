"use client";

import React, { createContext, useContext, useEffect, useState } from "react";
import { usePathname } from "next/navigation";
import { useTheme } from "next-themes";
import {
  LayoutDashboard,
  Building2,
  ShieldCheck,
  BookOpen,
  ShoppingBag,
  ShoppingCart,
  Boxes,
  Landmark,
  Users,
  Banknote,
  FileCheck2,
  ReceiptText,
  BarChart3,
  Settings,
  Bell,
  Search,
  Moon,
  Sun,
  Globe,
  ChevronDown,
  Menu,
  X,
  History,
  GitMerge,
  Package,
  Truck,
  Receipt,
  Clock,
  Calendar,
} from "lucide-react";

import arDict from "@/messages/ar.json";
import enDict from "@/messages/en.json";

type Locale = "ar" | "en";

interface LocaleContextType {
  locale: Locale;
  setLocale: (locale: Locale) => void;
  t: typeof arDict;
  dir: "rtl" | "ltr";
}

const LocaleContext = createContext<LocaleContextType>({
  locale: "ar",
  setLocale: () => {},
  t: arDict,
  dir: "rtl",
});

export function useLocale() {
  return useContext(LocaleContext);
}

export function ShellProvider({ children }: { children: React.ReactNode }) {
  const [locale, setLocaleState] = useState<Locale>("ar");
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    const saved = localStorage.getItem("erp_locale") as Locale;
    if (saved === "ar" || saved === "en") {
      setLocaleState(saved);
    }
    setMounted(true);
  }, []);

  const setLocale = (newLocale: Locale) => {
    setLocaleState(newLocale);
    localStorage.setItem("erp_locale", newLocale);
    document.documentElement.dir = newLocale === "ar" ? "rtl" : "ltr";
    document.documentElement.lang = newLocale === "ar" ? "ar-SA" : "en";
  };

  useEffect(() => {
    if (mounted) {
      document.documentElement.dir = locale === "ar" ? "rtl" : "ltr";
      document.documentElement.lang = locale === "ar" ? "ar-SA" : "en";
    }
  }, [locale, mounted]);

  const dict = locale === "ar" ? arDict : enDict;
  const dir = locale === "ar" ? "rtl" : "ltr";

  return (
    <LocaleContext.Provider value={{ locale, setLocale, t: dict, dir }}>
      {children}
    </LocaleContext.Provider>
  );
}

export function Shell({ children }: { children: React.ReactNode }) {
  const { locale, setLocale, t, dir } = useLocale();
  const { theme, setTheme } = useTheme();
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  const pathname = usePathname();

  const navItems = [
    { key: "dashboard", href: "/", icon: LayoutDashboard, label: t.nav.dashboard, active: pathname === "/" },
    { key: "organization", href: "/organization", icon: Building2, label: t.nav.organization, active: pathname === "/organization" },
    { key: "iam", href: "/authorization", icon: ShieldCheck, label: t.nav.iam, active: pathname === "/authorization" },
    { key: "audit", href: "/audit", icon: History, label: t.nav.audit, active: pathname === "/audit" },
    { key: "workflow", href: "/workflow", icon: GitMerge, label: t.nav.workflow, active: pathname === "/workflow" },
    { key: "accounting", href: "/accounting", icon: BookOpen, label: t.nav.accounting, active: pathname === "/accounting" },
    { key: "tax", href: "/tax", icon: ReceiptText, label: t.nav.tax || "Tax & VAT", active: pathname === "/tax" },
    { key: "products", href: "/products", icon: Package, label: (t.products as any)?.tab_items || "الأصناف والمنتجات", active: pathname === "/products" },
    { key: "inventory", href: "/inventory", icon: Boxes, label: t.nav.inventory, active: pathname === "/inventory" },
    { key: "customers", href: "/customers", icon: Users, label: (t.nav as any)?.customers || "العملاء والمدينون (AR)", active: pathname === "/customers" },
    { key: "sales", href: "/sales", icon: ShoppingBag, label: t.nav.sales || "المبيعات والفوترة", active: pathname === "/sales" },
    { key: "suppliers", href: "/suppliers", icon: Truck, label: (t.nav as any)?.suppliers || "الموردون والدائنون (AP)", active: pathname === "/suppliers" },
    { key: "payables", href: "/payables", icon: Receipt, label: (t.nav as any)?.payables || "الفواتير وسندات الصرف", active: pathname === "/payables" },
    { key: "purchasing", href: "/purchasing", icon: ShoppingCart, label: t.nav.purchasing || "المشتريات والتعميد", active: pathname === "/purchasing" },
    { key: "treasury", href: "/treasury", icon: Landmark, label: t.nav.treasury || "الخزينة والسيولة", active: pathname === "/treasury" },
    { key: "banking", href: "/banking", icon: Building2, label: (t.nav as any)?.banking || "البنوك والمطابقات", active: pathname === "/banking" },
    { key: "hr", href: "/hr", icon: Users, label: t.nav.hr, active: pathname === "/hr" },
    { key: "attendance", href: "/attendance", icon: Clock, label: (t.nav as any)?.attendance || "الحضور والانصراف", active: pathname === "/attendance" },
    { key: "leave", href: "/leave", icon: Calendar, label: (t.nav as any)?.leave || "الإجازات والأرصدة", active: pathname === "/leave" },
    { key: "payroll", href: "/payroll", icon: Banknote, label: t.nav.payroll, active: pathname === "/payroll" },
    { key: "gosi", href: "/gosi", icon: FileCheck2, label: t.nav.gosi, active: pathname === "/gosi" },
    { key: "zatca", href: "#", icon: ReceiptText, label: t.nav.zatca },
    { key: "reports", href: "#", icon: BarChart3, label: t.nav.reports },
    { key: "settings", href: "#", icon: Settings, label: t.nav.settings },
  ];

  return (
    <div className="min-h-screen flex flex-col bg-background text-foreground" dir={dir}>
      {/* Topbar */}
      <header className="sticky top-0 z-40 w-full border-b bg-card/95 backdrop-blur supports-[backdrop-filter]:bg-card/60">
        <div className="flex h-16 items-center justify-between px-4 sm:px-6">
          {/* Brand & Mobile Toggle */}
          <div className="flex items-center gap-3">
            <button
              onClick={() => setSidebarOpen(!sidebarOpen)}
              className="md:hidden p-2 rounded-md hover:bg-muted text-foreground"
              aria-label="Toggle menu"
            >
              {sidebarOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
            </button>
            <div className="flex items-center gap-2">
              <div className="h-9 w-9 rounded-lg bg-primary flex items-center justify-center text-primary-foreground font-bold text-lg shadow-sm">
                ERP
              </div>
              <div className="hidden sm:block">
                <span className="font-bold text-sm leading-none block">SAUDI ERP</span>
                <span className="text-[10px] text-muted-foreground leading-tight block">
                  Enterprise Core
                </span>
              </div>
            </div>
          </div>

          {/* Active Entity & Branch Badge */}
          <div className="hidden lg:flex items-center gap-2">
            <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-medium bg-secondary text-secondary-foreground border">
              <Building2 className="h-3.5 w-3.5 text-primary" />
              <span>{t.topbar.legal_entity}</span>
              <span className="text-muted-foreground">|</span>
              <span>{t.topbar.branch}</span>
            </div>
          </div>

          {/* Search Box */}
          <div className="hidden md:flex items-center flex-1 max-w-sm mx-4">
            <div className="relative w-full">
              <Search className="absolute inset-inline-start-3 top-2.5 h-4 w-4 text-muted-foreground" />
              <input
                type="text"
                placeholder={t.topbar.search}
                className="w-full rounded-md border bg-background ps-9 pe-4 py-1.5 text-xs text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring"
              />
            </div>
          </div>

          {/* Controls: Lang Switch, Theme Toggle, Notifications, User */}
          <div className="flex items-center gap-2">
            {/* Language Switcher */}
            <button
              id="lang-toggle-btn"
              onClick={() => setLocale(locale === "ar" ? "en" : "ar")}
              className="inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-md text-xs font-medium border hover:bg-muted transition-colors"
              title="Switch Language / تبديل اللغة"
            >
              <Globe className="h-4 w-4" />
              <span>{locale === "ar" ? "English" : "العربية"}</span>
            </button>

            {/* Theme Toggle */}
            {mounted && (
              <button
                id="theme-toggle-btn"
                onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
                className="p-2 rounded-md border hover:bg-muted transition-colors text-foreground"
                title={theme === "dark" ? t.topbar.light : t.topbar.dark}
                aria-label="Toggle theme"
              >
                {theme === "dark" ? (
                  <Sun className="h-4 w-4 text-yellow-400" />
                ) : (
                  <Moon className="h-4 w-4" />
                )}
              </button>
            )}

            {/* Notifications */}
            <button
              className="relative p-2 rounded-md hover:bg-muted transition-colors text-foreground"
              aria-label={t.topbar.notifications}
            >
              <Bell className="h-4 w-4" />
              <span className="absolute top-1.5 inset-inline-end-1.5 h-2 w-2 rounded-full bg-destructive" />
            </button>

            {/* User Profile */}
            <div className="flex items-center gap-2 ps-2 border-s">
              <div className="h-8 w-8 rounded-full bg-primary/20 text-primary flex items-center justify-center font-semibold text-xs border border-primary/30">
                SA
              </div>
              <div className="hidden xl:block text-start">
                <p className="text-xs font-semibold leading-none">{t.topbar.user_role}</p>
                <p className="text-[10px] text-muted-foreground mt-0.5">admin@erp.local</p>
              </div>
            </div>
          </div>
        </div>
      </header>

      {/* Main App Layout */}
      <div className="flex flex-1 overflow-hidden">
        {/* Sidebar */}
        <aside
          className={`
            fixed md:static inset-y-0 z-30 flex-shrink-0 w-64 bg-card border-e transition-transform duration-200 ease-in-out
            ${dir === "rtl" ? "right-0" : "left-0"}
            ${sidebarOpen ? "translate-x-0" : dir === "rtl" ? "translate-x-full md:translate-x-0" : "-translate-x-full md:translate-x-0"}
            top-16 bottom-0
          `}
        >
          <div className="h-full flex flex-col justify-between py-4 px-3 overflow-y-auto">
            <div className="space-y-1">
              <div className="px-3 py-2 text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                {t.nav.dashboard}
              </div>
              {navItems.map((item) => {
                const Icon = item.icon;
                return (
                  <a
                    key={item.key}
                    href={item.href}
                    className={`
                      flex items-center gap-3 px-3 py-2 rounded-md text-xs font-medium transition-colors
                      ${
                        item.active
                          ? "bg-primary text-primary-foreground shadow-sm"
                          : "text-muted-foreground hover:bg-muted hover:text-foreground"
                      }
                    `}
                  >
                    <Icon className="h-4 w-4 flex-shrink-0" />
                    <span>{item.label}</span>
                  </a>
                );
              })}
            </div>

            {/* Footer / System Info */}
            <div className="pt-4 border-t px-3 text-[11px] text-muted-foreground">
              <p className="font-semibold text-foreground">ERP Platform v0.1.0</p>
              <p className="mt-0.5">ZATCA & GOSI Ready</p>
            </div>
          </div>
        </aside>

        {/* Backdrop for mobile */}
        {sidebarOpen && (
          <div
            onClick={() => setSidebarOpen(false)}
            className="fixed inset-0 z-20 bg-background/80 backdrop-blur-sm md:hidden"
          />
        )}

        {/* Content Workspace */}
        <main className="flex-1 overflow-y-auto p-4 sm:p-6 lg:p-8">
          <div className="max-w-7xl mx-auto">{children}</div>
        </main>
      </div>
    </div>
  );
}
