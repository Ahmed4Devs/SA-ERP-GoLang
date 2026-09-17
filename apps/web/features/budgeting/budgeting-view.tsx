"use client";

import React, { useState } from "react";
import {
  PieChart,
  DollarSign,
  TrendingUp,
  AlertTriangle,
  CheckCircle2,
  Clock,
  ArrowRightLeft,
  Plus,
  Search,
  Filter,
  Layers,
  ChevronRight,
  Sparkles,
  BarChart3,
  Calendar,
  Building,
  Tag,
  FileSpreadsheet,
  X,
  FileCheck,
  Percent,
} from "lucide-react";
import { useLocale } from "@/components/layout/shell";

interface BudgetRecord {
  id: string;
  code: string;
  nameAr: string;
  nameEn: string;
  fiscalYear: string;
  currency: string;
  budgetType: "OPERATING" | "CAPITAL" | "PROJECT" | "CASH_FLOW";
  status: "DRAFT" | "SUBMITTED" | "ACTIVE" | "REVISED" | "CLOSED";
  totalBudgeted: number;
  totalActual: number;
  totalCommitted: number;
  descriptionAr: string;
  descriptionEn: string;
  createdAt: string;
  versionCount: number;
}

interface BudgetLineRecord {
  id: string;
  budgetId: string;
  accountCode: string;
  accountNameAr: string;
  accountNameEn: string;
  departmentNameAr: string;
  costCenterNameAr: string;
  projectCode: string;
  periodNameAr: string;
  budgetedAmount: number;
  actualAmount: number;
  committedAmount: number;
  descriptionAr: string;
}

interface TransferRecord {
  id: string;
  transferNumber: string;
  sourceAccountCode: string;
  sourceAccountNameAr: string;
  targetAccountCode: string;
  targetAccountNameAr: string;
  amount: number;
  reason: string;
  status: "COMPLETED" | "PENDING";
  createdAt: string;
}

const INITIAL_BUDGETS: BudgetRecord[] = [
  {
    id: "b-001",
    code: "BUD-2026-OPEX",
    nameAr: "الميزانية التشغيلية السنوية 2026",
    nameEn: "Annual Operating Budget 2026",
    fiscalYear: "2026",
    currency: "SAR",
    budgetType: "OPERATING",
    status: "ACTIVE",
    totalBudgeted: 1850000,
    totalActual: 425000,
    totalCommitted: 85000,
    descriptionAr: "الميزانية التشغيلية المعتمدة للمصروفات الإدارية والتسويقية والرواتب للعام المالي 2026",
    descriptionEn: "Approved annual operating budget for general, administrative and sales expenses",
    createdAt: "2026-01-01",
    versionCount: 1,
  },
  {
    id: "b-002",
    code: "BUD-2026-CAPEX",
    nameAr: "ميزانية النفقات الرأسمالية 2026",
    nameEn: "Capital Expenditure Budget 2026",
    fiscalYear: "2026",
    currency: "SAR",
    budgetType: "CAPITAL",
    status: "ACTIVE",
    totalBudgeted: 750000,
    totalActual: 180000,
    totalCommitted: 120000,
    descriptionAr: "الاستثمارات الرأسمالية المخططة لتحديث الآلات ومعدات تقنية المعلومات وأسطول النقل",
    descriptionEn: "Planned capital expenditure for IT equipment, fleet expansion and manufacturing machinery",
    createdAt: "2026-01-15",
    versionCount: 1,
  },
  {
    id: "b-003",
    code: "BUD-2026-MKT-EXP",
    nameAr: "ميزانية التوسع التسويقي وحملات يوم التأسيس",
    nameEn: "Market Expansion & Founding Day Campaign",
    fiscalYear: "2026",
    currency: "SAR",
    budgetType: "PROJECT",
    status: "SUBMITTED",
    totalBudgeted: 350000,
    totalActual: 0,
    totalCommitted: 0,
    descriptionAr: "حملات إعلانية وتغطيات إعلامية للترويج لعلامة الشركة خلال فعاليات الربع الأول",
    descriptionEn: "Targeted digital and media acquisition campaigns for Q1 seasonal events",
    createdAt: "2026-02-10",
    versionCount: 1,
  },
];

const INITIAL_LINES: BudgetLineRecord[] = [
  {
    id: "l-001",
    budgetId: "b-001",
    accountCode: "5101",
    accountNameAr: "رواتب وأجور الموظفين والبدلات",
    accountNameEn: "Salaries, Wages and Allowances",
    departmentNameAr: "الموارد البشرية",
    costCenterNameAr: "الإدارة العامة",
    projectCode: "HQ-OPS",
    periodNameAr: "الربع الأول 2026",
    budgetedAmount: 1000000,
    actualAmount: 250000,
    committedAmount: 0,
    descriptionAr: "مخصص الرواتب الأساسية وبدلات السكن والنقل والتأمينات",
  },
  {
    id: "l-002",
    budgetId: "b-001",
    accountCode: "5201",
    accountNameAr: "إيجار المقار والمباني والمستودعات",
    accountNameEn: "Building & Warehouse Leases",
    departmentNameAr: "المالية والإدارية",
    costCenterNameAr: "الإدارة العامة",
    projectCode: "HQ-FAC",
    periodNameAr: "الربع الأول 2026",
    budgetedAmount: 350000,
    actualAmount: 87500,
    committedAmount: 25000,
    descriptionAr: "إيجار المقر الرئيسي وفرع الرياض ومستودع السلي",
  },
  {
    id: "l-003",
    budgetId: "b-001",
    accountCode: "5301",
    accountNameAr: "مصروفات التسويق والإعلانات الرقمية",
    accountNameEn: "Digital Marketing & Advertising",
    departmentNameAr: "المبيعات والتسويق",
    costCenterNameAr: "مركز مبيعات الرياض",
    projectCode: "MKT-Q1",
    periodNameAr: "الربع الأول 2026",
    budgetedAmount: 300000,
    actualAmount: 62500,
    committedAmount: 40000,
    descriptionAr: "الحملات الإعلانية على منصات التواصل الاجتماعي والرعايات",
  },
  {
    id: "l-004",
    budgetId: "b-001",
    accountCode: "5401",
    accountNameAr: "تراخيص تقنية المعلومات واشتراكات السحابة",
    accountNameEn: "IT Software Subscriptions & Cloud",
    departmentNameAr: "تقنية المعلومات",
    costCenterNameAr: "الإدارة العامة",
    projectCode: "TECH-INFRA",
    periodNameAr: "الربع الأول 2026",
    budgetedAmount: 200000,
    actualAmount: 25000,
    committedAmount: 20000,
    descriptionAr: "اشتراكات خوادم AWS ومنظومة ERP والتراخيص الأمنية",
  },
  {
    id: "l-005",
    budgetId: "b-002",
    accountCode: "1503",
    accountNameAr: "أجهزة حاسب آلي وخوادم رئيسية",
    accountNameEn: "Computer Equipment & Enterprise Servers",
    departmentNameAr: "تقنية المعلومات",
    costCenterNameAr: "الإدارة العامة",
    projectCode: "CAP-SRV",
    periodNameAr: "النصف الأول 2026",
    budgetedAmount: 450000,
    actualAmount: 110000,
    committedAmount: 80000,
    descriptionAr: "تجهيز غرفة البيانات المركزية وترقية أجهزة الموظفين",
  },
  {
    id: "l-006",
    budgetId: "b-002",
    accountCode: "1504",
    accountNameAr: "أسطول سيارات وشاحنات التوزيع",
    accountNameEn: "Delivery Fleet Vehicles & Trucks",
    departmentNameAr: "العمليات واللوجستيات",
    costCenterNameAr: "مركز العمليات",
    projectCode: "CAP-FLEET",
    periodNameAr: "النصف الأول 2026",
    budgetedAmount: 300000,
    actualAmount: 70000,
    committedAmount: 40000,
    descriptionAr: "شراء شاحنتين مبردتين للتوزيع اللوجستي",
  },
];

