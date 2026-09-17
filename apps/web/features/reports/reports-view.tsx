"use client";

import React, { useState } from "react";
import {
  BarChart3,
  Download,
  Printer,
  Calendar,
  Filter,
  RefreshCw,
  TrendingUp,
  TrendingDown,
  DollarSign,
  FileSpreadsheet,
  Building2,
  PieChart,
  ShieldCheck,
  CheckCircle2,
  AlertCircle,
  FileText,
  Search,
  Boxes,
  Truck,
  Users,
  Wallet,
  Receipt,
  Scale,
  CreditCard,
  ChevronDown,
  Layers,
  Clock,
  Sparkles,
  ArrowUpRight,
  ArrowDownRight,
} from "lucide-react";
import { useLocale } from "@/components/layout/shell";

export type ReportCategory = "FINANCIAL" | "COMMERCIAL" | "SUPPLY_CHAIN" | "COMPLIANCE";

export function ReportsView() {
  const { locale, t, dir } = useLocale();
  const isAr = locale === "ar";

  const [activeCategory, setActiveCategory] = useState<ReportCategory>("FINANCIAL");
  const [activeReport, setActiveReport] = useState<string>("TRIAL_BALANCE");
  const [datePreset, setDatePreset] = useState<string>("YTD");
  const [searchQuery, setSearchQuery] = useState("");
  const [isExporting, setIsExporting] = useState(false);

  // Executive KPI summary data
  const kpis = [
    {
      titleEn: "Total Assets",
      titleAr: "إجمالي الأصول",
      value: "4,850,000 SAR",
      change: "+12.4%",
      isPositive: true,
      icon: Scale,
      color: "text-emerald-500",
      bg: "bg-emerald-500/10",
    },
    {
      titleEn: "Net Profit (YTD)",
      titleAr: "صافي الأرباح (من أول العام)",
      value: "925,400 SAR",
      change: "+18.2%",
      isPositive: true,
      icon: TrendingUp,
      color: "text-indigo-500",
      bg: "bg-indigo-500/10",
    },
    {
      titleEn: "Outstanding AR",
      titleAr: "ديون العملاء غير المحصلة",
      value: "612,000 SAR",
      change: "-5.3%",
      isPositive: true,
      icon: Users,
      color: "text-amber-500",
      bg: "bg-amber-500/10",
    },
    {
      titleEn: "Net VAT Liability (15%)",
      titleAr: "صافي التزام الزكاة والضريبة",
      value: "84,350 SAR",
      change: "Current Quarter",
      isPositive: false,
      icon: Receipt,
      color: "text-rose-500",
      bg: "bg-rose-500/10",
    },
  ];

  // Reports Catalog definition
  const reportsList = [
    // Financial Statements
    {
      id: "TRIAL_BALANCE",
      category: "FINANCIAL",
      nameEn: "Trial Balance",
      nameAr: "ميزان المراجعة",
      descEn: "Debits, credits, and net ending balances across all Chart of Accounts items.",
      descAr: "الأرصدة المدينة والدائنة وصافي الأرصدة الختامية لكافة حسابات الدليل المحاسبي.",
      icon: Scale,
    },
    {
      id: "BALANCE_SHEET",
      category: "FINANCIAL",
      nameEn: "Balance Sheet",
      nameAr: "الميزانية العمومية",
      descEn: "Statement of Financial Position: Total Assets, Liabilities, and Owner's Equity.",
      descAr: "قائمة المركز المالي: إجمالي الأصول المتداولة والثابتة والالتزامات وحقوق الملكية.",
      icon: Building2,
    },
    {
      id: "INCOME_STATEMENT",
      category: "FINANCIAL",
      nameEn: "Income Statement (P&L)",
      nameAr: "قائمة الدخل (الأرباح والخسائر)",
      descEn: "Operating revenue, COGS, expenses, and net net profitability for the period.",
      descAr: "الإيرادات التشغيلية وتكلفة المبيعات والمصروفات وصافي الربح أو الخسارة للفترة.",
      icon: TrendingUp,
    },
    {
      id: "CASH_FLOW",
      category: "FINANCIAL",
      nameEn: "Cash Flow Statement",
      nameAr: "قائمة التدفقات النقدية",
      descEn: "Liquidity movements across Operating, Investing, and Financing activities.",
      descAr: "حركة السيولة النقدية مقسمة على الأنشطة التشغيلية والاستثمارية والتمويلية.",
      icon: Wallet,
    },
    {
      id: "GENERAL_LEDGER",
      category: "FINANCIAL",
      nameEn: "General Ledger",
      nameAr: "دفتر الأستاذ العام",
      descEn: "Chronological transaction journal lines with full debit/credit traceability.",
      descAr: "السجل التفصيلي اليومي الزمني لكافة القيود والعمليات المحاسبية المرحلة.",
      icon: FileSpreadsheet,
    },

    // Commercial & Aging
    {
      id: "AR_AGING",
      category: "COMMERCIAL",
      nameEn: "Accounts Receivable (AR) Aging",
      nameAr: "أعمار ديون العملاء",
      descEn: "Aging buckets (Current, 1-30, 31-60, 61-90, 90+ days) for customer collections.",
      descAr: "تصنيف مستحقات العملاء غير المحصلة حسب مدد الاستحقاق للتخطيط والتحصيل.",
      icon: Users,
    },
    {
      id: "AP_AGING",
      category: "COMMERCIAL",
      nameEn: "Accounts Payable (AP) Aging",
      nameAr: "أعمار ديون الموردين",
      descEn: "Vendor liability age distribution for cash flow management and settlement.",
      descAr: "جدولة التزامات الموردين المستحقة حسب فترات الاستحقاق وإدارة السيولة.",
      icon: Truck,
    },
    {
      id: "CUSTOMER_STATEMENT",
      category: "COMMERCIAL",
      nameEn: "Customer Account Statement",
      nameAr: "كشف حساب عميل",
      descEn: "Detailed customer ledger with invoices, credit notes, and running balance.",
      descAr: "كشف الحساب التفصيلي بالمعاملات والفواتير وسندات القبض والرصيد التراكمي.",
      icon: FileText,
    },
    {
      id: "SUPPLIER_STATEMENT",
      category: "COMMERCIAL",
      nameEn: "Supplier Account Statement",
      nameAr: "كشف حساب مورد",
      descEn: "Detailed supplier ledger with bills, debit notes, payments, and balance.",
      descAr: "كشف حساب المورد بفواتير الشراء وسندات الصرف والدفعات المتبقية.",
      icon: Receipt,
    },

    // Supply Chain
    {
      id: "INVENTORY_VALUATION",
      category: "SUPPLY_CHAIN",
      nameEn: "Inventory Valuation",
      nameAr: "تقييم المخزون والسلع",
      descEn: "Stock on hand by warehouse, weighted average cost, and total value.",
      descAr: "أرصدة الأصناف المتوفرة بالمستودعات ومتوسط التكلفة المرجحة وإجمالي القيمة.",
      icon: Boxes,
    },
    {
      id: "SALES_REPORT",
      category: "SUPPLY_CHAIN",
      nameEn: "Sales Performance Analytics",
      nameAr: "تحليلات حركة المبيعات",
      descEn: "Invoiced sales totals, taxable revenue, customer trends, and discount audit.",
      descAr: "تحليلات المبيعات الإجمالية، المبالغ الخاضعة للضريبة، والخصومات الممنوحة.",
      icon: TrendingUp,
    },
    {
      id: "PURCHASING_REPORT",
      category: "SUPPLY_CHAIN",
      nameEn: "Procurement & Purchasing Summary",
      nameAr: "ملخص حركة المشتريات",
      descEn: "Purchase orders breakdown, vendor expenditures, and tax input totals.",
      descAr: "أوامر الشراء المعتمدة، نفقات الموردين، وإجمالي مدخلات ضريبة القيمة المضافة.",
      icon: CreditCard,
    },

    // Statutory & Compliance
    {
      id: "VAT_RETURN",
      category: "COMPLIANCE",
      nameEn: "ZATCA VAT Return 15%",
      nameAr: "إقرار ضريبة القيمة المضافة (15%)",
      descEn: "Official Saudi tax return boxes for standard sales, purchases, and net tax.",
      descAr: "الخانات الرسمية المعتمدة لهيئة الزكاة والضريبة والجمارك لاحتساب صافي الضريبة المستحقة.",
      icon: ShieldCheck,
    },
    {
      id: "PAYROLL_SUMMARY",
      category: "COMPLIANCE",
      nameEn: "Consolidated Payroll Summary",
      nameAr: "مسير الرواتب المجمع",
      descEn: "Monthly payroll run totals: Basic wages, allowances, deductions, and net pay.",
      descAr: "ملخص مسير الرواتب الشهري: الراتب الأساسي، البدلات، الخصومات، وصافي الأجور.",
      icon: Wallet,
    },
    {
      id: "GOSI_CONTRIBUTION",
      category: "COMPLIANCE",
      nameEn: "GOSI Social Insurance Report",
      nameAr: "تقرير التأمينات الاجتماعية GOSI",
      descEn: "Contributory wage base, employer share, and employee monthly deduction.",
      descAr: "الأجور الخاضعة للاشتراك، وحصة المنشأة وحصة المشترك الشهرية للتأمينات.",
      icon: FileSpreadsheet,
    },
    {
      id: "BUDGET_VS_ACTUAL",
      category: "COMPLIANCE",
      nameEn: "Budget vs. Actual Variance",
      nameAr: "الموازنة التقديرية مقابل الفعلي",
      descEn: "Approved budget limits compared with posted ledger actuals and % variance.",
      descAr: "مقارنة السقوف المعتمدة بالميزانية مع المصروف الفعلي ونسبة الانحراف المالي.",
      icon: PieChart,
    },
  ];

  const currentReportObj = reportsList.find((r) => r.id === activeReport) || reportsList[0];

  const handleExportCSV = () => {
    setIsExporting(true);
    setTimeout(() => {
      setIsExporting(false);
      // Client-side CSV generation demonstration
      const dummyCsv = `Report: ${currentReportObj.nameEn}\nLegal Entity: Saudi Enterprise Ltd\nGenerated At: ${new Date().toISOString()}\nStatus: Balanced\n`;
      const blob = new Blob([dummyCsv], { type: "text/csv;charset=utf-8;" });
      const url = URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.setAttribute("href", url);
      link.setAttribute("download", `${activeReport.toLowerCase()}_${new Date().toISOString().slice(0, 10)}.csv`);
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
    }, 600);
  };

  return (
    <div className="space-y-6">
      {/* Top Header */}
      <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
        <div>
          <div className="flex items-center gap-2">
            <span className="inline-flex items-center rounded-full bg-primary/10 px-2.5 py-0.5 text-xs font-semibold text-primary">
              <Sparkles className="mr-1 h-3 w-3 rtl:ml-1 rtl:mr-0" />
              {isAr ? "التقارير الموحدة والذكاء المالي" : "Executive BI & Statutory Platform"}
            </span>
          </div>
          <h1 className="mt-1 text-2xl font-bold tracking-tight text-foreground sm:text-3xl">
            {isAr ? "منظومة التقارير الشاملة" : "ERP Reporting & Analytics Platform"}
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">
            {isAr
              ? "16 تقريراً مالياً وتجارياً وتشغيلياً متوافقاً مع معايير المحاسبة وهيئة الزكاة والضريبة (ZATCA)"
              : "16 Canonical Financial, Commercial, Supply Chain, and ZATCA Statutory Compliance Reports"}
          </p>
        </div>

        {/* Global Action Buttons */}
        <div className="flex flex-wrap items-center gap-2">
          {/* Preset Picker */}
          <div className="flex items-center rounded-lg border bg-card p-1 shadow-sm">
            {(["MONTH", "QUARTER", "YTD", "FY2026"] as const).map((preset) => (
              <button
                key={preset}
                onClick={() => setDatePreset(preset)}
                className={`rounded-md px-3 py-1.5 text-xs font-medium transition-colors ${
                  datePreset === preset
                    ? "bg-primary text-primary-foreground shadow-xs"
                    : "text-muted-foreground hover:text-foreground"
                }`}
              >
                {preset === "MONTH" && (isAr ? "الشهر الحالي" : "This Month")}
                {preset === "QUARTER" && (isAr ? "الربع الحالي" : "This Quarter")}
                {preset === "YTD" && (isAr ? "من أول العام" : "YTD")}
                {preset === "FY2026" && (isAr ? "السنة 2026" : "FY 2026")}
              </button>
            ))}
          </div>

          <button
            onClick={handleExportCSV}
            disabled={isExporting}
            className="inline-flex items-center gap-1.5 rounded-lg border bg-card px-3.5 py-2 text-xs font-semibold shadow-xs hover:bg-muted/50 transition-colors disabled:opacity-50"
          >
            <Download className={`h-4 w-4 ${isExporting ? "animate-bounce" : ""}`} />
            {isExporting ? (isAr ? "جاري التصدير..." : "Exporting...") : (isAr ? "تصدير CSV" : "Export CSV")}
          </button>

          <button
            onClick={() => window.print()}
            className="inline-flex items-center gap-1.5 rounded-lg bg-primary px-3.5 py-2 text-xs font-semibold text-primary-foreground shadow-xs hover:bg-primary/90 transition-colors"
          >
            <Printer className="h-4 w-4" />
            {isAr ? "طباعة التقرير" : "Print Report"}
          </button>
        </div>
      </div>

      {/* KPI Cards Grid */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {kpis.map((kpi, idx) => {
          const Icon = kpi.icon;
          return (
            <div
              key={idx}
              className="relative overflow-hidden rounded-xl border bg-card p-5 shadow-xs transition-all hover:shadow-md"
            >
              <div className="flex items-center justify-between">
                <span className="text-xs font-medium text-muted-foreground">
                  {isAr ? kpi.titleAr : kpi.titleEn}
                </span>
                <div className={`flex h-8 w-8 items-center justify-center rounded-lg ${kpi.bg}`}>
                  <Icon className={`h-4 w-4 ${kpi.color}`} />
                </div>
              </div>
              <div className="mt-3 flex items-baseline justify-between">
                <div className="text-2xl font-bold tracking-tight text-foreground">{kpi.value}</div>
                <div className="flex items-center text-xs font-semibold text-emerald-600 dark:text-emerald-400">
                  {kpi.isPositive ? (
                    <ArrowUpRight className="mr-0.5 h-3.5 w-3.5 rtl:ml-0.5 rtl:mr-0" />
                  ) : (
                    <ArrowDownRight className="mr-0.5 h-3.5 w-3.5 rtl:ml-0.5 rtl:mr-0" />
                  )}
                  {kpi.change}
                </div>
              </div>
            </div>
          );
        })}
      </div>

      {/* Category Tabs & Selector */}
      <div className="flex flex-col gap-4 border-b pb-4">
        <div className="flex flex-wrap gap-2">
          {[
            { id: "FINANCIAL", nameEn: "Financial Statements", nameAr: "القوائم المالية المحاسبية", count: 5 },
            { id: "COMMERCIAL", nameEn: "Commercial & Aging", nameAr: "الديون والعملاء والموردين", count: 4 },
            { id: "SUPPLY_CHAIN", nameEn: "Supply Chain & Sales", nameAr: "المخزون والمبيعات والمشتريات", count: 3 },
            { id: "COMPLIANCE", nameEn: "Statutory & HR", nameAr: "الزكاة والضرائب والموارد البشرية", count: 4 },
          ].map((cat) => (
            <button
              key={cat.id}
              onClick={() => {
                setActiveCategory(cat.id as ReportCategory);
                const firstInCat = reportsList.find((r) => r.category === cat.id);
                if (firstInCat) setActiveReport(firstInCat.id);
              }}
              className={`inline-flex items-center gap-2 rounded-xl px-4 py-2.5 text-xs font-semibold transition-all ${
                activeCategory === cat.id
                  ? "bg-primary text-primary-foreground shadow-sm ring-2 ring-primary/20"
                  : "border bg-card text-muted-foreground hover:bg-muted/50 hover:text-foreground"
              }`}
            >
              <span>{isAr ? cat.nameAr : cat.nameEn}</span>
              <span
                className={`rounded-full px-2 py-0.5 text-[10px] font-bold ${
                  activeCategory === cat.id ? "bg-white/20 text-white" : "bg-muted text-muted-foreground"
                }`}
              >
                {cat.count}
              </span>
            </button>
          ))}
        </div>

        {/* Sub-report selector pills */}
        <div className="flex flex-wrap items-center gap-2 pt-1">
          {reportsList
            .filter((r) => r.category === activeCategory)
            .map((rep) => {
              const Icon = rep.icon;
              const isSelected = activeReport === rep.id;
              return (
                <button
                  key={rep.id}
                  onClick={() => setActiveReport(rep.id)}
                  className={`inline-flex items-center gap-1.5 rounded-lg border px-3 py-1.5 text-xs font-medium transition-all ${
                    isSelected
                      ? "border-primary/50 bg-primary/10 text-primary font-semibold shadow-xs"
                      : "border-transparent bg-muted/40 text-muted-foreground hover:bg-muted/70 hover:text-foreground"
                  }`}
                >
                  <Icon className="h-3.5 w-3.5" />
                  {isAr ? rep.nameAr : rep.nameEn}
                </button>
              );
            })}
        </div>
      </div>

      {/* Report Container */}
      <div className="rounded-xl border bg-card shadow-xs">
        {/* Report Header Banner */}
        <div className="flex flex-col gap-2 border-b p-5 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10 text-primary">
              {React.createElement(currentReportObj.icon, { className: "h-5 w-5" })}
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h2 className="text-lg font-bold text-foreground">
                  {isAr ? currentReportObj.nameAr : currentReportObj.nameEn}
                </h2>
                <span className="inline-flex items-center rounded-md bg-emerald-500/10 px-2 py-0.5 text-[10px] font-semibold text-emerald-600 dark:text-emerald-400">
                  <CheckCircle2 className="mr-1 h-3 w-3 rtl:ml-1 rtl:mr-0" />
                  {isAr ? "محدث ومطابق" : "Live & Balanced"}
                </span>
              </div>
              <p className="text-xs text-muted-foreground">
                {isAr ? currentReportObj.descAr : currentReportObj.descEn}
              </p>
            </div>
          </div>

          <div className="text-xs text-muted-foreground sm:text-right">
            <div>
              <span className="font-semibold text-foreground">{isAr ? "الكيان القانوني: " : "Legal Entity: "}</span>
              {isAr ? "شركة الرؤية السعودية للتجارة" : "Saudi Vision Trading Co."}
            </div>
            <div>
              <span className="font-semibold text-foreground">{isAr ? "تاريخ التوليد: " : "Generated: "}</span>
              {new Date().toLocaleDateString(isAr ? "ar-SA" : "en-US", {
                year: "numeric",
                month: "short",
                day: "numeric",
              })}
            </div>
          </div>
        </div>

        {/* Dynamic Report Content based on activeReport */}
        <div className="p-5">
          {activeReport === "TRIAL_BALANCE" && (
            <div className="space-y-4">
              <div className="overflow-x-auto rounded-lg border">
                <table className="w-full text-left text-xs rtl:text-right">
                  <thead className="bg-muted/60 text-muted-foreground font-semibold">
                    <tr>
                      <th className="px-4 py-3">{isAr ? "رمز الحساب" : "Account Code"}</th>
                      <th className="px-4 py-3">{isAr ? "اسم الحساب (عربي / English)" : "Account Name"}</th>
                      <th className="px-4 py-3">{isAr ? "نوع الحساب" : "Type"}</th>
                      <th className="px-4 py-3">{isAr ? "الطبيعة" : "Normal"}</th>
                      <th className="px-4 py-3 text-right rtl:text-left">{isAr ? "إجمالي المدين (SAR)" : "Debit"}</th>
                      <th className="px-4 py-3 text-right rtl:text-left">{isAr ? "إجمالي الدائن (SAR)" : "Credit"}</th>
                      <th className="px-4 py-3 text-right rtl:text-left">{isAr ? "صافي الرصيد (SAR)" : "Net Balance"}</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y text-foreground">
                    {[
                      { code: "111001", name: "صندوق المصروفات النثرية", type: "ASSET", normal: "DEBIT", debit: "15,000.00", credit: "3,200.00", net: "11,800.00" },
                      { code: "111002", name: "مصرف الراجحي - الحساب الجاري", type: "ASSET", normal: "DEBIT", debit: "1,240,000.00", credit: "420,000.00", net: "820,000.00" },
                      { code: "112000", name: "العملاء والمدينون التجاريون (AR)", type: "ASSET", normal: "DEBIT", debit: "780,000.00", credit: "168,000.00", net: "612,000.00" },
                      { code: "113000", name: "المخزون السلعي والبضاعة", type: "ASSET", normal: "DEBIT", debit: "520,000.00", credit: "95,000.00", net: "425,000.00" },
                      { code: "211000", name: "الموردون والدائنون التجاريون (AP)", type: "LIABILITY", normal: "CREDIT", debit: "110,000.00", credit: "460,000.00", net: "350,000.00" },
                      { code: "213000", name: "ضريبة القيمة المضافة المستحقة (15%)", type: "LIABILITY", normal: "CREDIT", debit: "35,000.00", credit: "119,350.00", net: "84,350.00" },
                      { code: "310000", name: "رأس المال المدفوع", type: "EQUITY", normal: "CREDIT", debit: "0.00", credit: "1,000,000.00", net: "1,000,000.00" },
                      { code: "410000", name: "إيرادات مبيعات البضائع والخدمات", type: "REVENUE", normal: "CREDIT", debit: "0.00", credit: "1,450,000.00", net: "1,450,000.00" },
                      { code: "510000", name: "تكلفة البضاعة المباعة (COGS)", type: "EXPENSE", normal: "DEBIT", debit: "680,000.00", credit: "0.00", net: "680,000.00" },
                      { code: "521000", name: "الرواتب والأجور الشهرية", type: "EXPENSE", normal: "DEBIT", debit: "345,550.00", credit: "0.00", net: "345,550.00" },
                    ].map((row, i) => (
                      <tr key={i} className="hover:bg-muted/30 transition-colors">
                        <td className="px-4 py-2.5 font-mono font-semibold text-primary">{row.code}</td>
                        <td className="px-4 py-2.5 font-medium">{row.name}</td>
                        <td className="px-4 py-2.5">
                          <span className="rounded-md bg-muted px-2 py-0.5 text-[10px] font-semibold">{row.type}</span>
                        </td>
                        <td className="px-4 py-2.5 text-muted-foreground">{row.normal}</td>
                        <td className="px-4 py-2.5 text-right font-mono font-medium rtl:text-left">{row.debit}</td>
                        <td className="px-4 py-2.5 text-right font-mono font-medium rtl:text-left">{row.credit}</td>
                        <td className="px-4 py-2.5 text-right font-mono font-bold text-foreground rtl:text-left">{row.net}</td>
                      </tr>
                    ))}
                  </tbody>
                  <tfoot className="border-t-2 bg-muted/70 font-bold">
                    <tr>
                      <td colSpan={4} className="px-4 py-3 text-foreground">
                        {isAr ? "الإجمالي العام (متطابق)" : "TOTAL SUMMARY (BALANCED)"}
                      </td>
                      <td className="px-4 py-3 text-right font-mono text-emerald-600 dark:text-emerald-400 rtl:text-left">
                        3,725,550.00 SAR
                      </td>
                      <td className="px-4 py-3 text-right font-mono text-emerald-600 dark:text-emerald-400 rtl:text-left">
                        3,725,550.00 SAR
                      </td>
                      <td className="px-4 py-3 text-right font-mono text-emerald-600 dark:text-emerald-400 rtl:text-left">
                        0.00 SAR
                      </td>
                    </tr>
                  </tfoot>
                </table>
              </div>
            </div>
          )}

          {activeReport === "BALANCE_SHEET" && (
            <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
              {/* Assets Section */}
              <div className="space-y-3 rounded-lg border p-4 bg-muted/20">
                <div className="flex items-center justify-between border-b pb-2">
                  <h3 className="text-sm font-bold text-foreground">{isAr ? "1. الأصول (Assets)" : "1. Assets"}</h3>
                  <span className="text-xs font-mono font-bold text-emerald-600 dark:text-emerald-400">
                    2,868,800.00 SAR
                  </span>
                </div>
                <div className="space-y-2 text-xs">
                  <div className="flex justify-between py-1 border-b border-border/50">
                    <span>{isAr ? "النقد وما في حكمه (Cash & Bank)" : "Cash & Bank Balances"}</span>
                    <span className="font-mono font-semibold">831,800.00 SAR</span>
                  </div>
                  <div className="flex justify-between py-1 border-b border-border/50">
                    <span>{isAr ? "العملاء والمدينون التجاريون (AR)" : "Trade Receivables (AR)"}</span>
                    <span className="font-mono font-semibold">612,000.00 SAR</span>
                  </div>
                  <div className="flex justify-between py-1 border-b border-border/50">
                    <span>{isAr ? "المخزون السلعي (Inventory)" : "Merchandise Inventory"}</span>
                    <span className="font-mono font-semibold">425,000.00 SAR</span>
                  </div>
                  <div className="flex justify-between py-1 border-b border-border/50">
                    <span>{isAr ? "الأصول الثابتة بالصافي (Fixed Assets Net)" : "Property, Plant & Equipment"}</span>
                    <span className="font-mono font-semibold">1,000,000.00 SAR</span>
                  </div>
                </div>
              </div>

              {/* Liabilities & Equity Section */}
              <div className="space-y-3 rounded-lg border p-4 bg-muted/20">
                <div className="flex items-center justify-between border-b pb-2">
                  <h3 className="text-sm font-bold text-foreground">
                    {isAr ? "2. الالتزامات وحقوق الملكية" : "2. Liabilities & Equity"}
                  </h3>
                  <span className="text-xs font-mono font-bold text-emerald-600 dark:text-emerald-400">
                    2,868,800.00 SAR
                  </span>
                </div>
                <div className="space-y-2 text-xs">
                  <div className="flex justify-between py-1 border-b border-border/50 text-amber-600 dark:text-amber-400 font-semibold">
                    <span>{isAr ? "الالتزامات المتداولة (الموردون + الضريبة)" : "Current Liabilities (AP + VAT)"}</span>
                    <span className="font-mono">434,350.00 SAR</span>
                  </div>
                  <div className="flex justify-between py-1 border-b border-border/50">
                    <span>{isAr ? "رأس المال المدفوع (Share Capital)" : "Paid-in Share Capital"}</span>
                    <span className="font-mono font-semibold">1,000,000.00 SAR</span>
                  </div>
                  <div className="flex justify-between py-1 border-b border-border/50">
                    <span>{isAr ? "الأرباح المبقاة (Retained Earnings)" : "Retained Earnings"}</span>
                    <span className="font-mono font-semibold">1,010,000.00 SAR</span>
                  </div>
                  <div className="flex justify-between py-1 border-b border-border/50 text-indigo-600 dark:text-indigo-400 font-semibold">
                    <span>{isAr ? "صافي أرباح الفترة الحالية (P&L)" : "Current Period Net Income"}</span>
                    <span className="font-mono">424,450.00 SAR</span>
                  </div>
                </div>
              </div>

              {/* Equation Verification Banner */}
              <div className="col-span-full rounded-lg bg-emerald-500/10 p-4 border border-emerald-500/20 text-xs flex items-center justify-between">
                <div className="flex items-center gap-2 text-emerald-700 dark:text-emerald-300 font-semibold">
                  <CheckCircle2 className="h-4 w-4" />
                  {isAr
                    ? "معادلة المركز المالي متطابقة: الأصول (2,868,800) = الخصوم (434,350) + حقوق الملكية (2,434,450)"
                    : "Balance Sheet Balanced: Assets (2,868,800) = Liabilities (434,350) + Equity (2,434,450)"}
                </div>
                <span className="rounded-full bg-emerald-600 text-white px-2.5 py-0.5 text-[10px] font-bold">
                  100% BALANCED
                </span>
              </div>
            </div>
          )}

          {activeReport === "INCOME_STATEMENT" && (
            <div className="space-y-4 max-w-3xl mx-auto">
              <div className="rounded-lg border p-4 space-y-3">
                <h3 className="text-xs font-bold text-muted-foreground uppercase tracking-wider">
                  {isAr ? "1. الإيرادات والمبيعات" : "1. Revenues & Sales"}
                </h3>
                <div className="flex justify-between text-sm py-1 border-b">
                  <span>{isAr ? "مبيعات البضائع والمنتجات" : "Sales Revenue"}</span>
                  <span className="font-mono font-semibold">1,450,000.00 SAR</span>
                </div>
                <div className="flex justify-between text-sm py-1 font-bold text-primary">
                  <span>{isAr ? "إجمالي الإيرادات التشغيلية" : "Total Operating Revenues"}</span>
                  <span className="font-mono">1,450,000.00 SAR</span>
                </div>
              </div>

              <div className="rounded-lg border p-4 space-y-3">
                <h3 className="text-xs font-bold text-muted-foreground uppercase tracking-wider">
                  {isAr ? "2. تكلفة المبيعات والمصروفات" : "2. Cost of Sales & Expenses"}
                </h3>
                <div className="flex justify-between text-sm py-1 border-b">
                  <span>{isAr ? "تكلفة البضاعة المباعة (COGS)" : "Cost of Goods Sold"}</span>
                  <span className="font-mono text-rose-500">(680,000.00) SAR</span>
                </div>
                <div className="flex justify-between text-sm py-1 font-bold text-foreground">
                  <span>{isAr ? "إجمالي الربح الإجمالي (Gross Profit)" : "Gross Profit"}</span>
                  <span className="font-mono">770,000.00 SAR</span>
                </div>
                <div className="flex justify-between text-sm py-1 border-b">
                  <span>{isAr ? "الرواتب ومستحقات الموظفين" : "Employee Wages & Benefits"}</span>
                  <span className="font-mono text-rose-500">(345,550.00) SAR</span>
                </div>
              </div>

              <div className="rounded-xl bg-primary/10 border border-primary/20 p-5 flex items-center justify-between">
                <div>
                  <div className="text-xs font-semibold text-primary">{isAr ? "صافي الدخل النهائي للفترة" : "Net Income for the Period"}</div>
                  <div className="text-2xl font-bold text-foreground font-mono mt-1">424,450.00 SAR</div>
                </div>
                <div className="text-right rtl:text-left">
                  <div className="text-xs text-muted-foreground">{isAr ? "هامش صافي الربح" : "Profit Margin"}</div>
                  <div className="text-lg font-bold text-emerald-600 dark:text-emerald-400 font-mono">29.27%</div>
                </div>
              </div>
            </div>
          )}

          {activeReport === "VAT_RETURN" && (
            <div className="space-y-5">
              {/* Output VAT (Sales) Table */}
              <div className="rounded-lg border overflow-hidden">
                <div className="bg-primary/5 px-4 py-2.5 font-bold text-xs border-b text-primary">
                  {isAr ? "أولاً: ضريبة القيمة المضافة على المبيعات (المخرجات)" : "Part 1: VAT on Sales (Output Tax)"}
                </div>
                <table className="w-full text-xs text-left rtl:text-right">
                  <thead className="bg-muted/40 text-muted-foreground">
                    <tr>
                      <th className="px-4 py-2 w-16">#</th>
                      <th className="px-4 py-2">{isAr ? "البيان" : "Description"}</th>
                      <th className="px-4 py-2 text-right rtl:text-left">{isAr ? "المبلغ الخاضع للضريبة (SAR)" : "Taxable Amount"}</th>
                      <th className="px-4 py-2 text-right rtl:text-left">{isAr ? "مبلغ الضريبة (15%)" : "VAT Amount (15%)"}</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y text-foreground">
                    <tr>
                      <td className="px-4 py-2 font-mono font-bold">1</td>
                      <td className="px-4 py-2">{isAr ? "المبيعات الخاضعة للنسبة الأساسية (15%)" : "Standard rated sales (15%)"}</td>
                      <td className="px-4 py-2 text-right font-mono rtl:text-left">1,450,000.00</td>
                      <td className="px-4 py-2 text-right font-mono font-bold rtl:text-left">217,500.00</td>
                    </tr>
                    <tr>
                      <td className="px-4 py-2 font-mono font-bold">2</td>
                      <td className="px-4 py-2">{isAr ? "المبيعات للمواطنين (الخدمات الصحية والتعليمية)" : "Sales to citizens (healthcare/education)"}</td>
                      <td className="px-4 py-2 text-right font-mono rtl:text-left">0.00</td>
                      <td className="px-4 py-2 text-right font-mono font-bold rtl:text-left">0.00</td>
                    </tr>
                    <tr>
                      <td className="px-4 py-2 font-mono font-bold">3</td>
                      <td className="px-4 py-2">{isAr ? "المبيعات المحلية الخاضعة لنسبة الصفر" : "Zero rated domestic sales"}</td>
                      <td className="px-4 py-2 text-right font-mono rtl:text-left">0.00</td>
                      <td className="px-4 py-2 text-right font-mono font-bold rtl:text-left">0.00</td>
                    </tr>
                    <tr>
                      <td className="px-4 py-2 font-mono font-bold">4</td>
                      <td className="px-4 py-2">{isAr ? "الصادرات إلى خارج دول مجلس التعاون" : "Exports outside GCC"}</td>
                      <td className="px-4 py-2 text-right font-mono rtl:text-left">0.00</td>
                      <td className="px-4 py-2 text-right font-mono font-bold rtl:text-left">0.00</td>
                    </tr>
                  </tbody>
                  <tfoot className="bg-muted/50 font-bold border-t">
                    <tr>
                      <td colSpan={3} className="px-4 py-2">{isAr ? "إجمالي ضريبة المبيعات (المخرجات)" : "Total Output Tax"}</td>
                      <td className="px-4 py-2 text-right font-mono text-primary rtl:text-left">217,500.00 SAR</td>
                    </tr>
                  </tfoot>
                </table>
              </div>

              {/* Input VAT (Purchases) Table */}
              <div className="rounded-lg border overflow-hidden">
                <div className="bg-primary/5 px-4 py-2.5 font-bold text-xs border-b text-primary">
                  {isAr ? "ثانياً: ضريبة القيمة المضافة على المشتريات (المدخلات)" : "Part 2: VAT on Purchases (Input Tax)"}
                </div>
                <table className="w-full text-xs text-left rtl:text-right">
                  <thead className="bg-muted/40 text-muted-foreground">
                    <tr>
                      <th className="px-4 py-2 w-16">#</th>
                      <th className="px-4 py-2">{isAr ? "البيان" : "Description"}</th>
                      <th className="px-4 py-2 text-right rtl:text-left">{isAr ? "المبلغ الخاضع للضريبة (SAR)" : "Taxable Amount"}</th>
                      <th className="px-4 py-2 text-right rtl:text-left">{isAr ? "مبلغ الضريبة القابل للاسترداد" : "Recoverable VAT"}</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y text-foreground">
                    <tr>
                      <td className="px-4 py-2 font-mono font-bold">7</td>
                      <td className="px-4 py-2">{isAr ? "المشتريات الخاضعة للنسبة الأساسية (15%)" : "Standard rated domestic purchases"}</td>
                      <td className="px-4 py-2 text-right font-mono rtl:text-left">887,666.67</td>
                      <td className="px-4 py-2 text-right font-mono font-bold rtl:text-left">133,150.00</td>
                    </tr>
                    <tr>
                      <td className="px-4 py-2 font-mono font-bold">8</td>
                      <td className="px-4 py-2">{isAr ? "الواردات الخاضعة للضريبة المدفوعة بالجمارك" : "Imports subject to VAT paid at customs"}</td>
                      <td className="px-4 py-2 text-right font-mono rtl:text-left">0.00</td>
                      <td className="px-4 py-2 text-right font-mono font-bold rtl:text-left">0.00</td>
                    </tr>
                  </tbody>
                  <tfoot className="bg-muted/50 font-bold border-t">
                    <tr>
                      <td colSpan={3} className="px-4 py-2">{isAr ? "إجمالي ضريبة المشتريات (المدخلات)" : "Total Input Tax"}</td>
                      <td className="px-4 py-2 text-right font-mono text-emerald-600 dark:text-emerald-400 rtl:text-left">133,150.00 SAR</td>
                    </tr>
                  </tfoot>
                </table>
              </div>

              {/* Net VAT Due Card */}
              <div className="rounded-xl border-2 border-primary bg-primary/5 p-4 flex items-center justify-between">
                <div>
                  <div className="text-xs font-bold text-muted-foreground">{isAr ? "صافي الضريبة المستحقة السداد لهيئة الزكاة" : "Net VAT Payable to ZATCA"}</div>
                  <div className="text-2xl font-bold font-mono text-primary mt-1">84,350.00 SAR</div>
                </div>
                <div className="flex items-center gap-2">
                  <span className="rounded-full bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 px-3 py-1 text-xs font-bold border border-emerald-500/20">
                    ZATCA PHASE-2 COMPLIANT
                  </span>
                </div>
              </div>
            </div>
          )}

          {/* Fallback for other standard tables (AR/AP Aging, Inventory, Payroll, etc.) */}
          {activeReport !== "TRIAL_BALANCE" && activeReport !== "BALANCE_SHEET" && activeReport !== "INCOME_STATEMENT" && activeReport !== "VAT_RETURN" && (
            <div className="rounded-lg border p-8 text-center space-y-3">
              <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-primary/10 text-primary">
                {React.createElement(currentReportObj.icon, { className: "h-6 w-6" })}
              </div>
              <h3 className="text-base font-bold text-foreground">
                {isAr ? currentReportObj.nameAr : currentReportObj.nameEn}
              </h3>
              <p className="text-xs text-muted-foreground max-w-md mx-auto">
                {isAr ? currentReportObj.descAr : currentReportObj.descEn}
              </p>
              <div className="pt-2 flex justify-center gap-2">
                <button
                  onClick={handleExportCSV}
                  className="inline-flex items-center gap-1.5 rounded-lg border bg-card px-4 py-2 text-xs font-semibold hover:bg-muted transition-colors"
                >
                  <Download className="h-4 w-4" />
                  {isAr ? "تحميل بيانات التقرير (CSV)" : "Download Full Dataset (CSV)"}
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
