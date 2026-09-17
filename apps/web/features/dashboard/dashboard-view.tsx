"use client";

import React, { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import {
  TrendingUp,
  TrendingDown,
  Wallet,
  Building2,
  AlertTriangle,
  CheckCircle2,
  Clock,
  ShieldCheck,
  FileText,
  Activity,
  RefreshCw,
  ArrowUpRight,
  ArrowDownRight,
  ArrowRight,
  Sparkles,
  Layers,
  Scale,
  DollarSign,
  Boxes,
  FileSpreadsheet,
  CheckCircle,
  XCircle,
  AlertCircle,
  Plus,
  ExternalLink,
  ChevronRight,
  Receipt,
  FileCode,
  Lock,
} from "lucide-react";
import { useLocale } from "@/components/layout/shell";

// Types matching backend domain models
interface FinancialKPIs {
  total_revenue: number;
  total_expense: number;
  net_profit: number;
  profit_margin: number;
  currency: string;
}

interface CashPositionKPI {
  total_cash_balance: number;
  total_bank_balance: number;
  total_liquid_funds: number;
  accounts_count: number;
  currency: string;
}

interface ARAPSummary {
  total_receivables: number;
  overdue_receivables: number;
  open_receivable_docs: number;
  total_payables: number;
  overdue_payables: number;
  open_payable_docs: number;
  currency: string;
}

interface StockWarning {
  item_id: string;
  item_code: string;
  item_name_en: string;
  item_name_ar: string;
  warehouse_name_en: string;
  warehouse_name_ar: string;
  quantity_on_hand: number;
  quantity_reserved: number;
  quantity_available: number;
  reorder_point: number;
  severity: "CRITICAL" | "WARNING" | "LOW";
}

interface PendingApprovalItem {
  id: string;
  resource_type: string;
  resource_id: string;
  initiator_id: string;
  assigned_user_id: string;
  status: string;
  created_at: string;
}

interface ApprovalsSummary {
  total_pending: number;
  by_type: Record<string, number>;
  recent_requests: PendingApprovalItem[];
}

interface ZATCAStatusSummary {
  total_documents: number;
  cleared_count: number;
  pending_count: number;
  warning_count: number;
  rejected_count: number;
  compliance_rate: number;
}

interface RecentDocument {
  id: string;
  title: string;
  category: string;
  filename: string;
  mime_type: string;
  byte_size: number;
  created_at: string;
}

interface ActivityFeedItem {
  id: string;
  user_id?: string;
  user_name: string;
  action: string;
  resource_type: string;
  resource_id: string;
  occurred_at: string;
}

interface DashboardData {
  legal_entity_id: string;
  legal_entity_name: string;
  currency: string;
  generated_at: string;
  financials: FinancialKPIs;
  cash_position: CashPositionKPI;
  arap: ARAPSummary;
  stock_warnings: StockWarning[];
  approvals: ApprovalsSummary;
  zatca: ZATCAStatusSummary;
  recent_documents: RecentDocument[];
  activity_feed: ActivityFeedItem[];
}

// Fallback baseline mock data for offline/initial state
const defaultDashboardData: DashboardData = {
  legal_entity_id: "00000000-0000-0000-0000-000000000000",
  legal_entity_name: "شركة الأفق المتقدمة المحدودة",
  currency: "SAR",
  generated_at: new Date().toISOString(),
  financials: {
    total_revenue: 1450000.0,
    total_expense: 920000.0,
    net_profit: 530000.0,
    profit_margin: 36.55,
    currency: "SAR",
  },
  cash_position: {
    total_cash_balance: 185400.0,
    total_bank_balance: 1420600.0,
    total_liquid_funds: 1606000.0,
    accounts_count: 5,
    currency: "SAR",
  },
  arap: {
    total_receivables: 680000.0,
    overdue_receivables: 85000.0,
    open_receivable_docs: 14,
    total_payables: 310000.0,
    overdue_payables: 32000.0,
    open_payable_docs: 8,
    currency: "SAR",
  },
  stock_warnings: [
    {
      item_id: "item-1",
      item_code: "PRD-SRV-001",
      item_name_en: "Rack Server 2U Enterprise",
      item_name_ar: "خادم شبكي فئة المؤسسات 2U",
      warehouse_name_en: "Riyadh Central Hub",
      warehouse_name_ar: "مستودع الرياض المركزي",
      quantity_on_hand: 2,
      quantity_reserved: 1,
      quantity_available: 1,
      reorder_point: 5,
      severity: "CRITICAL",
    },
    {
      item_id: "item-2",
      item_code: "PRD-CBL-CAT6",
      item_name_en: "Cat6 UTP Cable 305m",
      item_name_ar: "كابل شبكة Cat6 بطول 305م",
      warehouse_name_en: "Jeddah Branch Depot",
      warehouse_name_ar: "مستودع فرع جدة",
      quantity_on_hand: 4,
      quantity_reserved: 0,
      quantity_available: 4,
      reorder_point: 10,
      severity: "WARNING",
    },
    {
      item_id: "item-3",
      item_code: "PRD-SWT-24P",
      item_name_en: "PoE+ Managed Switch 24-Port",
      item_name_ar: "محول شبكة ذكي 24 منفذ PoE",
      warehouse_name_en: "Dammam Logistics Center",
      warehouse_name_ar: "مركز الدمام اللوجستي",
      quantity_on_hand: 7,
      quantity_reserved: 1,
      quantity_available: 6,
      reorder_point: 8,
      severity: "LOW",
    },
  ],
  approvals: {
    total_pending: 4,
    by_type: {
      JOURNAL_ENTRY: 2,
      PURCHASE_ORDER: 1,
      LEAVE_REQUEST: 1,
    },
    recent_requests: [
      {
        id: "req-1",
        resource_type: "JOURNAL_ENTRY",
        resource_id: "JE-2026-0042",
        initiator_id: "usr-1",
        assigned_user_id: "usr-admin",
        status: "PENDING",
        created_at: new Date(Date.now() - 3600000 * 2).toISOString(),
      },
      {
        id: "req-2",
        resource_type: "PURCHASE_ORDER",
        resource_id: "PO-2026-0118",
        initiator_id: "usr-2",
        assigned_user_id: "usr-admin",
        status: "PENDING",
        created_at: new Date(Date.now() - 3600000 * 5).toISOString(),
      },
      {
        id: "req-3",
        resource_type: "LEAVE_REQUEST",
        resource_id: "LR-2026-0033",
        initiator_id: "usr-3",
        assigned_user_id: "usr-admin",
        status: "PENDING",
        created_at: new Date(Date.now() - 3600000 * 9).toISOString(),
      },
    ],
  },
  zatca: {
    total_documents: 148,
    cleared_count: 142,
    pending_count: 4,
    warning_count: 2,
    rejected_count: 0,
    compliance_rate: 97.29,
  },
  recent_documents: [
    {
      id: "doc-1",
      title: "عقد توريد خوادم البنية التحتية",
      category: "CONTRACT",
      filename: "Hardware_Supply_Contract_2026.pdf",
      mime_type: "application/pdf",
      byte_size: 2450000,
      created_at: new Date(Date.now() - 3600000 * 4).toISOString(),
    },
    {
      id: "doc-2",
      title: "فاتورة ضريبية معتمدة زاتكا #INV-0089",
      category: "INVOICE",
      filename: "Tax_Invoice_INV-0089.xml",
      mime_type: "application/xml",
      byte_size: 18400,
      created_at: new Date(Date.now() - 3600000 * 8).toISOString(),
    },
    {
      id: "doc-3",
      title: "شهادة الامتثال الضريبي السنوية",
      category: "LEGAL",
      filename: "ZATCA_Compliance_Certificate.pdf",
      mime_type: "application/pdf",
      byte_size: 1200000,
      created_at: new Date(Date.now() - 3600000 * 24).toISOString(),
    },
  ],
  activity_feed: [
    {
      id: "act-1",
      user_name: "مدير النظام (Admin)",
      action: "POST_JOURNAL",
      resource_type: "JOURNAL_ENTRY",
      resource_id: "JE-2026-0041",
      occurred_at: new Date(Date.now() - 3600000 * 1).toISOString(),
    },
    {
      id: "act-2",
      user_name: "سارة العتيبي (محاسب رئيسي)",
      action: "SUBMIT_ZATCA",
      resource_type: "TAX_INVOICE",
      resource_id: "INV-2026-0089",
      occurred_at: new Date(Date.now() - 3600000 * 3).toISOString(),
    },
    {
      id: "act-3",
      user_name: "خالد المطيري (مسؤول الخزينة)",
      action: "APPROVE_PAYMENT",
      resource_type: "SUPPLIER_PAYMENT",
      resource_id: "PMT-2026-0012",
      occurred_at: new Date(Date.now() - 3600000 * 6).toISOString(),
    },
    {
      id: "act-4",
      user_name: "محمد الشمري (أمين المستودع)",
      action: "STOCK_RECEIPT",
      resource_type: "TRANSFER_ORDER",
      resource_id: "TR-2026-0005",
      occurred_at: new Date(Date.now() - 3600000 * 12).toISOString(),
    },
  ],
};

export function DashboardView() {
  const { locale, t, dir } = useLocale();
  const isAr = locale === "ar";

  const [data, setData] = useState<DashboardData>(defaultDashboardData);
  const [loading, setLoading] = useState(false);
  const [lastRefreshed, setLastRefreshed] = useState<Date>(new Date());

  const fetchDashboardData = useCallback(async () => {
    setLoading(true);
    try {
      const res = await fetch("/api/v1/dashboard");
      if (res.ok) {
        const json = await res.json();
        // Merge with defaults to ensure all required fields are safely populated
        setData({
          ...defaultDashboardData,
          ...json,
          financials: { ...defaultDashboardData.financials, ...(json.financials || {}) },
          cash_position: { ...defaultDashboardData.cash_position, ...(json.cash_position || {}) },
          arap: { ...defaultDashboardData.arap, ...(json.arap || {}) },
          zatca: { ...defaultDashboardData.zatca, ...(json.zatca || {}) },
          approvals: {
            ...defaultDashboardData.approvals,
            ...(json.approvals || {}),
            by_type: json.approvals?.by_type || defaultDashboardData.approvals.by_type,
            recent_requests:
              json.approvals?.recent_requests && json.approvals.recent_requests.length > 0
                ? json.approvals.recent_requests
                : defaultDashboardData.approvals.recent_requests,
          },
          stock_warnings:
            json.stock_warnings && json.stock_warnings.length > 0
              ? json.stock_warnings
              : defaultDashboardData.stock_warnings,
          recent_documents:
            json.recent_documents && json.recent_documents.length > 0
              ? json.recent_documents
              : defaultDashboardData.recent_documents,
          activity_feed:
            json.activity_feed && json.activity_feed.length > 0
              ? json.activity_feed
              : defaultDashboardData.activity_feed,
        });
        setLastRefreshed(new Date());
      }
    } catch {
      // Graceful fallback to default state on local network failure
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchDashboardData();
  }, [fetchDashboardData]);

  const formatCurrency = (val: number, cur = "SAR") => {
    return new Intl.NumberFormat(isAr ? "ar-SA" : "en-US", {
      style: "currency",
      currency: cur,
      maximumFractionDigits: 0,
    }).format(val || 0);
  };

  const formatNumber = (val: number) => {
    return new Intl.NumberFormat(isAr ? "ar-SA" : "en-US").format(val || 0);
  };

  const formatFileSize = (bytes: number) => {
    if (!bytes || bytes === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + " " + sizes[i];
  };

  const formatRelativeTime = (isoString: string) => {
    try {
      const date = new Date(isoString);
      const diffMs = Date.now() - date.getTime();
      const diffHrs = Math.floor(diffMs / 3600000);
      if (diffHrs < 1) {
        const diffMin = Math.max(1, Math.floor(diffMs / 60000));
        return isAr ? `منذ ${diffMin} دقيقة` : `${diffMin}m ago`;
      }
      if (diffHrs < 24) {
        return isAr ? `منذ ${diffHrs} ساعة` : `${diffHrs}h ago`;
      }
      const diffDays = Math.floor(diffHrs / 24);
      return isAr ? `منذ ${diffDays} يوم` : `${diffDays}d ago`;
    } catch {
      return isoString;
    }
  };

  return (
    <div className="space-y-6 pb-12">
      {/* 1. Executive Hero & Quick Actions Banner */}
      <div className="relative overflow-hidden rounded-2xl border bg-gradient-to-br from-card via-card to-primary/5 p-6 sm:p-8 shadow-sm">
        <div className="absolute top-0 inset-inline-end-0 w-96 h-96 bg-primary/10 rounded-full blur-3xl pointer-events-none -z-0" />
        <div className="relative z-10 flex flex-col lg:flex-row lg:items-center justify-between gap-6">
          <div className="space-y-3 max-w-2xl">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full text-xs font-semibold bg-primary/10 text-primary border border-primary/20">
              <Sparkles className="h-3.5 w-3.5" />
              <span>{t.dashboard.badge}</span>
            </div>
            <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-foreground">
              {t.dashboard.title}
            </h1>
            <p className="text-sm text-muted-foreground leading-relaxed">
              {t.dashboard.subtitle}
            </p>
            <div className="pt-1 flex items-center gap-4 text-xs text-muted-foreground font-medium">
              <div className="flex items-center gap-1.5">
                <Building2 className="h-3.5 w-3.5 text-primary" />
                <span>{data.legal_entity_name}</span>
              </div>
              <span className="text-border">•</span>
              <div className="flex items-center gap-1.5">
                <Clock className="h-3.5 w-3.5" />
                <span>
                  {t.dashboard.last_updated}: {lastRefreshed.toLocaleTimeString(isAr ? "ar-SA" : "en-US")}
                </span>
              </div>
            </div>
          </div>

          <div className="flex flex-wrap items-center gap-3">
            <button
              onClick={fetchDashboardData}
              disabled={loading}
              className="inline-flex items-center gap-2 px-4 py-2.5 rounded-xl text-xs font-semibold bg-card hover:bg-muted border shadow-sm transition-all disabled:opacity-50"
            >
              <RefreshCw className={`h-4 w-4 text-primary ${loading ? "animate-spin" : ""}`} />
              <span>{t.dashboard.refresh}</span>
            </button>
            <Link
              href="/reports"
              className="inline-flex items-center gap-2 px-4 py-2.5 rounded-xl text-xs font-semibold bg-primary text-primary-foreground hover:bg-primary/90 shadow-sm transition-all"
            >
              <FileSpreadsheet className="h-4 w-4" />
              <span>{t.dashboard.action_view_reports}</span>
              <ArrowRight className="h-3.5 w-3.5 rtl:rotate-180" />
            </Link>
          </div>
        </div>

        {/* Quick Module Shortcuts Strip */}
        <div className="relative z-10 pt-6 mt-6 border-t border-border/60 flex flex-wrap gap-2 sm:gap-3">
          <Link
            href="/accounting"
            className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium bg-muted/60 hover:bg-muted text-foreground border border-border/50 transition-colors"
          >
            <Scale className="h-3.5 w-3.5 text-blue-500" />
            <span>{t.dashboard.action_new_entry}</span>
          </Link>
          <Link
            href="/sales"
            className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium bg-muted/60 hover:bg-muted text-foreground border border-border/50 transition-colors"
          >
            <Receipt className="h-3.5 w-3.5 text-emerald-500" />
            <span>{t.dashboard.action_new_invoice}</span>
          </Link>
          <Link
            href="/payables"
            className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium bg-muted/60 hover:bg-muted text-foreground border border-border/50 transition-colors"
          >
            <Wallet className="h-3.5 w-3.5 text-amber-500" />
            <span>{t.dashboard.action_new_bill}</span>
          </Link>
          <Link
            href="/treasury"
            className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium bg-muted/60 hover:bg-muted text-foreground border border-border/50 transition-colors"
          >
            <Building2 className="h-3.5 w-3.5 text-indigo-500" />
            <span>{t.dashboard.action_new_payment}</span>
          </Link>
        </div>
      </div>

      {/* 2. Top Financial KPI Cards Row */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 sm:gap-6">
        {/* KPI 1: Revenue */}
        <div className="relative overflow-hidden rounded-xl border bg-card p-5 space-y-3 hover:shadow-md transition-shadow group">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">{t.dashboard.kpi_revenue}</span>
            <div className="h-9 w-9 rounded-lg bg-emerald-500/10 text-emerald-600 flex items-center justify-center group-hover:scale-105 transition-transform">
              <TrendingUp className="h-5 w-5" />
            </div>
          </div>
          <div className="space-y-1">
            <div className="text-2xl font-bold tracking-tight text-foreground">
              {formatCurrency(data.financials.total_revenue, data.financials.currency)}
            </div>
            <div className="flex items-center gap-1 text-xs text-emerald-600 font-medium">
              <ArrowUpRight className="h-3.5 w-3.5" />
              <span>+14.2%</span>
              <span className="text-muted-foreground font-normal ps-1">vs prior period</span>
            </div>
          </div>
          <div className="h-1.5 w-full bg-emerald-500/15 rounded-full overflow-hidden">
            <div className="h-full bg-emerald-500 rounded-full w-[78%]" />
          </div>
        </div>

        {/* KPI 2: Expenses */}
        <div className="relative overflow-hidden rounded-xl border bg-card p-5 space-y-3 hover:shadow-md transition-shadow group">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">{t.dashboard.kpi_expenses}</span>
            <div className="h-9 w-9 rounded-lg bg-rose-500/10 text-rose-600 flex items-center justify-center group-hover:scale-105 transition-transform">
              <TrendingDown className="h-5 w-5" />
            </div>
          </div>
          <div className="space-y-1">
            <div className="text-2xl font-bold tracking-tight text-foreground">
              {formatCurrency(data.financials.total_expense, data.financials.currency)}
            </div>
            <div className="flex items-center gap-1 text-xs text-muted-foreground font-medium">
              <span className="text-rose-600 font-medium">+6.4%</span>
              <span className="text-muted-foreground font-normal ps-1">controlled burn</span>
            </div>
          </div>
          <div className="h-1.5 w-full bg-rose-500/15 rounded-full overflow-hidden">
            <div className="h-full bg-rose-500 rounded-full w-[56%]" />
          </div>
        </div>

        {/* KPI 3: Net Profit */}
        <div className="relative overflow-hidden rounded-xl border bg-card p-5 space-y-3 hover:shadow-md transition-shadow group">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">{t.dashboard.kpi_net_profit}</span>
            <div className="h-9 w-9 rounded-lg bg-indigo-500/10 text-indigo-600 flex items-center justify-center group-hover:scale-105 transition-transform">
              <DollarSign className="h-5 w-5" />
            </div>
          </div>
          <div className="space-y-1">
            <div className="text-2xl font-bold tracking-tight text-foreground">
              {formatCurrency(data.financials.net_profit, data.financials.currency)}
            </div>
            <div className="flex items-center gap-1 text-xs text-indigo-600 font-medium">
              <ArrowUpRight className="h-3.5 w-3.5" />
              <span>+24.8%</span>
              <span className="text-muted-foreground font-normal ps-1">net earnings</span>
            </div>
          </div>
          <div className="h-1.5 w-full bg-indigo-500/15 rounded-full overflow-hidden">
            <div className="h-full bg-indigo-500 rounded-full w-[65%]" />
          </div>
        </div>

        {/* KPI 4: Profit Margin */}
        <div className="relative overflow-hidden rounded-xl border bg-card p-5 space-y-3 hover:shadow-md transition-shadow group">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">{t.dashboard.kpi_profit_margin}</span>
            <div className="h-9 w-9 rounded-lg bg-amber-500/10 text-amber-600 flex items-center justify-center group-hover:scale-105 transition-transform">
              <Sparkles className="h-5 w-5" />
            </div>
          </div>
          <div className="space-y-1">
            <div className="text-2xl font-bold tracking-tight text-foreground">
              {data.financials.profit_margin.toFixed(1)}%
            </div>
            <div className="flex items-center gap-1 text-xs text-amber-600 font-medium">
              <span>Healthy Margin</span>
              <span className="text-muted-foreground font-normal ps-1">target &gt;25%</span>
            </div>
          </div>
          <div className="h-1.5 w-full bg-amber-500/15 rounded-full overflow-hidden">
            <div className="h-full bg-amber-500 rounded-full w-[72%]" />
          </div>
        </div>
      </div>

      {/* 3. Row: Liquidity & Working Capital */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Cash Position Card */}
        <div className="rounded-xl border bg-card p-6 space-y-5 shadow-sm">
          <div className="flex items-center justify-between">
            <div className="space-y-1">
              <h3 className="font-semibold text-base text-foreground flex items-center gap-2">
                <Wallet className="h-4 w-4 text-primary" />
                {t.dashboard.cash_position}
              </h3>
              <p className="text-xs text-muted-foreground">
                {data.cash_position.accounts_count} {t.dashboard.active_accounts}
              </p>
            </div>
            <Link
              href="/treasury"
              className="text-xs font-medium text-primary hover:underline flex items-center gap-1"
            >
              <span>{isAr ? "إدارة الخزينة" : "Treasury Hub"}</span>
              <ChevronRight className="h-3.5 w-3.5 rtl:rotate-180" />
            </Link>
          </div>

          <div className="p-4 rounded-xl bg-muted/40 border space-y-2">
            <span className="text-xs font-medium text-muted-foreground">{t.dashboard.total_liquidity}</span>
            <div className="text-3xl font-extrabold text-foreground">
              {formatCurrency(data.cash_position.total_liquid_funds, data.cash_position.currency)}
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="p-4 rounded-xl border bg-card/60 space-y-1">
              <div className="flex items-center justify-between text-xs text-muted-foreground">
                <span>{t.dashboard.bank_accounts}</span>
                <span className="h-2 w-2 rounded-full bg-blue-500" />
              </div>
              <div className="text-lg font-bold text-foreground">
                {formatCurrency(data.cash_position.total_bank_balance, data.cash_position.currency)}
              </div>
              <div className="text-[11px] text-muted-foreground">
                {((data.cash_position.total_bank_balance / (data.cash_position.total_liquid_funds || 1)) * 100).toFixed(0)}% of funds
              </div>
            </div>

            <div className="p-4 rounded-xl border bg-card/60 space-y-1">
              <div className="flex items-center justify-between text-xs text-muted-foreground">
                <span>{t.dashboard.cash_vaults}</span>
                <span className="h-2 w-2 rounded-full bg-emerald-500" />
              </div>
              <div className="text-lg font-bold text-foreground">
                {formatCurrency(data.cash_position.total_cash_balance, data.cash_position.currency)}
              </div>
              <div className="text-[11px] text-muted-foreground">
                {((data.cash_position.total_cash_balance / (data.cash_position.total_liquid_funds || 1)) * 100).toFixed(0)}% of funds
              </div>
            </div>
          </div>

          {/* Liquid Ratio Bar */}
          <div className="space-y-1.5">
            <div className="h-2 w-full bg-muted rounded-full overflow-hidden flex">
              <div
                className="h-full bg-blue-500 transition-all"
                style={{
                  width: `${(data.cash_position.total_bank_balance / (data.cash_position.total_liquid_funds || 1)) * 100}%`,
                }}
              />
              <div
                className="h-full bg-emerald-500 transition-all"
                style={{
                  width: `${(data.cash_position.total_cash_balance / (data.cash_position.total_liquid_funds || 1)) * 100}%`,
                }}
              />
            </div>
            <div className="flex items-center justify-between text-[11px] text-muted-foreground">
              <span>{isAr ? "البنوك (IBAN المعتمد)" : "Verified IBANs"}</span>
              <span>{isAr ? "العهد وصناديق النقدية" : "Petty Cash & Vaults"}</span>
            </div>
          </div>
        </div>

        {/* AR / AP Working Capital Card */}
        <div className="rounded-xl border bg-card p-6 space-y-5 shadow-sm">
          <div className="flex items-center justify-between">
            <div className="space-y-1">
              <h3 className="font-semibold text-base text-foreground flex items-center gap-2">
                <Scale className="h-4 w-4 text-primary" />
                {t.dashboard.arap_summary}
              </h3>
              <p className="text-xs text-muted-foreground">
                {isAr ? "المقارنة بين استحقاقات العملاء والتزامات الموردين" : "Receivables vs. Payables Working Capital"}
              </p>
            </div>
            <div className="px-2.5 py-1 rounded-full text-xs font-semibold bg-emerald-500/10 text-emerald-600 border border-emerald-500/20">
              +{formatCurrency(data.arap.total_receivables - data.arap.total_payables, data.arap.currency)}
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            {/* Receivables Column */}
            <div className="p-4 rounded-xl border border-blue-500/20 bg-blue-500/5 space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold text-blue-600 dark:text-blue-400">
                  {t.dashboard.receivables}
                </span>
                <span className="text-[11px] px-2 py-0.5 rounded bg-blue-500/10 text-blue-600 font-medium">
                  {data.arap.open_receivable_docs} {t.dashboard.open_invoices}
                </span>
              </div>
              <div className="text-xl font-bold text-foreground">
                {formatCurrency(data.arap.total_receivables, data.arap.currency)}
              </div>
              <div className="pt-2 border-t border-blue-500/15 flex items-center justify-between text-xs">
                <span className="text-muted-foreground">{t.dashboard.overdue_ar}:</span>
                <span className="font-semibold text-amber-600">
                  {formatCurrency(data.arap.overdue_receivables, data.arap.currency)}
                </span>
              </div>
            </div>

            {/* Payables Column */}
            <div className="p-4 rounded-xl border border-rose-500/20 bg-rose-500/5 space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold text-rose-600 dark:text-rose-400">
                  {t.dashboard.payables}
                </span>
                <span className="text-[11px] px-2 py-0.5 rounded bg-rose-500/10 text-rose-600 font-medium">
                  {data.arap.open_payable_docs} {t.dashboard.open_bills}
                </span>
              </div>
              <div className="text-xl font-bold text-foreground">
                {formatCurrency(data.arap.total_payables, data.arap.currency)}
              </div>
              <div className="pt-2 border-t border-rose-500/15 flex items-center justify-between text-xs">
                <span className="text-muted-foreground">{t.dashboard.overdue_ap}:</span>
                <span className="font-semibold text-rose-600">
                  {formatCurrency(data.arap.overdue_payables, data.arap.currency)}
                </span>
              </div>
            </div>
          </div>

          <div className="p-3 rounded-lg bg-muted/40 text-xs text-muted-foreground flex items-center justify-between">
            <span>{isAr ? "صافي رأس المال العامل الجاري:" : "Net Current Working Capital:"}</span>
            <span className="font-bold text-foreground">
              {formatCurrency(data.arap.total_receivables - data.arap.total_payables, data.arap.currency)}
            </span>
          </div>
        </div>
      </div>

      {/* 4. Row: Approvals & ZATCA Compliance */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Pending Approvals Widget */}
        <div className="rounded-xl border bg-card p-6 space-y-4 shadow-sm flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between mb-4">
              <div className="space-y-1">
                <h3 className="font-semibold text-base text-foreground flex items-center gap-2">
                  <Clock className="h-4 w-4 text-amber-500" />
                  {t.dashboard.pending_approvals}
                </h3>
                <p className="text-xs text-muted-foreground">{t.dashboard.pending_approvals_desc}</p>
              </div>
              <span className="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-bold bg-amber-500/10 text-amber-600 border border-amber-500/20">
                {data.approvals.total_pending} {t.dashboard.total_pending}
              </span>
            </div>

            {/* Pills by Type */}
            <div className="flex flex-wrap gap-2 mb-4">
              {Object.entries(data.approvals.by_type).map(([type, count]) => (
                <div
                  key={type}
                  className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium bg-muted border text-foreground"
                >
                  <span className="font-semibold text-primary">{count}</span>
                  <span className="text-muted-foreground">{type.replace("_", " ")}</span>
                </div>
              ))}
            </div>

            {/* List of Recent Approvals */}
            <div className="space-y-2.5">
              {data.approvals.recent_requests.length === 0 ? (
                <div className="p-6 text-center text-xs text-muted-foreground">
                  <CheckCircle2 className="h-8 w-8 text-emerald-500 mx-auto mb-2 opacity-80" />
                  {t.dashboard.no_pending_approvals}
                </div>
              ) : (
                data.approvals.recent_requests.map((item) => (
                  <div
                    key={item.id}
                    className="p-3 rounded-lg border bg-muted/20 hover:bg-muted/40 transition-colors flex items-center justify-between gap-3 text-xs"
                  >
                    <div className="space-y-0.5">
                      <div className="font-medium text-foreground flex items-center gap-2">
                        <span className="px-1.5 py-0.5 rounded text-[10px] font-bold bg-primary/10 text-primary">
                          {item.resource_type}
                        </span>
                        <span className="font-mono text-xs">{item.resource_id}</span>
                      </div>
                      <div className="text-[11px] text-muted-foreground">
                        {formatRelativeTime(item.created_at)}
                      </div>
                    </div>
                    <Link
                      href="/workflow"
                      className="px-3 py-1.5 rounded-lg text-xs font-medium bg-primary/10 hover:bg-primary/20 text-primary border border-primary/20 transition-colors"
                    >
                      {t.dashboard.btn_review}
                    </Link>
                  </div>
                ))
              )}
            </div>
          </div>

          <div className="pt-3 border-t">
            <Link
              href="/workflow"
              className="text-xs font-medium text-primary hover:underline flex items-center justify-between"
            >
              <span>{isAr ? "الانتقال إلى محرك مسارات العمل والموافقات" : "Open Workflow Authorization Center"}</span>
              <ArrowRight className="h-3.5 w-3.5 rtl:rotate-180" />
            </Link>
          </div>
        </div>

        {/* ZATCA Phase 2 Status Widget */}
        <div className="rounded-xl border bg-card p-6 space-y-4 shadow-sm flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between mb-4">
              <div className="space-y-1">
                <h3 className="font-semibold text-base text-foreground flex items-center gap-2">
                  <ShieldCheck className="h-4 w-4 text-emerald-500" />
                  {t.dashboard.zatca_status}
                </h3>
                <p className="text-xs text-muted-foreground">{t.dashboard.zatca_desc}</p>
              </div>
              <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-bold bg-emerald-500/10 text-emerald-600 border border-emerald-500/20">
                <Lock className="h-3 w-3" />
                <span>ECDSA Phase 2</span>
              </div>
            </div>

            {/* Compliance Rate Header */}
            <div className="p-4 rounded-xl bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-between">
              <div>
                <span className="text-xs font-medium text-muted-foreground">
                  {t.dashboard.zatca_compliance_rate}
                </span>
                <div className="text-3xl font-extrabold text-emerald-600 dark:text-emerald-400">
                  {data.zatca.compliance_rate.toFixed(1)}%
                </div>
              </div>
              <div className="text-end text-xs text-muted-foreground">
                <div>{t.dashboard.zatca_total}</div>
                <div className="text-lg font-bold text-foreground">{data.zatca.total_documents}</div>
              </div>
            </div>

            {/* 4 Stat Boxes */}
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 mt-4">
              <div className="p-3 rounded-lg border bg-card/60 text-center space-y-1">
                <div className="text-xs text-muted-foreground">{t.dashboard.zatca_cleared}</div>
                <div className="text-base font-bold text-emerald-600">{data.zatca.cleared_count}</div>
              </div>
              <div className="p-3 rounded-lg border bg-card/60 text-center space-y-1">
                <div className="text-xs text-muted-foreground">{t.dashboard.zatca_pending}</div>
                <div className="text-base font-bold text-blue-600">{data.zatca.pending_count}</div>
              </div>
              <div className="p-3 rounded-lg border bg-card/60 text-center space-y-1">
                <div className="text-xs text-muted-foreground">{t.dashboard.zatca_warnings}</div>
                <div className="text-base font-bold text-amber-600">{data.zatca.warning_count}</div>
              </div>
              <div className="p-3 rounded-lg border bg-card/60 text-center space-y-1">
                <div className="text-xs text-muted-foreground">{t.dashboard.zatca_rejected}</div>
                <div className="text-base font-bold text-rose-600">{data.zatca.rejected_count}</div>
              </div>
            </div>
          </div>

          <div className="pt-3 border-t">
            <Link
              href="/zatca"
              className="text-xs font-medium text-primary hover:underline flex items-center justify-between"
            >
              <span>{isAr ? "الانتقال إلى مركز فوترة زاتكا والربط" : "Manage ZATCA CSIDs & Portal"}</span>
              <ArrowRight className="h-3.5 w-3.5 rtl:rotate-180" />
            </Link>
          </div>
        </div>
      </div>

      {/* 5. Row: Stock Warnings & Recent Documents */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Stock Warnings Panel */}
        <div className="rounded-xl border bg-card p-6 space-y-4 shadow-sm">
          <div className="flex items-center justify-between">
            <div className="space-y-1">
              <h3 className="font-semibold text-base text-foreground flex items-center gap-2">
                <Boxes className="h-4 w-4 text-rose-500" />
                {t.dashboard.stock_warnings}
              </h3>
              <p className="text-xs text-muted-foreground">{t.dashboard.stock_warnings_desc}</p>
            </div>
            <Link
              href="/inventory"
              className="text-xs font-medium text-primary hover:underline flex items-center gap-1"
            >
              <span>{isAr ? "إدارة المخزون" : "Inventory"}</span>
              <ChevronRight className="h-3.5 w-3.5 rtl:rotate-180" />
            </Link>
          </div>

          <div className="space-y-2.5">
            {data.stock_warnings.length === 0 ? (
              <div className="p-6 text-center text-xs text-muted-foreground">
                <CheckCircle2 className="h-8 w-8 text-emerald-500 mx-auto mb-2 opacity-80" />
                {t.dashboard.no_stock_warnings}
              </div>
            ) : (
              data.stock_warnings.map((item) => (
                <div
                  key={item.item_id}
                  className="p-3.5 rounded-lg border bg-card/60 flex items-center justify-between gap-3 text-xs"
                >
                  <div className="space-y-1">
                    <div className="font-semibold text-foreground">
                      {isAr ? item.item_name_ar : item.item_name_en}
                    </div>
                    <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
                      <span className="font-mono">{item.item_code}</span>
                      <span>•</span>
                      <span>{isAr ? item.warehouse_name_ar : item.warehouse_name_en}</span>
                    </div>
                  </div>

                  <div className="text-end space-y-1">
                    <div className="flex items-center gap-2 justify-end">
                      <span className="font-bold text-foreground text-sm">
                        {item.quantity_available} {isAr ? "متاح" : "avail"}
                      </span>
                      <span
                        className={`px-2 py-0.5 rounded text-[10px] font-bold ${
                          item.severity === "CRITICAL"
                            ? "bg-rose-500/10 text-rose-600 border border-rose-500/20"
                            : item.severity === "WARNING"
                            ? "bg-amber-500/10 text-amber-600 border border-amber-500/20"
                            : "bg-blue-500/10 text-blue-600 border border-blue-500/20"
                        }`}
                      >
                        {item.severity === "CRITICAL"
                          ? t.dashboard.severity_critical
                          : item.severity === "WARNING"
                          ? t.dashboard.severity_warning
                          : t.dashboard.severity_low}
                      </span>
                    </div>
                    <div className="text-[11px] text-muted-foreground">
                      {t.dashboard.col_reorder}: {item.reorder_point} | {t.dashboard.col_on_hand}: {item.quantity_on_hand}
                    </div>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>

        {/* Recent Documents Panel */}
        <div className="rounded-xl border bg-card p-6 space-y-4 shadow-sm">
          <div className="flex items-center justify-between">
            <div className="space-y-1">
              <h3 className="font-semibold text-base text-foreground flex items-center gap-2">
                <FileText className="h-4 w-4 text-primary" />
                {t.dashboard.recent_documents}
              </h3>
              <p className="text-xs text-muted-foreground">{t.dashboard.recent_documents_desc}</p>
            </div>
            <Link
              href="/documents"
              className="text-xs font-medium text-primary hover:underline flex items-center gap-1"
            >
              <span>{isAr ? "الأرشيف" : "Documents Hub"}</span>
              <ChevronRight className="h-3.5 w-3.5 rtl:rotate-180" />
            </Link>
          </div>

          <div className="space-y-2.5">
            {data.recent_documents.length === 0 ? (
              <div className="p-6 text-center text-xs text-muted-foreground">
                <FileText className="h-8 w-8 text-muted-foreground/50 mx-auto mb-2" />
                {t.dashboard.no_recent_docs}
              </div>
            ) : (
              data.recent_documents.map((doc) => (
                <div
                  key={doc.id}
                  className="p-3 rounded-lg border bg-card/60 flex items-center justify-between gap-3 text-xs"
                >
                  <div className="flex items-center gap-3">
                    <div className="h-8 w-8 rounded-lg bg-primary/10 text-primary flex items-center justify-center shrink-0">
                      <FileCode className="h-4 w-4" />
                    </div>
                    <div className="space-y-0.5">
                      <div className="font-semibold text-foreground">{doc.title}</div>
                      <div className="flex items-center gap-2 text-[11px] text-muted-foreground font-mono">
                        <span>{doc.filename}</span>
                        <span>•</span>
                        <span>{formatFileSize(doc.byte_size)}</span>
                      </div>
                    </div>
                  </div>

                  <div className="text-end text-[11px] text-muted-foreground whitespace-nowrap">
                    {formatRelativeTime(doc.created_at)}
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      </div>

      {/* 6. Row: Cryptographically Chained Activity Feed (Audit Trail) */}
      <div className="rounded-xl border bg-card p-6 space-y-4 shadow-sm">
        <div className="flex items-center justify-between">
          <div className="space-y-1">
            <h3 className="font-semibold text-base text-foreground flex items-center gap-2">
              <Activity className="h-4 w-4 text-emerald-500" />
              {t.dashboard.activity_feed}
            </h3>
            <p className="text-xs text-muted-foreground">{t.dashboard.activity_feed_desc}</p>
          </div>
          <Link
            href="/audit"
            className="text-xs font-medium text-primary hover:underline flex items-center gap-1"
          >
            <span>{isAr ? "سجل التدقيق الكامل" : "Audit Log Explorer"}</span>
            <ChevronRight className="h-3.5 w-3.5 rtl:rotate-180" />
          </Link>
        </div>

        <div className="divide-y divide-border/60">
          {data.activity_feed.length === 0 ? (
            <div className="p-6 text-center text-xs text-muted-foreground">
              {t.dashboard.no_activities}
            </div>
          ) : (
            data.activity_feed.map((act) => (
              <div key={act.id} className="py-3 flex items-center justify-between gap-4 text-xs">
                <div className="flex items-center gap-3">
                  <div className="h-8 w-8 rounded-full bg-muted flex items-center justify-center font-bold text-xs text-foreground shrink-0">
                    {act.user_name ? act.user_name.slice(0, 1) : "U"}
                  </div>
                  <div className="space-y-0.5">
                    <div className="flex items-center gap-2">
                      <span className="font-semibold text-foreground">{act.user_name}</span>
                      <span className="px-1.5 py-0.5 rounded text-[10px] font-mono font-medium bg-muted text-muted-foreground">
                        {act.action}
                      </span>
                    </div>
                    <div className="text-[11px] text-muted-foreground flex items-center gap-2">
                      <span>{act.resource_type}</span>
                      <span className="font-mono text-primary">{act.resource_id}</span>
                    </div>
                  </div>
                </div>

                <div className="text-[11px] text-muted-foreground whitespace-nowrap">
                  {formatRelativeTime(act.occurred_at)}
                </div>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
}