const INITIAL_TRANSFERS: TransferRecord[] = [
  {
    id: "trf-001",
    transferNumber: "TRF-20260215-001",
    sourceAccountCode: "5201 - إيجار المقار",
    sourceAccountNameAr: "إيجار المقار والمباني",
    targetAccountCode: "5401 - تقنية المعلومات",
    targetAccountNameAr: "تراخيص تقنية المعلومات واشتراكات السحابة",
    amount: 15000,
    reason: "تغطية تكاليف توسعة البنية التحتية السحابية من وفر إيجار مستودع الخرج",
    status: "COMPLETED",
    createdAt: "2026-02-15 11:30",
  },
  {
    id: "trf-002",
    transferNumber: "TRF-20260228-002",
    sourceAccountCode: "5301 - التسويق",
    sourceAccountNameAr: "مصروفات التسويق والإعلانات",
    targetAccountCode: "5101 - الرواتب",
    targetAccountNameAr: "رواتب وأجور الموظفين والبدلات",
    amount: 25000,
    reason: "تمويل حوافز الربع الأول لفرق المبيعات الميدانية",
    status: "COMPLETED",
    createdAt: "2026-02-28 15:45",
  },
];

export function BudgetingView() {
  const { locale } = useLocale();
  const isAr = locale === "ar";

  const [activeTab, setActiveTab] = useState<"budgets" | "lines" | "variance" | "transfers">("budgets");
  const [budgets, setBudgets] = useState<BudgetRecord[]>(INITIAL_BUDGETS);
  const [lines, setLines] = useState<BudgetLineRecord[]>(INITIAL_LINES);
  const [transfers, setTransfers] = useState<TransferRecord[]>(INITIAL_TRANSFERS);

  const [selectedBudgetId, setSelectedBudgetId] = useState<string>("b-001");
  const [searchQuery, setSearchQuery] = useState<string>("");
  const [filterType, setFilterType] = useState<string>("ALL");
  const [notification, setNotification] = useState<string | null>(null);

  // Modals
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [isAddLineOpen, setIsAddLineOpen] = useState(false);
  const [isTransferOpen, setIsTransferOpen] = useState(false);

  // Forms
  const [newBudget, setNewBudget] = useState({
    code: "",
    nameAr: "",
    nameEn: "",
    fiscalYear: "2026",
    budgetType: "OPERATING" as BudgetRecord["budgetType"],
    descriptionAr: "",
    totalBudgeted: "",
  });

  const [newLine, setNewLine] = useState({
    accountCode: "5101",
    accountNameAr: "رواتب وأجور الموظفين والبدلات",
    departmentNameAr: "الموارد البشرية",
    costCenterNameAr: "الإدارة العامة",
    projectCode: "HQ-GEN",
    periodNameAr: "الربع الأول 2026",
    budgetedAmount: "",
    descriptionAr: "",
  });

  const [transferData, setTransferData] = useState({
    sourceLineId: "l-002",
    targetLineId: "l-004",
    amount: "",
    reason: "",
  });

  // Notifications
  const showNotification = (msg: string) => {
    setNotification(msg);
    setTimeout(() => setNotification(null), 4000);
  };

  // KPIs
  const totalBudgetedAll = budgets.reduce((acc, b) => acc + b.totalBudgeted, 0);
  const totalActualAll = budgets.reduce((acc, b) => acc + b.totalActual, 0);
  const totalCommittedAll = budgets.reduce((acc, b) => acc + b.totalCommitted, 0);
  const totalRemainingAll = totalBudgetedAll - totalActualAll;
  const overallUsageRate = totalBudgetedAll > 0 ? ((totalActualAll / totalBudgetedAll) * 100).toFixed(1) : "0.0";

  // Actions
  const handleApproveBudget = (id: string) => {
    setBudgets((prev) =>
      prev.map((b) => (b.id === id ? { ...b, status: "ACTIVE" } : b))
    );
    showNotification(isAr ? "تم اعتماد الميزانية وتفعيلها بنجاح" : "Budget successfully approved and activated");
  };

  const handleSubmitBudget = (id: string) => {
    setBudgets((prev) =>
      prev.map((b) => (b.id === id ? { ...b, status: "SUBMITTED" } : b))
    );
    showNotification(isAr ? "تم تقديم الميزانية للاعتماد المالي" : "Budget submitted for board approval");
  };

  const handleCreateBudget = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newBudget.code || !newBudget.nameAr) return;

    const amt = parseFloat(newBudget.totalBudgeted) || 0;
    const b: BudgetRecord = {
      id: `b-${Date.now()}`,
      code: newBudget.code,
      nameAr: newBudget.nameAr,
      nameEn: newBudget.nameEn || newBudget.code,
      fiscalYear: newBudget.fiscalYear,
      currency: "SAR",
      budgetType: newBudget.budgetType,
      status: "DRAFT",
      totalBudgeted: amt,
      totalActual: 0,
      totalCommitted: 0,
      descriptionAr: newBudget.descriptionAr,
      descriptionEn: "",
      createdAt: new Date().toISOString().split("T")[0],
      versionCount: 1,
    };

    setBudgets([b, ...budgets]);
    setIsCreateOpen(false);
    setNewBudget({
      code: "",
      nameAr: "",
      nameEn: "",
      fiscalYear: "2026",
      budgetType: "OPERATING",
      descriptionAr: "",
      totalBudgeted: "",
    });
    showNotification(isAr ? "تم إنشاء مسودة الميزانية بنجاح" : "Draft budget created successfully");
  };

  const handleAddLine = (e: React.FormEvent) => {
    e.preventDefault();
    const amt = parseFloat(newLine.budgetedAmount) || 0;
    if (amt <= 0) return;

    const l: BudgetLineRecord = {
      id: `l-${Date.now()}`,
      budgetId: selectedBudgetId,
      accountCode: newLine.accountCode,
      accountNameAr: newLine.accountNameAr,
      accountNameEn: newLine.accountNameAr,
      departmentNameAr: newLine.departmentNameAr,
      costCenterNameAr: newLine.costCenterNameAr,
      projectCode: newLine.projectCode,
      periodNameAr: newLine.periodNameAr,
      budgetedAmount: amt,
      actualAmount: 0,
      committedAmount: 0,
      descriptionAr: newLine.descriptionAr,
    };

    setLines([...lines, l]);
    // update parent budget total
    setBudgets((prev) =>
      prev.map((b) =>
        b.id === selectedBudgetId ? { ...b, totalBudgeted: b.totalBudgeted + amt } : b
      )
    );

    setIsAddLineOpen(false);
    setNewLine({
      accountCode: "5101",
      accountNameAr: "رواتب وأجور الموظفين والبدلات",
      departmentNameAr: "الموارد البشرية",
      costCenterNameAr: "الإدارة العامة",
      projectCode: "HQ-GEN",
      periodNameAr: "الربع الأول 2026",
      budgetedAmount: "",
      descriptionAr: "",
    });
    showNotification(isAr ? "تمت إضافة بند الميزانية بنجاح" : "Budget line item added successfully");
  };

  const handleExecuteTransfer = (e: React.FormEvent) => {
    e.preventDefault();
    const amt = parseFloat(transferData.amount) || 0;
    if (amt <= 0) return;

    const src = lines.find((l) => l.id === transferData.sourceLineId);
    const tgt = lines.find((l) => l.id === transferData.targetLineId);

    if (!src || !tgt) return;
    if (src.budgetedAmount < amt) {
      alert(isAr ? "المبلغ المتوفر في البند المصدر غير كافٍ للمناقلة" : "Insufficient funds in source budget line");
      return;
    }

    setLines((prev) =>
      prev.map((l) => {
        if (l.id === src.id) return { ...l, budgetedAmount: l.budgetedAmount - amt };
        if (l.id === tgt.id) return { ...l, budgetedAmount: l.budgetedAmount + amt };
        return l;
      })
    );

    const trf: TransferRecord = {
      id: `trf-${Date.now()}`,
      transferNumber: `TRF-${new Date().toISOString().slice(0, 10).replace(/-/g, "")}-${Math.floor(100 + Math.random() * 900)}`,
      sourceAccountCode: `${src.accountCode} - ${src.accountNameAr}`,
      sourceAccountNameAr: src.accountNameAr,
      targetAccountCode: `${tgt.accountCode} - ${tgt.accountNameAr}`,
      targetAccountNameAr: tgt.accountNameAr,
      amount: amt,
      reason: transferData.reason || "مناقلة مالية تشغيلية داخلية",
      status: "COMPLETED",
      createdAt: new Date().toISOString().replace("T", " ").slice(0, 16),
    };

    setTransfers([trf, ...transfers]);
    setIsTransferOpen(false);
    setTransferData({
      sourceLineId: "l-002",
      targetLineId: "l-004",
      amount: "",
      reason: "",
    });
    showNotification(isAr ? "تم تنفيذ مناقلة الميزانية وتحديث الأرصدة بنجاح" : "Budget reallocation executed successfully");
  };

  // Filtered views
  const filteredBudgets = budgets.filter((b) => {
    const matchesSearch =
      b.code.toLowerCase().includes(searchQuery.toLowerCase()) ||
      b.nameAr.includes(searchQuery) ||
      b.nameEn.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesType = filterType === "ALL" || b.budgetType === filterType;
    return matchesSearch && matchesType;
  });

  const activeBudgetLines = lines.filter((l) => l.budgetId === selectedBudgetId);
  const currentSelectedBudget = budgets.find((b) => b.id === selectedBudgetId) || budgets[0];

  return (
    <div className="space-y-6 pb-12">
      {/* Toast Notification */}
      {notification && (
        <div className="fixed top-20 left-1/2 -translate-x-1/2 z-50 bg-emerald-600 text-white px-5 py-3 rounded-xl shadow-lg flex items-center gap-3 animate-in fade-in slide-in-from-top-4">
          <CheckCircle2 className="h-5 w-5" />
          <span className="font-medium text-sm">{notification}</span>
        </div>
      )}

      {/* Header Banner */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-gradient-to-r from-card via-card to-secondary/30 p-6 rounded-2xl border shadow-sm">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-primary/10 text-primary border border-primary/20">
              <Sparkles className="h-3.5 w-3.5" />
              {isAr ? "المرحلة 24: إدارة الميزانيات والرقابة المالية" : "Phase 24: Budget Management & Controls"}
            </span>
            <span className="text-xs text-muted-foreground">• العام المالي 2026</span>
          </div>
          <h1 className="text-2xl font-bold tracking-tight text-foreground">
            {isAr ? "نظام الميزانيات التقديرية والمناقلات" : "Enterprise Budgeting & Financial Controls"}
          </h1>
          <p className="text-sm text-muted-foreground mt-1">
            {isAr
              ? "تخطيط الميزانيات التشغيلية والرأسمالية، الربط متعدد الأبعاد مع دليل الحسابات، ومراقبة الانحرافات اللحظية"
              : "Comprehensive multi-dimensional budgeting, GL line allocations, fund reallocations, and real-time variance monitoring"}
          </p>
        </div>

        <div className="flex items-center gap-2.5 flex-wrap">
          <button
            onClick={() => setIsTransferOpen(true)}
            className="inline-flex items-center gap-2 px-4 py-2 rounded-xl text-sm font-medium border bg-background hover:bg-muted text-foreground transition-colors shadow-sm"
          >
            <ArrowRightLeft className="h-4 w-4 text-primary" />
            <span>{isAr ? "مناقلة مالية" : "Budget Reallocation"}</span>
          </button>
          <button
            onClick={() => setIsCreateOpen(true)}
            className="inline-flex items-center gap-2 px-4 py-2 rounded-xl text-sm font-medium bg-primary text-primary-foreground hover:bg-primary/90 transition-colors shadow-sm"
          >
            <Plus className="h-4 w-4" />
            <span>{isAr ? "إنشاء ميزانية جديدة" : "New Budget"}</span>
          </button>
        </div>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* Card 1: Total Budgeted */}
        <div className="p-5 rounded-2xl border bg-card/60 backdrop-blur shadow-sm flex flex-col justify-between hover:border-primary/40 transition-colors">
          <div className="flex items-center justify-between text-muted-foreground mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider">
              {isAr ? "إجمالي الميزانية المعتمدة" : "Total Budgeted"}
            </span>
            <div className="p-2 rounded-xl bg-primary/10 text-primary">
              <PieChart className="h-5 w-5" />
            </div>
          </div>
          <div>
            <div className="text-2xl font-bold tracking-tight text-foreground">
              {totalBudgetedAll.toLocaleString()} <span className="text-sm font-normal text-muted-foreground">SAR</span>
            </div>
            <p className="text-xs text-muted-foreground mt-1 flex items-center gap-1">
              <span className="text-emerald-500 font-medium">3</span> {isAr ? "ميزانيات مسجلة للعام الحالي" : "budgets recorded"}
            </p>
          </div>
        </div>

        {/* Card 2: Actual Spent */}
        <div className="p-5 rounded-2xl border bg-card/60 backdrop-blur shadow-sm flex flex-col justify-between hover:border-primary/40 transition-colors">
          <div className="flex items-center justify-between text-muted-foreground mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider">
              {isAr ? "المنصرف الفعلي (Actual)" : "Actual Spent"}
            </span>
            <div className="p-2 rounded-xl bg-blue-500/10 text-blue-500">
              <DollarSign className="h-5 w-5" />
            </div>
          </div>
          <div>
            <div className="text-2xl font-bold tracking-tight text-foreground">
              {totalActualAll.toLocaleString()} <span className="text-sm font-normal text-muted-foreground">SAR</span>
            </div>
            <p className="text-xs text-muted-foreground mt-1 flex items-center gap-1">
              <span className="text-blue-500 font-medium">+{totalCommittedAll.toLocaleString()} SAR</span> {isAr ? "التزامات معلقة" : "committed"}
            </p>
          </div>
        </div>

        {/* Card 3: Remaining Balance */}
        <div className="p-5 rounded-2xl border bg-card/60 backdrop-blur shadow-sm flex flex-col justify-between hover:border-primary/40 transition-colors">
          <div className="flex items-center justify-between text-muted-foreground mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider">
              {isAr ? "الرصيد المتبقي المتاح" : "Remaining Funds"}
            </span>
            <div className="p-2 rounded-xl bg-emerald-500/10 text-emerald-500">
              <TrendingUp className="h-5 w-5" />
            </div>
          </div>
          <div>
            <div className="text-2xl font-bold tracking-tight text-emerald-600 dark:text-emerald-400">
              {totalRemainingAll.toLocaleString()} <span className="text-sm font-normal text-muted-foreground">SAR</span>
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              {isAr ? "سيولة مخصصة غير مستنفذة" : "Unspent allocated capital"}
            </p>
          </div>
        </div>

        {/* Card 4: Consumption Rate */}
        <div className="p-5 rounded-2xl border bg-card/60 backdrop-blur shadow-sm flex flex-col justify-between hover:border-primary/40 transition-colors">
          <div className="flex items-center justify-between text-muted-foreground mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider">
              {isAr ? "نسبة استهلاك الميزانية" : "Consumption Rate"}
            </span>
            <div className="p-2 rounded-xl bg-amber-500/10 text-amber-500">
              <Percent className="h-5 w-5" />
            </div>
          </div>
          <div>
            <div className="text-2xl font-bold tracking-tight text-foreground">
              {overallUsageRate}%
            </div>
            <div className="w-full bg-muted rounded-full h-2 mt-2 overflow-hidden">
              <div
                className="bg-primary h-2 rounded-full transition-all duration-500"
                style={{ width: `${Math.min(parseFloat(overallUsageRate), 100)}%` }}
              />
            </div>
          </div>
        </div>
      </div>

      {/* Navigation Tabs */}
      <div className="flex items-center gap-2 border-b pb-2 overflow-x-auto">
        <button
          onClick={() => setActiveTab("budgets")}
          className={`flex items-center gap-2 px-4 py-2.5 rounded-xl text-sm font-medium transition-all ${
            activeTab === "budgets"
              ? "bg-primary text-primary-foreground shadow-sm"
              : "text-muted-foreground hover:bg-muted hover:text-foreground"
          }`}
        >
          <PieChart className="h-4 w-4" />
          <span>{isAr ? "الميزانيات المعتمدة" : "Master Budgets"}</span>
          <span className="px-1.5 py-0.5 rounded-md text-xs bg-background/20 font-bold">
            {budgets.length}
          </span>
        </button>

        <button
          onClick={() => setActiveTab("lines")}
          className={`flex items-center gap-2 px-4 py-2.5 rounded-xl text-sm font-medium transition-all ${
            activeTab === "lines"
              ? "bg-primary text-primary-foreground shadow-sm"
              : "text-muted-foreground hover:bg-muted hover:text-foreground"
          }`}
        >
          <Layers className="h-4 w-4" />
          <span>{isAr ? "البنود وتخصيص الأبعاد" : "Line Allocations"}</span>
          <span className="px-1.5 py-0.5 rounded-md text-xs bg-background/20 font-bold">
            {lines.length}
          </span>
        </button>

        <button
          onClick={() => setActiveTab("variance")}
          className={`flex items-center gap-2 px-4 py-2.5 rounded-xl text-sm font-medium transition-all ${
            activeTab === "variance"
              ? "bg-primary text-primary-foreground shadow-sm"
              : "text-muted-foreground hover:bg-muted hover:text-foreground"
          }`}
        >
          <BarChart3 className="h-4 w-4" />
          <span>{isAr ? "الميزانية مقابل الفعلي (Variance)" : "Budget vs Actual"}</span>
        </button>

        <button
          onClick={() => setActiveTab("transfers")}
          className={`flex items-center gap-2 px-4 py-2.5 rounded-xl text-sm font-medium transition-all ${
            activeTab === "transfers"
              ? "bg-primary text-primary-foreground shadow-sm"
              : "text-muted-foreground hover:bg-muted hover:text-foreground"
          }`}
        >
          <ArrowRightLeft className="h-4 w-4" />
          <span>{isAr ? "سجل المناقلات المالية" : "Fund Reallocations"}</span>
          <span className="px-1.5 py-0.5 rounded-md text-xs bg-background/20 font-bold">
            {transfers.length}
          </span>
        </button>
      </div>

      {/* TAB 1: Budgets Master */}
      {activeTab === "budgets" && (
        <div className="space-y-4">
          {/* Controls Bar */}
          <div className="flex flex-col sm:flex-row items-center justify-between gap-3 bg-card p-3.5 rounded-xl border">
            <div className="relative w-full sm:w-80">
              <Search className="absolute right-3 top-2.5 h-4 w-4 text-muted-foreground" />
              <input
                type="text"
                placeholder={isAr ? "بحث بالرمز أو اسم الميزانية..." : "Search by code or budget name..."}
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pr-9 pl-4 py-1.5 text-sm bg-background border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
              />
            </div>

            <div className="flex items-center gap-2 w-full sm:w-auto">
              <Filter className="h-4 w-4 text-muted-foreground" />
              <select
                value={filterType}
                onChange={(e) => setFilterType(e.target.value)}
                className="bg-background border rounded-lg px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
              >
                <option value="ALL">{isAr ? "كافة الأنواع" : "All Types"}</option>
                <option value="OPERATING">{isAr ? "تشغيلية (OPEX)" : "Operating"}</option>
                <option value="CAPITAL">{isAr ? "رأسمالية (CAPEX)" : "Capital"}</option>
                <option value="PROJECT">{isAr ? "مشاريع خاصة" : "Project"}</option>
              </select>
            </div>
          </div>

          {/* Budgets Table */}
          <div className="border rounded-2xl overflow-hidden bg-card shadow-sm">
            <div className="overflow-x-auto">
              <table className="w-full text-sm text-right">
                <thead className="bg-muted/60 text-muted-foreground text-xs uppercase font-medium border-b">
                  <tr>
                    <th className="px-4 py-3.5">{isAr ? "رمز الميزانية" : "Code"}</th>
                    <th className="px-4 py-3.5">{isAr ? "اسم الميزانية" : "Budget Name"}</th>
                    <th className="px-4 py-3.5">{isAr ? "النوع" : "Type"}</th>
                    <th className="px-4 py-3.5">{isAr ? "المبلغ المعتمد" : "Budgeted (SAR)"}</th>
                    <th className="px-4 py-3.5">{isAr ? "المنصرف الفعلي" : "Actual (SAR)"}</th>
                    <th className="px-4 py-3.5">{isAr ? "الرصيد المتاح" : "Remaining"}</th>
                    <th className="px-4 py-3.5">{isAr ? "الحالة" : "Status"}</th>
                    <th className="px-4 py-3.5 text-center">{isAr ? "الإجراءات" : "Actions"}</th>
                  </tr>
                </thead>
                <tbody className="divide-y">
                  {filteredBudgets.map((b) => {
                    const remaining = b.totalBudgeted - b.totalActual;
                    const usage = b.totalBudgeted > 0 ? ((b.totalActual / b.totalBudgeted) * 100).toFixed(0) : "0";
                    return (
                      <tr key={b.id} className="hover:bg-muted/40 transition-colors">
                        <td className="px-4 py-3.5 font-mono font-semibold text-primary">
                          {b.code}
                        </td>
                        <td className="px-4 py-3.5">
                          <div className="font-medium text-foreground">{isAr ? b.nameAr : b.nameEn}</div>
                          <div className="text-xs text-muted-foreground line-clamp-1">{isAr ? b.descriptionAr : b.descriptionEn}</div>
                        </td>
                        <td className="px-4 py-3.5">
                          <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-secondary text-secondary-foreground border">
                            {b.budgetType}
                          </span>
                        </td>
                        <td className="px-4 py-3.5 font-semibold text-foreground">
                          {b.totalBudgeted.toLocaleString()} SAR
                        </td>
                        <td className="px-4 py-3.5 text-blue-600 dark:text-blue-400 font-medium">
                          {b.totalActual.toLocaleString()} SAR
                          <span className="text-[10px] text-muted-foreground block">({usage}%)</span>
                        </td>
                        <td className="px-4 py-3.5 text-emerald-600 dark:text-emerald-400 font-semibold">
                          {remaining.toLocaleString()} SAR
                        </td>
                        <td className="px-4 py-3.5">
                          {b.status === "ACTIVE" && (
                            <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-emerald-500/10 text-emerald-600 border border-emerald-500/20">
                              <CheckCircle2 className="h-3 w-3" />
                              {isAr ? "معتمدة ونشطة" : "Active"}
                            </span>
                          )}
                          {b.status === "SUBMITTED" && (
                            <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-amber-500/10 text-amber-600 border border-amber-500/20">
                              <Clock className="h-3 w-3" />
                              {isAr ? "بانتظار الاعتماد" : "Submitted"}
                            </span>
                          )}
                          {b.status === "DRAFT" && (
                            <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-muted text-muted-foreground border">
                              {isAr ? "مسودة" : "Draft"}
                            </span>
                          )}
                        </td>
                        <td className="px-4 py-3.5 text-center">
                          <div className="flex items-center justify-center gap-2">
                            {b.status === "DRAFT" && (
                              <button
                                onClick={() => handleSubmitBudget(b.id)}
                                className="px-2.5 py-1 text-xs rounded-lg font-medium bg-secondary hover:bg-muted text-foreground transition-colors border"
                              >
                                {isAr ? "تقديم" : "Submit"}
                              </button>
                            )}
                            {b.status === "SUBMITTED" && (
                              <button
                                onClick={() => handleApproveBudget(b.id)}
                                className="px-2.5 py-1 text-xs rounded-lg font-medium bg-emerald-600 hover:bg-emerald-700 text-white transition-colors shadow-sm"
                              >
                                {isAr ? "اعتماد" : "Approve"}
                              </button>
                            )}
                            <button
                              onClick={() => {
                                setSelectedBudgetId(b.id);
                                setActiveTab("lines");
                              }}
                              className="px-2.5 py-1 text-xs rounded-lg font-medium bg-primary/10 hover:bg-primary/20 text-primary transition-colors"
                            >
                              {isAr ? "عرض البنود" : "View Lines"}
                            </button>
                          </div>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* TAB 2: Budget Lines */}
      {activeTab === "lines" && (
        <div className="space-y-4">
          {/* Header Bar */}
          <div className="flex flex-col sm:flex-row items-center justify-between gap-3 bg-card p-4 rounded-xl border">
            <div className="flex items-center gap-3">
              <div className="p-2.5 rounded-xl bg-primary/10 text-primary">
                <Layers className="h-5 w-5" />
              </div>
              <div>
                <div className="text-xs text-muted-foreground">{isAr ? "الميزانية المحددة حالياً" : "Current Active Budget"}</div>
                <div className="font-bold text-base text-foreground flex items-center gap-2">
                  <span>{isAr ? currentSelectedBudget.nameAr : currentSelectedBudget.nameEn}</span>
                  <span className="font-mono text-xs px-2 py-0.5 rounded bg-muted text-muted-foreground border">
                    {currentSelectedBudget.code}
                  </span>
                </div>
              </div>
            </div>

            <div className="flex items-center gap-2.5">
              <select
                value={selectedBudgetId}
                onChange={(e) => setSelectedBudgetId(e.target.value)}
                className="bg-background border rounded-lg px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
              >
                {budgets.map((b) => (
                  <option key={b.id} value={b.id}>
                    {b.code} - {isAr ? b.nameAr : b.nameEn}
                  </option>
                ))}
              </select>

              <button
                onClick={() => setIsAddLineOpen(true)}
                className="inline-flex items-center gap-2 px-3.5 py-1.5 rounded-lg text-sm font-medium bg-primary text-primary-foreground hover:bg-primary/90 transition-colors shadow-sm"
              >
                <Plus className="h-4 w-4" />
                <span>{isAr ? "إضافة بند مالي" : "Add Line"}</span>
              </button>
            </div>
          </div>

          {/* Lines Table */}
          <div className="border rounded-2xl overflow-hidden bg-card shadow-sm">
            <div className="overflow-x-auto">
              <table className="w-full text-sm text-right">
                <thead className="bg-muted/60 text-muted-foreground text-xs uppercase font-medium border-b">
                  <tr>
                    <th className="px-4 py-3.5">{isAr ? "رقم الحساب" : "Account"}</th>
                    <th className="px-4 py-3.5">{isAr ? "اسم الحساب المالي" : "Account Name"}</th>
                    <th className="px-4 py-3.5">{isAr ? "القسم / الإدارة" : "Department"}</th>
                    <th className="px-4 py-3.5">{isAr ? "مركز التكلفة" : "Cost Center"}</th>
                    <th className="px-4 py-3.5">{isAr ? "الفترة المالية" : "Period"}</th>
                    <th className="px-4 py-3.5">{isAr ? "المخصص المعتمد" : "Budgeted (SAR)"}</th>
                    <th className="px-4 py-3.5">{isAr ? "المنصرف" : "Actual (SAR)"}</th>
                    <th className="px-4 py-3.5">{isAr ? "المتبقي" : "Remaining"}</th>
                  </tr>
                </thead>
                <tbody className="divide-y">
                  {activeBudgetLines.map((l) => {
                    const rem = l.budgetedAmount - l.actualAmount;
                    return (
                      <tr key={l.id} className="hover:bg-muted/40 transition-colors">
                        <td className="px-4 py-3.5 font-mono font-semibold text-primary">
                          {l.accountCode}
                        </td>
                        <td className="px-4 py-3.5">
                          <div className="font-medium text-foreground">{l.accountNameAr}</div>
                          <div className="text-xs text-muted-foreground">{l.descriptionAr}</div>
                        </td>
                        <td className="px-4 py-3.5 text-muted-foreground">{l.departmentNameAr}</td>
                        <td className="px-4 py-3.5 text-muted-foreground">{l.costCenterNameAr}</td>
                        <td className="px-4 py-3.5">
                          <span className="inline-flex items-center px-2 py-0.5 rounded text-xs bg-muted font-medium">
                            {l.periodNameAr}
                          </span>
                        </td>
                        <td className="px-4 py-3.5 font-semibold text-foreground">
                          {l.budgetedAmount.toLocaleString()} SAR
                        </td>
                        <td className="px-4 py-3.5 text-blue-600 dark:text-blue-400 font-medium">
                          {l.actualAmount.toLocaleString()} SAR
                        </td>
                        <td className="px-4 py-3.5 text-emerald-600 dark:text-emerald-400 font-semibold">
                          {rem.toLocaleString()} SAR
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* TAB 3: Budget vs Actual Variance */}
      {activeTab === "variance" && (
        <div className="space-y-4">
          <div className="bg-card p-4 rounded-xl border flex flex-col sm:flex-row items-center justify-between gap-3">
            <div className="flex items-center gap-3">
              <div className="p-2.5 rounded-xl bg-amber-500/10 text-amber-500">
                <BarChart3 className="h-5 w-5" />
              </div>
              <div>
                <h3 className="font-bold text-foreground">{isAr ? "مراقبة الانحرافات والرقابة على الإنفاق" : "Variance & Expenditure Monitor"}</h3>
                <p className="text-xs text-muted-foreground">{isAr ? "مقارنة لحظية للمصروفات الفعلية المقيدة بدفتر الأستاذ مقابل الاعتمادات التقديرية" : "Real-time General Ledger actuals compared to approved budget limits"}</p>
              </div>
            </div>

            <div className="flex items-center gap-3 text-xs">
              <span className="flex items-center gap-1.5"><span className="h-3 w-3 rounded-full bg-emerald-500" /> {isAr ? "ضمن الحدود (<80%)" : "On Track"}</span>
              <span className="flex items-center gap-1.5"><span className="h-3 w-3 rounded-full bg-amber-500" /> {isAr ? "تنبيه اقتراب (80-100%)" : "Warning"}</span>
              <span className="flex items-center gap-1.5"><span className="h-3 w-3 rounded-full bg-rose-500" /> {isAr ? "تجاوز الميزانية (>100%)" : "Over Budget"}</span>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {activeBudgetLines.map((l) => {
              const rate = l.budgetedAmount > 0 ? (l.actualAmount / l.budgetedAmount) * 100 : 0;
              const variance = l.budgetedAmount - l.actualAmount;
              let statusColor = "text-emerald-600 bg-emerald-500/10 border-emerald-500/20";
              let barColor = "bg-emerald-500";
              let labelAr = "ضمن الحدود المسموحة";

              if (rate > 100) {
                statusColor = "text-rose-600 bg-rose-500/10 border-rose-500/20";
                barColor = "bg-rose-500";
                labelAr = "تجاوز في المخصص المعتمد";
              } else if (rate >= 80) {
                statusColor = "text-amber-600 bg-amber-500/10 border-amber-500/20";
                barColor = "bg-amber-500";
                labelAr = "اقتراب من سقف الاعتماد";
              }

              return (
                <div key={l.id} className="p-5 rounded-2xl border bg-card/70 backdrop-blur shadow-sm space-y-3">
                  <div className="flex items-center justify-between">
                    <div>
                      <span className="font-mono text-xs font-bold text-primary px-2 py-0.5 rounded bg-primary/10">
                        {l.accountCode}
                      </span>
                      <h4 className="font-bold text-foreground text-sm mt-1">{l.accountNameAr}</h4>
                      <p className="text-xs text-muted-foreground">{l.departmentNameAr} • {l.costCenterNameAr}</p>
                    </div>

                    <span className={`inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-semibold border ${statusColor}`}>
                      {labelAr}
                    </span>
                  </div>

                  {/* Progress bar */}
                  <div className="space-y-1.5">
                    <div className="flex items-center justify-between text-xs font-semibold">
                      <span className="text-muted-foreground">{isAr ? "نسبة الاستهلاك:" : "Usage:"} {rate.toFixed(1)}%</span>
                      <span className="text-foreground">{l.actualAmount.toLocaleString()} / {l.budgetedAmount.toLocaleString()} SAR</span>
                    </div>
                    <div className="w-full bg-muted rounded-full h-2.5 overflow-hidden">
                      <div className={`${barColor} h-2.5 rounded-full transition-all duration-500`} style={{ width: `${Math.min(rate, 100)}%` }} />
                    </div>
                  </div>

                  {/* Numbers footer */}
                  <div className="grid grid-cols-2 pt-2 border-t text-xs">
                    <div>
                      <span className="text-muted-foreground block">{isAr ? "الارتباطات والمعلقات:" : "Committed:"}</span>
                      <span className="font-semibold text-foreground">{l.committedAmount.toLocaleString()} SAR</span>
                    </div>
                    <div className="text-left">
                      <span className="text-muted-foreground block">{isAr ? "صافي الانحراف المالي:" : "Net Variance:"}</span>
                      <span className={`font-bold ${variance >= 0 ? "text-emerald-600" : "text-rose-600"}`}>
                        {variance >= 0 ? `+${variance.toLocaleString()}` : variance.toLocaleString()} SAR
                      </span>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* TAB 4: Transfers */}
      {activeTab === "transfers" && (
        <div className="space-y-4">
          <div className="flex items-center justify-between bg-card p-4 rounded-xl border">
            <div>
              <h3 className="font-bold text-foreground">{isAr ? "سجل المناقلات وإعادة التخصيص" : "Budget Transfers & Audit Log"}</h3>
              <p className="text-xs text-muted-foreground">{isAr ? "توثيق شامل لكافة مناقلات الفوائض المالية بين مراكز التكلفة والحسابات الإدارية" : "Immutable ledger of all inter-department and account fund adjustments"}</p>
            </div>
            <button
              onClick={() => setIsTransferOpen(true)}
              className="inline-flex items-center gap-2 px-3.5 py-1.5 rounded-lg text-sm font-medium bg-primary text-primary-foreground hover:bg-primary/90 transition-colors shadow-sm"
            >
              <ArrowRightLeft className="h-4 w-4" />
              <span>{isAr ? "إجراء مناقلة جديدة" : "New Transfer"}</span>
            </button>
          </div>

          <div className="border rounded-2xl overflow-hidden bg-card shadow-sm">
            <div className="overflow-x-auto">
              <table className="w-full text-sm text-right">
                <thead className="bg-muted/60 text-muted-foreground text-xs uppercase font-medium border-b">
                  <tr>
                    <th className="px-4 py-3.5">{isAr ? "رقم المناقلة" : "Transfer #"}</th>
                    <th className="px-4 py-3.5">{isAr ? "من الحساب (المصدر)" : "From Account"}</th>
                    <th className="px-4 py-3.5">{isAr ? "إلى الحساب (الهدف)" : "To Account"}</th>
                    <th className="px-4 py-3.5">{isAr ? "المبلغ المنقول" : "Amount (SAR)"}</th>
                    <th className="px-4 py-3.5">{isAr ? "مبرر المناقلة" : "Business Justification"}</th>
                    <th className="px-4 py-3.5">{isAr ? "التاريخ" : "Timestamp"}</th>
                    <th className="px-4 py-3.5">{isAr ? "الحالة" : "Status"}</th>
                  </tr>
                </thead>
                <tbody className="divide-y">
                  {transfers.map((t) => (
                    <tr key={t.id} className="hover:bg-muted/40 transition-colors">
                      <td className="px-4 py-3.5 font-mono font-semibold text-primary">
                        {t.transferNumber}
                      </td>
                      <td className="px-4 py-3.5">
                        <span className="font-medium text-foreground">{t.sourceAccountNameAr}</span>
                        <span className="text-xs text-muted-foreground block font-mono">{t.sourceAccountCode}</span>
                      </td>
                      <td className="px-4 py-3.5">
                        <span className="font-medium text-foreground">{t.targetAccountNameAr}</span>
                        <span className="text-xs text-muted-foreground block font-mono">{t.targetAccountCode}</span>
                      </td>
                      <td className="px-4 py-3.5 font-bold text-foreground">
                        {t.amount.toLocaleString()} SAR
                      </td>
                      <td className="px-4 py-3.5 text-xs text-muted-foreground max-w-xs">
                        {t.reason}
                      </td>
                      <td className="px-4 py-3.5 text-xs text-muted-foreground font-mono">
                        {t.createdAt}
                      </td>
                      <td className="px-4 py-3.5">
                        <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-emerald-500/10 text-emerald-600 border border-emerald-500/20">
                          <CheckCircle2 className="h-3 w-3" />
                          {isAr ? "منفذة" : "Completed"}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* MODAL 1: Create Budget */}
      {isCreateOpen && (
        <div className="fixed inset-0 z-50 bg-background/80 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-card border rounded-2xl max-w-lg w-full p-6 shadow-2xl space-y-4 animate-in fade-in zoom-in-95">
            <div className="flex items-center justify-between border-b pb-3">
              <h3 className="text-lg font-bold text-foreground flex items-center gap-2">
                <PieChart className="h-5 w-5 text-primary" />
                {isAr ? "إنشاء ميزانية تقديرية جديدة" : "Create New Budget"}
              </h3>
              <button
                onClick={() => setIsCreateOpen(false)}
                className="text-muted-foreground hover:text-foreground p-1 rounded-lg"
              >
                <X className="h-5 w-5" />
              </button>
            </div>

            <form onSubmit={handleCreateBudget} className="space-y-4">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-semibold text-muted-foreground mb-1">
                    {isAr ? "رمز الميزانية (Code) *" : "Budget Code *"}
                  </label>
                  <input
                    type="text"
                    required
                    placeholder="BUD-2026-IT"
                    value={newBudget.code}
                    onChange={(e) => setNewBudget({ ...newBudget, code: e.target.value })}
                    className="w-full bg-background border rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-primary focus:outline-none"
                  />
                </div>
                <div>
                  <label className="block text-xs font-semibold text-muted-foreground mb-1">
                    {isAr ? "العام المالي" : "Fiscal Year"}
                  </label>
                  <input
                    type="text"
                    disabled
                    value={newBudget.fiscalYear}
                    className="w-full bg-muted border rounded-lg px-3 py-2 text-sm text-muted-foreground"
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {isAr ? "اسم الميزانية باللغة العربية *" : "Budget Name (Arabic) *"}
                </label>
                <input
                  type="text"
                  required
                  placeholder="ميزانية تطوير البنية التحتية والتحول الرقمي"
                  value={newBudget.nameAr}
                  onChange={(e) => setNewBudget({ ...newBudget, nameAr: e.target.value })}
                  className="w-full bg-background border rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-primary focus:outline-none"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {isAr ? "اسم الميزانية بالإنجليزية" : "Budget Name (English)"}
                </label>
                <input
                  type="text"
                  placeholder="Digital Transformation & IT Infra Budget"
                  value={newBudget.nameEn}
                  onChange={(e) => setNewBudget({ ...newBudget, nameEn: e.target.value })}
                  className="w-full bg-background border rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-primary focus:outline-none"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-semibold text-muted-foreground mb-1">
                    {isAr ? "نوع الميزانية" : "Budget Type"}
                  </label>
                  <select
                    value={newBudget.budgetType}
                    onChange={(e) =>
                      setNewBudget({
                        ...newBudget,
                        budgetType: e.target.value as BudgetRecord["budgetType"],
                      })
                    }
                    className="w-full bg-background border rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-primary focus:outline-none"
                  >
                    <option value="OPERATING">{isAr ? "تشغيلية (OPEX)" : "Operating"}</option>
                    <option value="CAPITAL">{isAr ? "رأسمالية (CAPEX)" : "Capital"}</option>
                    <option value="PROJECT">{isAr ? "مشاريع استراتيجية" : "Project"}</option>
                  </select>
                </div>
                <div>
                  <label className="block text-xs font-semibold text-muted-foreground mb-1">
                    {isAr ? "المبلغ التقديري المبدئي (SAR)" : "Initial Budget (SAR)"}
                  </label>
                  <input
                    type="number"
                    placeholder="500000"
                    value={newBudget.totalBudgeted}
                    onChange={(e) => setNewBudget({ ...newBudget, totalBudgeted: e.target.value })}
                    className="w-full bg-background border rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-primary focus:outline-none"
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {isAr ? "الوصف والأهداف الاستراتيجية" : "Description & Strategic Goals"}
                </label>
                <textarea
                  rows={2}
                  placeholder={isAr ? "أهداف الميزانية والخطط الاستثمارية المعتمدة..." : "Budget objectives..."}
                  value={newBudget.descriptionAr}
                  onChange={(e) => setNewBudget({ ...newBudget, descriptionAr: e.target.value })}
                  className="w-full bg-background border rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-primary focus:outline-none"
                />
              </div>

              <div className="flex items-center justify-end gap-2 pt-3 border-t">
                <button
                  type="button"
                  onClick={() => setIsCreateOpen(false)}
                  className="px-4 py-2 text-sm font-medium rounded-lg border hover:bg-muted transition-colors"
                >
                  {isAr ? "إلغاء" : "Cancel"}
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 text-sm font-medium rounded-lg bg-primary text-primary-foreground hover:bg-primary/90 transition-colors"
                >
                  {isAr ? "حفظ الميزانية" : "Save Budget"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL 2: Add Line */}
      {isAddLineOpen && (
        <div className="fixed inset-0 z-50 bg-background/80 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-card border rounded-2xl max-w-lg w-full p-6 shadow-2xl space-y-4 animate-in fade-in zoom-in-95">
            <div className="flex items-center justify-between border-b pb-3">
              <h3 className="text-lg font-bold text-foreground flex items-center gap-2">
                <Layers className="h-5 w-5 text-primary" />
                {isAr ? "إضافة بند مالي متعدد الأبعاد" : "Add Multi-Dimensional Line"}
              </h3>
              <button
                onClick={() => setIsAddLineOpen(false)}
                className="text-muted-foreground hover:text-foreground p-1 rounded-lg"
              >
                <X className="h-5 w-5" />
              </button>
            </div>

            <form onSubmit={handleAddLine} className="space-y-4">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-semibold text-muted-foreground mb-1">
                    {isAr ? "الحساب المحاسبي" : "GL Account"}
                  </label>
                  <select
                    value={newLine.accountCode}
                    onChange={(e) => {
                      const code = e.target.value;
                      let name = "رواتب وأجور";
                      if (code === "5201") name = "إيجار المقار والمباني";
                      if (code === "5301") name = "مصروفات التسويق";
                      if (code === "5401") name = "تراخيص تقنية المعلومات";
                      setNewLine({ ...newLine, accountCode: code, accountNameAr: name });
                    }}
                    className="w-full bg-background border rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-primary focus:outline-none"
                  >
                    <option value="5101">5101 - رواتب وأجور</option>
                    <option value="5201">5201 - إيجار المقار</option>
                    <option value="5301">5301 - التسويق والدعاية</option>
                    <option value="5401">5401 - تراخيص تقنية المعلومات</option>
                  </select>
                </div>

                <div>
                  <label className="block text-xs font-semibold text-muted-foreground mb-1">
                    {isAr ? "القسم المسؤول" : "Department"}
                  </label>
                  <select
                    value={newLine.departmentNameAr}
                    onChange={(e) => setNewLine({ ...newLine, departmentNameAr: e.target.value })}
                    className="w-full bg-background border rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-primary focus:outline-none"
                  >
                    <option value="الموارد البشرية">الموارد البشرية</option>
                    <option value="المالية والإدارية">المالية والإدارية</option>
                    <option value="المبيعات والتسويق">المبيعات والتسويق</option>
                    <option value="تقنية المعلومات">تقنية المعلومات</option>
                  </select>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-semibold text-muted-foreground mb-1">
                    {isAr ? "مركز التكلفة" : "Cost Center"}
                  </label>
                  <select
                    value={newLine.costCenterNameAr}
                    onChange={(e) => setNewLine({ ...newLine, costCenterNameAr: e.target.value })}
                    className="w-full bg-background border rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-primary focus:outline-none"
                  >
                    <option value="الإدارة العامة">الإدارة العامة</option>
                    <option value="مركز مبيعات الرياض">مركز مبيعات الرياض</option>
                    <option value="مركز العمليات">مركز العمليات</option>
                  </select>
                </div>

                <div>
                  <label className="block text-xs font-semibold text-muted-foreground mb-1">
                    {isAr ? "الفترة المالية" : "Fiscal Period"}
                  </label>
                  <select
                    value={newLine.periodNameAr}
                    onChange={(e) => setNewLine({ ...newLine, periodNameAr: e.target.value })}
                    className="w-full bg-background border rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-primary focus:outline-none"
                  >
                    <option value="الربع الأول 2026">الربع الأول 2026</option>
                    <option value="الربع الثاني 2026">الربع الثاني 2026</option>
                    <option value="النصف الأول 2026">النصف الأول 2026</option>
                    <option value="العام الكامل 2026">العام الكامل 2026</option>
                  </select>
                </div>
              </div>

              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {isAr ? "المبلغ المخصص (SAR) *" : "Allocated Amount (SAR) *"}
                </label>
                <input
                  type="number"
                  required
                  placeholder="50000"
                  value={newLine.budgetedAmount}
                  onChange={(e) => setNewLine({ ...newLine, budgetedAmount: e.target.value })}
                  className="w-full bg-background border rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-primary focus:outline-none"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {isAr ? "ملاحظات تفصيلية للبند" : "Line Description"}
                </label>
                <input
                  type="text"
                  placeholder="تفاصيل التخصيص والمشروع المستهدف..."
                  value={newLine.descriptionAr}
                  onChange={(e) => setNewLine({ ...newLine, descriptionAr: e.target.value })}
                  className="w-full bg-background border rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-primary focus:outline-none"
                />
              </div>

              <div className="flex items-center justify-end gap-2 pt-3 border-t">
                <button
                  type="button"
                  onClick={() => setIsAddLineOpen(false)}
                  className="px-4 py-2 text-sm font-medium rounded-lg border hover:bg-muted transition-colors"
                >
                  {isAr ? "إلغاء" : "Cancel"}
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 text-sm font-medium rounded-lg bg-primary text-primary-foreground hover:bg-primary/90 transition-colors"
                >
                  {isAr ? "إضافة البند" : "Add Line"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL 3: Budget Transfer */}
      {isTransferOpen && (
        <div className="fixed inset-0 z-50 bg-background/80 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-card border rounded-2xl max-w-lg w-full p-6 shadow-2xl space-y-4 animate-in fade-in zoom-in-95">
            <div className="flex items-center justify-between border-b pb-3">
              <h3 className="text-lg font-bold text-foreground flex items-center gap-2">
                <ArrowRightLeft className="h-5 w-5 text-primary" />
                {isAr ? "مناقلة مالية بين بنود الميزانية" : "Inter-Line Budget Reallocation"}
              </h3>
              <button
                onClick={() => setIsTransferOpen(false)}
                className="text-muted-foreground hover:text-foreground p-1 rounded-lg"
              >
                <X className="h-5 w-5" />
              </button>
            </div>

            <form onSubmit={handleExecuteTransfer} className="space-y-4">
              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {isAr ? "البند المصدر (الخصم منه) *" : "Source Line (Deducted From) *"}
                </label>
                <select
                  value={transferData.sourceLineId}
                  onChange={(e) => setTransferData({ ...transferData, sourceLineId: e.target.value })}
                  className="w-full bg-background border rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-primary focus:outline-none"
                >
                  {lines.map((l) => (
                    <option key={l.id} value={l.id}>
                      {l.accountCode} - {l.accountNameAr} (المتاح: {(l.budgetedAmount - l.actualAmount).toLocaleString()} SAR)
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {isAr ? "البند الهدف (الزيادة إليه) *" : "Target Line (Added To) *"}
                </label>
                <select
                  value={transferData.targetLineId}
                  onChange={(e) => setTransferData({ ...transferData, targetLineId: e.target.value })}
                  className="w-full bg-background border rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-primary focus:outline-none"
                >
                  {lines.map((l) => (
                    <option key={l.id} value={l.id}>
                      {l.accountCode} - {l.accountNameAr} (الحالي: {l.budgetedAmount.toLocaleString()} SAR)
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {isAr ? "مبلغ المناقلة (SAR) *" : "Transfer Amount (SAR) *"}
                </label>
                <input
                  type="number"
                  required
                  placeholder="10000"
                  value={transferData.amount}
                  onChange={(e) => setTransferData({ ...transferData, amount: e.target.value })}
                  className="w-full bg-background border rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-primary focus:outline-none"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  {isAr ? "سبب ومبرر المناقلة المالي *" : "Business Justification *"}
                </label>
                <textarea
                  required
                  rows={2}
                  placeholder={isAr ? "مبرر المناقلة الإداري والموافقة المرجعية..." : "Reason for budget reallocation..."}
                  value={transferData.reason}
                  onChange={(e) => setTransferData({ ...transferData, reason: e.target.value })}
                  className="w-full bg-background border rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-primary focus:outline-none"
                />
              </div>

              <div className="flex items-center justify-end gap-2 pt-3 border-t">
                <button
                  type="button"
                  onClick={() => setIsTransferOpen(false)}
                  className="px-4 py-2 text-sm font-medium rounded-lg border hover:bg-muted transition-colors"
                >
                  {isAr ? "إلغاء" : "Cancel"}
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 text-sm font-medium rounded-lg bg-primary text-primary-foreground hover:bg-primary/90 transition-colors"
                >
                  {isAr ? "تنفيذ المناقلة" : "Execute Transfer"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
