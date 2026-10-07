"use client";

import React, { createContext, useContext, useEffect, useState } from "react";
import Link from "next/link";
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
  Menu,
  X,
  History,
  GitMerge,
  Package,
  Truck,
  Receipt,
  Clock,
  Calendar,
  Layers,
  PieChart,
  FolderArchive,
  ChevronRight,
  Sparkles,
  LogOut,
} from "lucide-react";

import arDict from "@/messages/ar.json";
import enDict from "@/messages/en.json";
import { useAuth } from "@/lib/auth-context";

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
    document.documentElement.lang = newLocale === "ar" ? "ar" : "en";
  };

  useEffect(() => {
    if (mounted) {
      document.documentElement.dir = locale === "ar" ? "rtl" : "ltr";
      document.documentElement.lang = locale === "ar" ? "ar" : "en";
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
  const isAr = locale === "ar";
  const { theme, setTheme } = useTheme();
  const [navigationSearch, setNavigationSearch] = useState("");
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [mounted, setMounted] = useState(false);
  const { user, assignments, legalEntityId, branchId, branches, branchesLoading, branchError, canSelectAllBranches, selectBranch, loading, authError, scopeError, refresh, selectLegalEntity, logout } = useAuth();

  useEffect(() => {
    setMounted(true);
  }, []);

  const pathname = usePathname();

  if (pathname === "/login") {
    return <>{children}</>;
  }

  if (loading) {
    return <div className="min-h-screen grid place-items-center text-sm text-muted-foreground">جاري التحقق من الجلسة…</div>;
  }

  if (authError && !user) {
    return (
      <div className="min-h-screen grid place-items-center p-6" dir={dir}>
        <div role="alert" className="max-w-lg rounded-xl border border-destructive/40 bg-destructive/10 p-5 text-sm text-destructive">
          <p className="font-semibold">{isAr ? "تعذر التحقق من الجلسة" : "Unable to verify your session"}</p>
          <p className="mt-2">{authError}</p>
          <button type="button" onClick={() => void refresh()} className="mt-4 rounded-lg bg-primary px-4 py-2 text-primary-foreground">
            {isAr ? "إعادة المحاولة" : "Try again"}
          </button>
        </div>
      </div>
    );
  }

  if (!user) return null;

  const entityOptions = assignments.filter((assignment, index, all) => assignment.legal_entity_id && all.findIndex((other) => other.legal_entity_id === assignment.legal_entity_id) === index);
  const activeEntity = entityOptions.find((assignment) => assignment.legal_entity_id === legalEntityId);

  const navItems = [
    { key: "dashboard", href: "/", icon: LayoutDashboard, label: t.nav.dashboard, active: pathname === "/" },
    { key: "design-system", href: "/design-system", icon: Sparkles, label: isAr ? "نظام التصميم UI" : "Design System", active: pathname === "/design-system" },
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
    { key: "fixed-assets", href: "/fixed-assets", icon: Layers, label: (t.nav as any)?.fixedAssets || "الأصول الثابتة وإهلاكها", active: pathname === "/fixed-assets" },
    { key: "budgets", href: "/budgets", icon: PieChart, label: (t.nav as any)?.budgets || "الميزانيات التقديرية", active: pathname === "/budgets" },
    { key: "zatca", href: "/zatca", icon: ReceiptText, label: t.nav.zatca, active: pathname === "/zatca" },
    { key: "documents", href: "/documents", icon: FolderArchive, label: (t.nav as any)?.documents || "إدارة المستندات والوثائق", active: pathname === "/documents" },
    { key: "notifications", href: "/notifications", icon: Bell, label: (t.nav as any)?.notifications || "مركز التنبيهات والإشعارات", active: pathname === "/notifications" },
    { key: "reports", href: "/reports", icon: BarChart3, label: t.nav.reports, active: pathname === "/reports" },
    { key: "settings", href: "/settings", icon: Settings, label: t.nav.settings, active: pathname === "/settings" },
  ];

  // Get active item for breadcrumb
  const currentNav = navItems.find((n) => n.active) || navItems[0];

  return (
    <div className="min-h-screen flex flex-col bg-background text-foreground" dir={dir}>
      {/* Topbar */}
      <header className="sticky top-0 z-40 w-full border-b bg-card/95 backdrop-blur supports-[backdrop-filter]:bg-card/75 shadow-xs">
        <div className="flex min-h-16 flex-wrap items-center justify-between gap-y-2 px-4 py-2 sm:px-6 lg:flex-nowrap">
          {/* Brand & Mobile Toggle */}
          <div className="flex items-center gap-3">
            <button
              onClick={() => setSidebarOpen(!sidebarOpen)}
              className="md:hidden p-2 rounded-lg hover:bg-muted text-foreground transition-colors"
              aria-label="Toggle menu"
            >
              {sidebarOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
            </button>
            <Link href="/" className="flex items-center gap-2.5">
              <div className="h-9 w-9 rounded-xl bg-primary flex items-center justify-center text-primary-foreground font-bold text-sm shadow-sm">
                ERP
              </div>
              <div className="hidden sm:block">
                <span className="font-bold text-sm leading-none block">SAUDI ERP</span>
                <span className="text-[10px] text-muted-foreground leading-tight block mt-0.5">
                  Enterprise Core
                </span>
              </div>
            </Link>
          </div>

          {/* Active Entity & Branch Badge */}
          <div className="order-last flex basis-full items-center gap-2 lg:order-none lg:basis-auto">
            <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-medium bg-secondary text-secondary-foreground border">
              <Building2 className="h-3.5 w-3.5 text-primary" />
              {entityOptions.length > 0 ? (
                <select
                  aria-label={t.topbar.legal_entity}
                  value={legalEntityId ?? ""}
                  onChange={(event) => selectLegalEntity(event.target.value)}
                  className="bg-transparent outline-none max-w-64"
                >
                  {entityOptions.map((assignment) => (
                    <option key={assignment.legal_entity_id} value={assignment.legal_entity_id}>
                      {(isAr ? assignment.legal_entity_name_ar : assignment.legal_entity_name_en) || assignment.legal_entity_id}
                    </option>
                  ))}
                </select>
              ) : (
                <span>{activeEntity?.legal_entity_name_ar ?? t.topbar.legal_entity}</span>
              )}
            </div>
            <label className="inline-flex items-center gap-2 rounded-full border bg-secondary px-3 py-1 text-xs">
              <span>{t.business.branch}</span>
              <select aria-label={t.business.branch} value={branchId ?? ""} disabled={branchesLoading || Boolean(branchError)} onChange={event => selectBranch(event.target.value)} className="max-w-48 bg-transparent outline-none disabled:opacity-50">
                {canSelectAllBranches && <option value="">{t.business.allBranches}</option>}
                {!canSelectAllBranches && !branches.length && <option value="">{branchesLoading ? t.business.branchLoading : t.business.noValues}</option>}
                {branches.map(branch => <option key={branch.id} value={branch.id}>{(isAr ? branch.name_ar : branch.name_en) || branch.code || branch.id}</option>)}
              </select>
            </label>
          </div>

          {/* Search Box */}
          <div className="hidden md:flex items-center flex-1 max-w-sm mx-4">
            <div className="relative w-full">
              <Search className="absolute inset-inline-start-3 top-2.5 h-4 w-4 text-muted-foreground" />
              <input
                type="text"
                aria-label={t.business.searchModules}
                value={navigationSearch}
                onChange={event => setNavigationSearch(event.target.value)}
                onKeyDown={event => { if (event.key === "Escape") setNavigationSearch(""); }}
                placeholder={t.business.searchModules}
                className="w-full rounded-lg border bg-background ps-9 pe-4 py-1.5 text-xs text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring transition-all"
              />
              {navigationSearch.trim() && <div role="region" aria-label={t.business.searchModules} className="absolute top-full z-50 mt-2 max-h-80 w-full overflow-y-auto rounded-lg border bg-card p-2 shadow-lg">
                {navItems.filter(item => `${item.label} ${item.key}`.toLocaleLowerCase().includes(navigationSearch.toLocaleLowerCase().trim())).map(item => <Link key={item.key} href={item.href} onClick={() => setNavigationSearch("")} className="block rounded px-3 py-2 text-sm hover:bg-muted">{item.label}</Link>)}
                {!navItems.some(item => `${item.label} ${item.key}`.toLocaleLowerCase().includes(navigationSearch.toLocaleLowerCase().trim())) && <p className="p-3 text-sm text-muted-foreground">{t.business.noPages}</p>}
              </div>}
            </div>
          </div>

          {/* Controls: Lang Switch, Theme Toggle, Notifications, User */}
          <div className="flex items-center gap-2">
            {/* Language Switcher */}
            <button
              id="lang-toggle-btn"
              onClick={() => setLocale(locale === "ar" ? "en" : "ar")}
              className="inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg text-xs font-semibold border hover:bg-muted transition-colors"
              title="Switch Language / تبديل اللغة"
            >
              <Globe className="h-4 w-4 text-primary" />
              <span>{locale === "ar" ? "English" : "العربية"}</span>
            </button>

            {/* Theme Toggle */}
            {mounted && (
              <label
                id="theme-toggle-btn"
                className="inline-flex items-center gap-1 rounded-lg border p-1.5 text-xs text-foreground"
              >
                {theme === "dark" ? (
                  <Sun className="h-4 w-4 text-amber-400" />
                ) : (
                  <Moon className="h-4 w-4 text-blue-600" />
                )}
                <select aria-label={t.topbar.theme} value={theme ?? "system"} onChange={(event) => setTheme(event.target.value)} className="max-w-20 bg-background outline-none">
                  <option value="light">{t.topbar.light}</option>
                  <option value="dark">{t.topbar.dark}</option>
                  <option value="system">{t.topbar.system}</option>
                </select>
              </label>
            )}

            {/* Notifications */}
            <Link
              href="/notifications"
              className="relative p-2 rounded-lg hover:bg-muted transition-colors text-foreground"
              aria-label={t.topbar.notifications}
            >
              <Bell className="h-4 w-4" />
              <span className="absolute top-1.5 inset-inline-end-1.5 h-2 w-2 rounded-full bg-destructive animate-pulse" />
            </Link>

            {/* User Profile */}
            <div className="flex items-center gap-2 ps-2 border-s">
              <div className="h-8 w-8 rounded-full bg-primary/10 text-primary flex items-center justify-center font-bold text-xs border border-primary/20">
                {user.display_name.slice(0, 2).toUpperCase()}
              </div>
              <div className="hidden xl:block text-start">
                <p className="text-xs font-semibold leading-none">{user.display_name}</p>
                <p className="text-[10px] text-muted-foreground mt-0.5">{user.email || user.username}</p>
              </div>
              <button
                type="button"
                onClick={() => void logout()}
                className="p-2 rounded-lg hover:bg-muted"
                aria-label={isAr ? "تسجيل الخروج" : "Sign out"}
              >
                <LogOut className="h-4 w-4" />
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Main App Layout */}
      <div className="flex flex-1 overflow-hidden">
        {/* Sidebar */}
        <aside
          className={`
            fixed md:static inset-y-0 z-30 flex-shrink-0 w-64 bg-card border-inline-end transition-transform duration-200 ease-in-out
            ${dir === "rtl" ? "right-0" : "left-0"}
            ${sidebarOpen ? "translate-x-0" : dir === "rtl" ? "translate-x-full md:translate-x-0" : "-translate-x-full md:translate-x-0"}
            top-16 bottom-0
          `}
        >
          <div className="h-full flex flex-col justify-between py-4 px-3 overflow-y-auto">
            <div className="space-y-1">
              <div className="px-3 py-2 text-[10px] font-bold text-muted-foreground uppercase tracking-wider">
                {isAr ? "الوحدات والأنظمة الرئيسية" : "Main System Modules"}
              </div>
              {navItems.map((item) => {
                const Icon = item.icon;
                return (
                  <Link
                    key={item.key}
                    href={item.href}
                    onClick={() => setSidebarOpen(false)}
                    className={`
                      flex items-center gap-3 px-3 py-2 rounded-lg text-xs font-semibold transition-colors
                      ${
                        item.active
                          ? "bg-primary text-primary-foreground shadow-xs font-bold"
                          : "text-muted-foreground hover:bg-muted hover:text-foreground"
                      }
                    `}
                  >
                    <Icon className="h-4 w-4 flex-shrink-0" />
                    <span>{item.label}</span>
                  </Link>
                );
              })}
            </div>

            {/* Footer / System Info */}
            <div className="pt-4 border-t px-3 text-[11px] text-muted-foreground">
              <p className="font-semibold text-foreground">ERP Platform v1.0.0</p>
              <p className="mt-0.5">Saudi ERP Core</p>
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
          <div className="max-w-7xl mx-auto space-y-4">
            {branchError && <p role="alert" className="rounded-lg border border-amber-500/40 p-3 text-sm">{t.business.branchError}: {branchError}</p>}
            {scopeError && (
              <div role="alert" className="rounded-lg border border-amber-500/40 bg-amber-500/10 p-3 text-sm text-amber-700 dark:text-amber-300">
                {isAr ? "تم تسجيل الدخول، لكن تعذر تحميل نطاقات الوصول: " : "Signed in, but access scopes could not be loaded: "}
                {scopeError}
              </div>
            )}
            {/* Breadcrumb strip */}
            {pathname !== "/" && (
              <div className="flex items-center gap-2 text-xs text-muted-foreground pb-2">
                <Link href="/" className="hover:text-foreground transition-colors">
                  {t.nav.dashboard}
                </Link>
                <ChevronRight className="h-3.5 w-3.5 rtl:rotate-180 text-muted-foreground/40" />
                <span className="font-semibold text-foreground">{currentNav.label}</span>
              </div>
            )}
            {children}
          </div>
        </main>
      </div>
    </div>
  );
}
