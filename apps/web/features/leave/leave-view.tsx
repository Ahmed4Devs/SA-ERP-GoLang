"use client";

import React, { useState } from "react";
import {
  Calendar,
  CheckCircle2,
  XCircle,
  Clock,
  Plus,
  Search,
  BookOpen,
  Award,
  Sparkles,
  TrendingUp,
  Building2,
  FileCheck2,
  ShieldCheck,
  AlertCircle,
  FileText,
  HelpCircle,
  Users,
  Activity,
} from "lucide-react";
import { useLocale } from "@/components/layout/shell";

interface LeaveTypeItem {
  id: string;
  code: string;
  nameEn: string;
  nameAr: string;
  isPaid: boolean;
  requiresAttachment: boolean;
  saudiLawArticle: string;
  descriptionAr: string;
  descriptionEn: string;
}

interface LeaveBalanceItem {
  id: string;
  employeeId: string;
  employeeNameEn: string;
  employeeNameAr: string;
  leaveTypeCode: string;
  leaveTypeNameAr: string;
  year: number;
  entitledDays: number;
  accruedDays: number;
  usedDays: number;
  reservedDays: number;
  availableDays: number;
}

interface LeaveRequestItem {
  id: string;
  requestNumber: string;
  employeeId: string;
  employeeNameEn: string;
  employeeNameAr: string;
  leaveTypeCode: string;
  leaveTypeNameAr: string;
  startDate: string;
  endDate: string;
  workingDays: number;
  reason: string;
  status: "PENDING" | "APPROVED" | "REJECTED" | "CANCELLED";
  submittedAt: string;
  decisionNotes?: string;
}

interface LeaveTransactionItem {
  id: string;
  employeeNameEn: string;
  employeeNameAr: string;
  leaveTypeCode: string;
  transactionType: "ACCRUAL" | "USAGE" | "ADJUSTMENT" | "EXPIRATION" | "REVERSAL";
  amountDays: number;
  balanceAfter: number;
  referenceType: string;
  createdAt: string;
}

export function LeaveView() {
  const { t, locale, dir } = useLocale();
  const [activeTab, setActiveTab] = useState<"requests" | "balances" | "policies" | "ledger">("requests");

  // State: 9 Saudi Statutory Leave Types
  const [statutoryTypes] = useState<LeaveTypeItem[]>([
    {
      id: "lt-1",
      code: "ANNUAL",
      nameEn: "Annual Leave",
      nameAr: "الإجازة السنوية",
      isPaid: true,
      requiresAttachment: false,
      saudiLawArticle: "المادة 109",
      descriptionAr: "21 يوماً بأجر كامل وتزداد إلى 30 يوماً إذا أمضى العامل 5 سنوات متصلة في خدمة صاحب العمل.",
      descriptionEn: "21 fully-paid days, increasing to 30 days after 5 continuous years of service.",
    },
    {
      id: "lt-2",
      code: "SICK",
      nameEn: "Sick Leave",
      nameAr: "الإجازة المرضية",
      isPaid: true,
      requiresAttachment: true,
      saudiLawArticle: "المادة 117",
      descriptionAr: "خلال السنة الواحدة: 30 يوماً بأجر كامل، ثم 60 يوماً بثلاثة أرباع الأجر، ثم 30 يوماً دون أجر.",
      descriptionEn: "Per single year: 30 days full pay, 60 days 75% pay, 30 days unpaid.",
    },
    {
      id: "lt-3",
      code: "MATERNITY",
      nameEn: "Maternity Leave",
      nameAr: "إجازة الوضع",
      isPaid: true,
      requiresAttachment: true,
      saudiLawArticle: "المادة 151",
      descriptionAr: "للمرأة العاملة الحق في إجازة وضع بأجر كامل لمدة 10 أسابيع توزعها كيف تشاء قبل أو بعد الوضع.",
      descriptionEn: "10 weeks fully-paid maternity leave for female employees upon child birth.",
    },
    {
      id: "lt-4",
      code: "PATERNITY",
      nameEn: "Paternity Leave",
      nameAr: "إجازة المولود (الأبوة)",
      isPaid: true,
      requiresAttachment: true,
      saudiLawArticle: "المادة 113",
      descriptionAr: "3 أيام بأجر كامل للعامل في حال ولادة مولود له.",
      descriptionEn: "3 days fully-paid leave upon the birth of employee's child.",
    },
    {
      id: "lt-5",
      code: "MARRIAGE",
      nameEn: "Marriage Leave",
      nameAr: "إجازة الزواج",
      isPaid: true,
      requiresAttachment: true,
      saudiLawArticle: "المادة 113",
      descriptionAr: "5 أيام بأجر كامل للعامل في حال زواجه.",
      descriptionEn: "5 days fully-paid leave upon marriage.",
    },
    {
      id: "lt-6",
      code: "BEREAVEMENT",
      nameEn: "Bereavement Leave",
      nameAr: "إجازة الوفاة",
      isPaid: true,
      requiresAttachment: true,
      saudiLawArticle: "المادة 113",
      descriptionAr: "5 أيام بأجر كامل في حال وفاة زوج العامل أو أحد أصوله أو فروعه.",
      descriptionEn: "5 days fully-paid leave in the event of death of spouse, ascendant, or descendant.",
    },
    {
      id: "lt-7",
      code: "HAJJ",
      nameEn: "Hajj Leave",
      nameAr: "إجازة الحج",
      isPaid: true,
      requiresAttachment: true,
      saudiLawArticle: "المادة 114",
      descriptionAr: "إجازة بأجر لا تقل مدتها عن 10 أيام ولا تزيد على 15 يوماً لأداء فريضة الحج لمرة واحدة طوال الخدمة بعد سنتين عمل.",
      descriptionEn: "10-15 paid days for Hajj pilgrimage once throughout service after 2 consecutive years.",
    },
    {
      id: "lt-8",
      code: "EXAM",
      nameEn: "Examination Leave",
      nameAr: "إجازة الامتحانات",
      isPaid: true,
      requiresAttachment: true,
      saudiLawArticle: "المادة 115",
      descriptionAr: "إجازة بأجر كامل لتأدية الامتحانات عن الأيام الفعلية بشرط الانتساب لمؤسسة تعليمية معتمدة.",
      descriptionEn: "Fully paid leave for actual exam days upon enrollment in recognized institution.",
    },
    {
      id: "lt-9",
      code: "UNPAID",
      nameEn: "Unpaid Leave",
      nameAr: "إجازة بدون أجر",
      isPaid: false,
      requiresAttachment: false,
      saudiLawArticle: "المادة 116",
      descriptionAr: "يجوز للعامل الحصول على إجازة دون أجر بموافقة صاحب العمل، ويعد عقد العمل موقوفاً خلالها.",
      descriptionEn: "Unpaid leave subject to employer written agreement; contract is deemed suspended.",
    },
  ]);

  // State: Employee Balances
  const [balances, setBalances] = useState<LeaveBalanceItem[]>([
    {
      id: "bal-1",
      employeeId: "EMP-2025-001",
      employeeNameEn: "Mohammed Al-Otaibi",
      employeeNameAr: "محمد العتيبي",
      leaveTypeCode: "ANNUAL",
      leaveTypeNameAr: "الإجازة السنوية",
      year: 2025,
      entitledDays: 21.0,
      accruedDays: 8.75,
      usedDays: 0,
      reservedDays: 5.0,
      availableDays: 3.75,
    },
    {
      id: "bal-2",
      employeeId: "EMP-2025-002",
      employeeNameEn: "Sarah Al-Ghamdi",
      employeeNameAr: "سارة الغامدي",
      leaveTypeCode: "ANNUAL",
      leaveTypeNameAr: "الإجازة السنوية",
      year: 2025,
      entitledDays: 30.0,
      accruedDays: 12.5,
      usedDays: 4.0,
      reservedDays: 0,
      availableDays: 8.5,
    },
    {
      id: "bal-3",
      employeeId: "EMP-2025-003",
      employeeNameEn: "Tariq Mansour",
      employeeNameAr: "طارق منصور",
      leaveTypeCode: "ANNUAL",
      leaveTypeNameAr: "الإجازة السنوية",
      year: 2025,
      entitledDays: 21.0,
      accruedDays: 8.75,
      usedDays: 2.0,
      reservedDays: 0,
      availableDays: 6.75,
    },
  ]);

  // State: Leave Requests
  const [requests, setRequests] = useState<LeaveRequestItem[]>([
    {
      id: "req-1",
      requestNumber: "LR-2025-0001",
      employeeId: "EMP-2025-001",
      employeeNameEn: "Mohammed Al-Otaibi",
      employeeNameAr: "محمد العتيبي",
      leaveTypeCode: "ANNUAL",
      leaveTypeNameAr: "الإجازة السنوية",
      startDate: "2025-06-01",
      endDate: "2025-06-05",
      workingDays: 5,
      reason: "Family travel during school break / سفر عائلي",
      status: "PENDING",
      submittedAt: "2025-05-15",
    },
    {
      id: "req-2",
      requestNumber: "LR-2025-0002",
      employeeId: "EMP-2025-002",
      employeeNameEn: "Sarah Al-Ghamdi",
      employeeNameAr: "سارة الغامدي",
      leaveTypeCode: "SICK",
      leaveTypeNameAr: "الإجازة المرضية (المادة 117)",
      startDate: "2025-04-10",
      endDate: "2025-04-13",
      workingDays: 4,
      reason: "Sehhaty medical certificate attached / تقرير منصة صحتي",
      status: "APPROVED",
      submittedAt: "2025-04-09",
      decisionNotes: "Sehhaty medical report verified / تم التحقق من تقرير صحتي",
    },
    {
      id: "req-3",
      requestNumber: "LR-2025-0003",
      employeeId: "EMP-2025-003",
      employeeNameEn: "Tariq Mansour",
      employeeNameAr: "طارق منصور",
      leaveTypeCode: "PATERNITY",
      leaveTypeNameAr: "إجازة المولود (3 أيام)",
      startDate: "2025-03-02",
      endDate: "2025-03-04",
      workingDays: 3,
      reason: "Birth certificate attached / شهادة ميلاد المولود",
      status: "APPROVED",
      submittedAt: "2025-03-01",
      decisionNotes: "Birth certificate verified / معتمد وفق المادة 113",
    },
  ]);

  // State: Transaction Ledger
  const [transactions, setTransactions] = useState<LeaveTransactionItem[]>([
    {
      id: "tx-1",
      employeeNameEn: "Mohammed Al-Otaibi",
      employeeNameAr: "محمد العتيبي",
      leaveTypeCode: "ANNUAL",
      transactionType: "ACCRUAL",
      amountDays: 1.75,
      balanceAfter: 8.75,
      referenceType: "Monthly Accrual May 2025 / استحقاق شهر مايو",
      createdAt: "2025-05-01",
    },
    {
      id: "tx-2",
      employeeNameEn: "Sarah Al-Ghamdi",
      employeeNameAr: "سارة الغامدي",
      leaveTypeCode: "SICK",
      transactionType: "USAGE",
      amountDays: -4.0,
      balanceAfter: 26.0,
      referenceType: "Request LR-2025-0002 (Sehhaty)",
      createdAt: "2025-04-10",
    },
    {
      id: "tx-3",
      employeeNameEn: "Tariq Mansour",
      employeeNameAr: "طارق منصور",
      leaveTypeCode: "PATERNITY",
      transactionType: "USAGE",
      amountDays: -3.0,
      balanceAfter: 0,
      referenceType: "Request LR-2025-0003 (Article 113)",
      createdAt: "2025-03-02",
    },
  ]);

  // Search filter
  const [searchTerm, setSearchTerm] = useState("");

  // Modals state
  const [showNewRequestModal, setShowNewRequestModal] = useState(false);
  const [newRequestData, setNewRequestData] = useState({
    employeeId: "EMP-2025-001",
    leaveTypeCode: "ANNUAL",
    startDate: "2025-06-15",
    endDate: "2025-06-19",
    reason: "",
  });

  const [showAdjustBalanceModal, setShowAdjustBalanceModal] = useState(false);
  const [adjustData, setAdjustData] = useState({
    employeeId: "EMP-2025-001",
    leaveTypeCode: "ANNUAL",
    adjustmentDays: 2.0,
    reason: "",
  });

  // Calculate working days skipping Friday & Saturday
  const calculateWorkingDays = (startStr: string, endStr: string): number => {
    const start = new Date(startStr);
    const end = new Date(endStr);
    if (end < start) return 0;
    let days = 0;
    const cur = new Date(start);
    while (cur <= end) {
      const day = cur.getDay(); // 5 = Friday, 6 = Saturday
      if (day !== 5 && day !== 6) {
        days++;
      }
      cur.setDate(cur.getDate() + 1);
    }
    return days;
  };

  // Submit Leave Request
  const handleSubmitRequest = (e: React.FormEvent) => {
    e.preventDefault();
    const workingDays = calculateWorkingDays(newRequestData.startDate, newRequestData.endDate);
    const emp = balances.find((b) => b.employeeId === newRequestData.employeeId) || {
      employeeNameEn: "Mohammed Al-Otaibi",
      employeeNameAr: "محمد العتيبي",
    };
    const lt = statutoryTypes.find((t) => t.code === newRequestData.leaveTypeCode) || {
      nameAr: "الإجازة السنوية",
    };

    const newReq: LeaveRequestItem = {
      id: `req-${Date.now()}`,
      requestNumber: `LR-2025-${Math.floor(1000 + Math.random() * 9000)}`,
      employeeId: newRequestData.employeeId,
      employeeNameEn: emp.employeeNameEn,
      employeeNameAr: emp.employeeNameAr,
      leaveTypeCode: newRequestData.leaveTypeCode,
      leaveTypeNameAr: lt.nameAr,
      startDate: newRequestData.startDate,
      endDate: newRequestData.endDate,
      workingDays: workingDays || 1,
      reason: newRequestData.reason,
      status: "PENDING",
      submittedAt: new Date().toISOString().split("T")[0],
    };

    setRequests([newReq, ...requests]);

    // Update reserved balance
    setBalances(
      balances.map((b) =>
        b.employeeId === newRequestData.employeeId && b.leaveTypeCode === newRequestData.leaveTypeCode
          ? {
              ...b,
              reservedDays: b.reservedDays + (workingDays || 1),
              availableDays: Math.max(0, b.availableDays - (workingDays || 1)),
            }
          : b
      )
    );

    setShowNewRequestModal(false);
  };

  // Approve / Reject Decision
  const handleDecision = (reqId: string, decision: "APPROVED" | "REJECTED") => {
    const target = requests.find((r) => r.id === reqId);
    if (!target) return;

    setRequests(
      requests.map((r) =>
        r.id === reqId
          ? {
              ...r,
              status: decision,
              decisionNotes:
                decision === "APPROVED"
                  ? locale === "ar"
                    ? "معتمد من إدارة الموارد البشرية"
                    : "Approved by HR Department"
                  : locale === "ar"
                  ? "مرفوض لعدم اكتمال المتطلبات"
                  : "Rejected due to operational constraints",
            }
          : r
      )
    );

    if (decision === "APPROVED") {
      // Deduct from used, release reserved
      setBalances(
        balances.map((b) =>
          b.employeeId === target.employeeId && b.leaveTypeCode === target.leaveTypeCode
            ? {
                ...b,
                usedDays: b.usedDays + target.workingDays,
                reservedDays: Math.max(0, b.reservedDays - target.workingDays),
              }
            : b
        )
      );

      // Append to transaction ledger
      const tx: LeaveTransactionItem = {
        id: `tx-${Date.now()}`,
        employeeNameEn: target.employeeNameEn,
        employeeNameAr: target.employeeNameAr,
        leaveTypeCode: target.leaveTypeCode,
        transactionType: "USAGE",
        amountDays: -target.workingDays,
        balanceAfter: 2.75,
        referenceType: `Approved Request ${target.requestNumber}`,
        createdAt: new Date().toISOString().split("T")[0],
      };
      setTransactions([tx, ...transactions]);
    } else {
      // Revert reserved
      setBalances(
        balances.map((b) =>
          b.employeeId === target.employeeId && b.leaveTypeCode === target.leaveTypeCode
            ? {
                ...b,
                reservedDays: Math.max(0, b.reservedDays - target.workingDays),
                availableDays: b.availableDays + target.workingDays,
              }
            : b
        )
      );
    }
  };

  // Adjust Balance
  const handleAdjustBalance = (e: React.FormEvent) => {
    e.preventDefault();
    const emp = balances.find((b) => b.employeeId === adjustData.employeeId) || {
      employeeNameEn: "Mohammed Al-Otaibi",
      employeeNameAr: "محمد العتيبي",
    };

    setBalances(
      balances.map((b) =>
        b.employeeId === adjustData.employeeId && b.leaveTypeCode === adjustData.leaveTypeCode
          ? {
              ...b,
              accruedDays: b.accruedDays + Number(adjustData.adjustmentDays),
              availableDays: b.availableDays + Number(adjustData.adjustmentDays),
            }
          : b
      )
    );

    const tx: LeaveTransactionItem = {
      id: `tx-${Date.now()}`,
      employeeNameEn: emp.employeeNameEn,
      employeeNameAr: emp.employeeNameAr,
      leaveTypeCode: adjustData.leaveTypeCode,
      transactionType: "ADJUSTMENT",
      amountDays: Number(adjustData.adjustmentDays),
      balanceAfter: 10.75,
      referenceType: `Manual Adjustment: ${adjustData.reason}`,
      createdAt: new Date().toISOString().split("T")[0],
    };
    setTransactions([tx, ...transactions]);
    setShowAdjustBalanceModal(false);
  };

  return (
    <div className="space-y-6 pb-12">
      {/* Header Banner */}
      <div className="relative overflow-hidden rounded-2xl bg-gradient-to-r from-emerald-950 via-teal-900 to-slate-900 p-6 md:p-8 text-white shadow-xl">
        <div className="relative z-10 flex flex-col md:flex-row md:items-center md:justify-between gap-4">
          <div>
            <div className="flex items-center gap-2 mb-2">
              <span className="px-3 py-1 rounded-full text-xs font-semibold bg-emerald-500/20 text-emerald-200 border border-emerald-400/30 flex items-center gap-1.5">
                <Calendar className="h-3.5 w-3.5 text-emerald-400" />
                {locale === "ar" ? "المرحلة 20: الإجازات والأرصدة" : "Phase 20: Leave & Balances"}
              </span>
              <span className="px-3 py-1 rounded-full text-xs font-semibold bg-blue-500/20 text-blue-200 border border-blue-400/30 flex items-center gap-1">
                <ShieldCheck className="h-3.5 w-3.5 text-blue-300" />
                {locale === "ar" ? "9 أنواع نظامية معتمدة" : "9 Saudi Statutory Types"}
              </span>
            </div>
            <h1 className="text-2xl md:text-3xl font-bold tracking-tight">
              {(t as any).leave?.title || "إدارة الإجازات والأرصدة النظامية"}
            </h1>
            <p className="mt-2 text-sm md:text-base text-emerald-200/80 max-w-3xl">
              {(t as any).leave?.description ||
                "إدارة الإجازات التسع النظامية وفق نظام العمل السعودي، تسوية الأرصدة وسجل الحركات غير القابل للحذف"}
            </p>
          </div>

          <div className="flex flex-wrap items-center gap-3">
            <button
              onClick={() => setShowNewRequestModal(true)}
              className="px-4 py-2.5 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white font-medium text-sm flex items-center gap-2 shadow-lg shadow-emerald-950/40 transition-all cursor-pointer"
            >
              <Plus className="h-4 w-4" />
              {(t as any).leave?.btn_submit_request || "تقديم طلب إجازة"}
            </button>
            <button
              onClick={() => setShowAdjustBalanceModal(true)}
              className="px-4 py-2.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-white font-medium text-sm flex items-center gap-2 border border-slate-700 transition-all cursor-pointer"
            >
              <FileCheck2 className="h-4 w-4" />
              {(t as any).leave?.btn_adjust_balance || "تسوية رصيد"}
            </button>
          </div>
        </div>

        {/* Decorative background glow */}
        <div className="absolute -bottom-10 -right-10 w-64 h-64 bg-emerald-500/10 rounded-full blur-3xl pointer-events-none" />
        <div className="absolute top-0 -left-10 w-64 h-64 bg-teal-500/10 rounded-full blur-3xl pointer-events-none" />
      </div>

      {/* KPI Stats Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* KPI 1: Statutory Types */}
        <div className="p-5 rounded-xl bg-card border border-border/80 shadow-sm hover:shadow-md transition-shadow">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">
              {(t as any).leave?.kpi_statutory_types || "أنواع الإجازات النظامية"}
            </span>
            <div className="p-2.5 rounded-lg bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
              <BookOpen className="h-5 w-5" />
            </div>
          </div>
          <div className="mt-3">
            <div className="text-2xl font-bold">9</div>
            <div className="text-xs text-muted-foreground mt-1">
              {locale === "ar" ? "سنوية، مرضية، وضع، أبوة، زواج، وفاة، حج، امتحانات، بدون أجر" : "Annual, Sick, Maternity, Paternity, etc."}
            </div>
          </div>
        </div>

        {/* KPI 2: Pending Requests */}
        <div className="p-5 rounded-xl bg-card border border-border/80 shadow-sm hover:shadow-md transition-shadow">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">
              {(t as any).leave?.kpi_pending_requests || "طلبات معلقة بانتظار الاعتماد"}
            </span>
            <div className="p-2.5 rounded-lg bg-amber-500/10 text-amber-600 dark:text-amber-400">
              <Clock className="h-5 w-5" />
            </div>
          </div>
          <div className="mt-3">
            <div className="text-2xl font-bold text-amber-600 dark:text-amber-400">
              {requests.filter((r) => r.status === "PENDING").length}
            </div>
            <div className="text-xs text-muted-foreground mt-1">
              {locale === "ar" ? "تتطلب مراجعة واعتماد الموارد البشرية" : "Awaiting HR decision"}
            </div>
          </div>
        </div>

        {/* KPI 3: Approved This Month */}
        <div className="p-5 rounded-xl bg-card border border-border/80 shadow-sm hover:shadow-md transition-shadow">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">
              {(t as any).leave?.kpi_approved_leaves || "إجازات معتمدة هذا الشهر"}
            </span>
            <div className="p-2.5 rounded-lg bg-blue-500/10 text-blue-600 dark:text-blue-400">
              <CheckCircle2 className="h-5 w-5" />
            </div>
          </div>
          <div className="mt-3">
            <div className="text-2xl font-bold text-blue-600 dark:text-blue-400">
              {requests.filter((r) => r.status === "APPROVED").length}
            </div>
            <div className="text-xs text-muted-foreground mt-1">
              {locale === "ar" ? "تم خصمها من الرصيد وترحيلها للأجور" : "Deducted & posted to ledger"}
            </div>
          </div>
        </div>

        {/* KPI 4: Total Accrued Balance */}
        <div className="p-5 rounded-xl bg-card border border-border/80 shadow-sm hover:shadow-md transition-shadow">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">
              {(t as any).leave?.kpi_accrued_balances || "إجمالي الأرصدة المتاحة"}
            </span>
            <div className="p-2.5 rounded-lg bg-purple-500/10 text-purple-600 dark:text-purple-400">
              <TrendingUp className="h-5 w-5" />
            </div>
          </div>
          <div className="mt-3">
            <div className="text-2xl font-bold text-purple-600 dark:text-purple-400">
              {balances.reduce((acc, curr) => acc + curr.availableDays, 0).toFixed(1)}{" "}
              <span className="text-sm font-normal text-muted-foreground">{locale === "ar" ? "يوماً" : "days"}</span>
            </div>
            <div className="text-xs text-muted-foreground mt-1">
              {locale === "ar" ? "رصيد فعلي متاح للاستخدام" : "Current available leave days"}
            </div>
          </div>
        </div>
      </div>

      {/* Tabs Navigation */}
      <div className="flex border-b border-border space-x-1 sm:space-x-2">
        <button
          onClick={() => setActiveTab("requests")}
          className={`px-4 py-2.5 text-sm font-medium border-b-2 transition-colors flex items-center gap-2 cursor-pointer ${
            activeTab === "requests"
              ? "border-emerald-600 text-emerald-600 dark:text-emerald-400"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <Calendar className="h-4 w-4" />
          {(t as any).leave?.tab_requests || "طلبات الإجازة"}
        </button>

        <button
          onClick={() => setActiveTab("balances")}
          className={`px-4 py-2.5 text-sm font-medium border-b-2 transition-colors flex items-center gap-2 cursor-pointer ${
            activeTab === "balances"
              ? "border-emerald-600 text-emerald-600 dark:text-emerald-400"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <Award className="h-4 w-4" />
          {(t as any).leave?.tab_balances || "أرصدة الموظفين"}
        </button>

        <button
          onClick={() => setActiveTab("policies")}
          className={`px-4 py-2.5 text-sm font-medium border-b-2 transition-colors flex items-center gap-2 cursor-pointer ${
            activeTab === "policies"
              ? "border-emerald-600 text-emerald-600 dark:text-emerald-400"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <BookOpen className="h-4 w-4" />
          {(t as any).leave?.tab_policies || "الأنواع والسياسات النظامية (9)"}
        </button>

        <button
          onClick={() => setActiveTab("ledger")}
          className={`px-4 py-2.5 text-sm font-medium border-b-2 transition-colors flex items-center gap-2 cursor-pointer ${
            activeTab === "ledger"
              ? "border-emerald-600 text-emerald-600 dark:text-emerald-400"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <Activity className="h-4 w-4" />
          {(t as any).leave?.tab_ledger || "سجل حركات الأرصدة (Audit Ledger)"}
        </button>
      </div>

      {/* TAB 1: LEAVE REQUESTS */}
      {activeTab === "requests" && (
        <div className="space-y-4">
          <div className="flex flex-col sm:flex-row items-center justify-between gap-3">
            <div className="relative w-full sm:w-80">
              <Search className="absolute right-3 top-2.5 h-4 w-4 text-muted-foreground pointer-events-none" />
              <input
                type="text"
                placeholder={locale === "ar" ? "بحث برقم الطلب أو الموظف..." : "Search by request # or name..."}
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-3 pr-9 py-2 rounded-xl border border-input bg-card text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
              />
            </div>
            <button
              onClick={() => setShowNewRequestModal(true)}
              className="px-4 py-2 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white text-sm font-medium flex items-center gap-2 shadow transition-all cursor-pointer w-full sm:w-auto justify-center"
            >
              <Plus className="h-4 w-4" />
              {(t as any).leave?.btn_submit_request || "تقديم طلب إجازة"}
            </button>
          </div>

          <div className="overflow-x-auto rounded-xl border border-border bg-card shadow-sm">
            <table className="w-full text-sm text-right">
              <thead className="bg-muted/50 text-muted-foreground text-xs uppercase font-medium">
                <tr>
                  <th className="px-4 py-3">{(t as any).leave?.col_request_num || "رقم الطلب"}</th>
                  <th className="px-4 py-3">{(t as any).leave?.col_emp_name || "اسم الموظف"}</th>
                  <th className="px-4 py-3">{(t as any).leave?.col_leave_type || "نوع الإجازة"}</th>
                  <th className="px-4 py-3">{(t as any).leave?.col_start_date || "تاريخ البداية"}</th>
                  <th className="px-4 py-3">{(t as any).leave?.col_end_date || "تاريخ النهاية"}</th>
                  <th className="px-4 py-3">{(t as any).leave?.col_days || "أيام العمل"}</th>
                  <th className="px-4 py-3">{(t as any).leave?.col_status || "الحالة"}</th>
                  <th className="px-4 py-3">{(t as any).leave?.col_actions || "الإجراءات"}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {requests
                  .filter(
                    (r) =>
                      r.requestNumber.toLowerCase().includes(searchTerm.toLowerCase()) ||
                      r.employeeNameAr.includes(searchTerm) ||
                      r.employeeNameEn.toLowerCase().includes(searchTerm.toLowerCase())
                  )
                  .map((req) => (
                    <tr key={req.id} className="hover:bg-muted/30 transition-colors">
                      <td className="px-4 py-3.5 font-mono text-xs font-bold text-emerald-600 dark:text-emerald-400">
                        {req.requestNumber}
                      </td>
                      <td className="px-4 py-3.5 font-medium">
                        <div>{locale === "ar" ? req.employeeNameAr : req.employeeNameEn}</div>
                        <div className="text-xs text-muted-foreground font-mono">{req.employeeId}</div>
                      </td>
                      <td className="px-4 py-3.5">
                        <span className="px-2 py-0.5 rounded-full text-xs font-semibold bg-muted">
                          {req.leaveTypeNameAr}
                        </span>
                      </td>
                      <td className="px-4 py-3.5 font-mono text-xs">{req.startDate}</td>
                      <td className="px-4 py-3.5 font-mono text-xs">{req.endDate}</td>
                      <td className="px-4 py-3.5 font-bold">
                        {req.workingDays} {locale === "ar" ? "أيام" : "days"}
                        <span className="block text-[10px] text-muted-foreground font-normal">
                          {locale === "ar" ? "(باستثناء الجمعة والسبت)" : "(Excl. Fri/Sat)"}
                        </span>
                      </td>
                      <td className="px-4 py-3.5">
                        <span
                          className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-semibold ${
                            req.status === "APPROVED"
                              ? "bg-emerald-500/15 text-emerald-600 dark:text-emerald-400"
                              : req.status === "PENDING"
                              ? "bg-amber-500/15 text-amber-600 dark:text-amber-400"
                              : "bg-rose-500/15 text-rose-600 dark:text-rose-400"
                          }`}
                        >
                          {req.status === "APPROVED" && (locale === "ar" ? "معتمد" : "Approved")}
                          {req.status === "PENDING" && (locale === "ar" ? "معلق" : "Pending")}
                          {req.status === "REJECTED" && (locale === "ar" ? "مرفوض" : "Rejected")}
                        </span>
                      </td>
                      <td className="px-4 py-3.5">
                        {req.status === "PENDING" ? (
                          <div className="flex items-center gap-1.5">
                            <button
                              onClick={() => handleDecision(req.id, "APPROVED")}
                              className="px-2.5 py-1 rounded-lg bg-emerald-500/10 hover:bg-emerald-500/20 text-emerald-600 dark:text-emerald-400 text-xs font-medium flex items-center gap-1 cursor-pointer transition-colors"
                            >
                              <CheckCircle2 className="h-3.5 w-3.5" />
                              {(t as any).leave?.btn_approve || "اعتماد"}
                            </button>
                            <button
                              onClick={() => handleDecision(req.id, "REJECTED")}
                              className="px-2.5 py-1 rounded-lg bg-rose-500/10 hover:bg-rose-500/20 text-rose-600 dark:text-rose-400 text-xs font-medium flex items-center gap-1 cursor-pointer transition-colors"
                            >
                              <XCircle className="h-3.5 w-3.5" />
                              {(t as any).leave?.btn_reject || "رفض"}
                            </button>
                          </div>
                        ) : (
                          <span className="text-xs text-muted-foreground">{req.decisionNotes || "--"}</span>
                        )}
                      </td>
                    </tr>
                  ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* TAB 2: EMPLOYEE BALANCES */}
      {activeTab === "balances" && (
        <div className="space-y-4">
          <div className="flex justify-between items-center">
            <div>
              <h3 className="text-base font-semibold">
                {locale === "ar" ? "سجل أرصدة الإجازات السنوية المستحقة والمستخدمة" : "Employee Leave Balances & Accruals"}
              </h3>
              <p className="text-xs text-muted-foreground mt-0.5">
                {locale === "ar"
                  ? "احتساب آلي شهري للاستحقاق (Accrual) مع حجز الرصيد عند تقديم الطلب والخصم عند الاعتماد"
                  : "Automatic monthly accrual with reservation on submission and deduction upon approval"}
              </p>
            </div>
            <button
              onClick={() => setShowAdjustBalanceModal(true)}
              className="px-4 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-white text-sm font-medium flex items-center gap-2 border border-slate-700 transition-all cursor-pointer"
            >
              <FileCheck2 className="h-4 w-4" />
              {(t as any).leave?.btn_adjust_balance || "تسوية رصيد"}
            </button>
          </div>

          <div className="overflow-x-auto rounded-xl border border-border bg-card shadow-sm">
            <table className="w-full text-sm text-right">
              <thead className="bg-muted/50 text-muted-foreground text-xs uppercase font-medium">
                <tr>
                  <th className="px-4 py-3">{(t as any).leave?.col_emp_name || "اسم الموظف"}</th>
                  <th className="px-4 py-3">{(t as any).leave?.col_leave_type || "نوع الإجازة"}</th>
                  <th className="px-4 py-3">{locale === "ar" ? "السنة المالية" : "Year"}</th>
                  <th className="px-4 py-3">{(t as any).leave?.col_entitled || "الاستحقاق السنوي"}</th>
                  <th className="px-4 py-3">{(t as any).leave?.col_accrued || "المستحق الفعلي"}</th>
                  <th className="px-4 py-3">{(t as any).leave?.col_used || "المستخدم"}</th>
                  <th className="px-4 py-3">{(t as any).leave?.col_reserved || "المحجوز"}</th>
                  <th className="px-4 py-3">{(t as any).leave?.col_balance_avail || "الرصيد المتاح"}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {balances.map((b) => (
                  <tr key={b.id} className="hover:bg-muted/30 transition-colors">
                    <td className="px-4 py-3.5 font-medium">
                      <div>{locale === "ar" ? b.employeeNameAr : b.employeeNameEn}</div>
                      <div className="text-xs text-muted-foreground font-mono">{b.employeeId}</div>
                    </td>
                    <td className="px-4 py-3.5 text-muted-foreground">{b.leaveTypeNameAr}</td>
                    <td className="px-4 py-3.5 font-mono">{b.year}</td>
                    <td className="px-4 py-3.5 font-bold">{b.entitledDays}d</td>
                    <td className="px-4 py-3.5 font-semibold text-blue-600 dark:text-blue-400">
                      {b.accruedDays.toFixed(2)}d
                    </td>
                    <td className="px-4 py-3.5 text-rose-600 dark:text-rose-400 font-semibold">
                      {b.usedDays}d
                    </td>
                    <td className="px-4 py-3.5 text-amber-600 dark:text-amber-400 font-semibold">
                      {b.reservedDays}d
                    </td>
                    <td className="px-4 py-3.5">
                      <span className="px-3 py-1 rounded-full text-xs font-bold bg-emerald-500/15 text-emerald-600 dark:text-emerald-400">
                        {b.availableDays.toFixed(2)} {locale === "ar" ? "يوماً" : "days"}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* TAB 3: STATUTORY POLICIES & TYPES */}
      {activeTab === "policies" && (
        <div className="space-y-4">
          <div>
            <h3 className="text-base font-semibold">
              {locale === "ar"
                ? "أنواع وسياسات الإجازات التسع المعتمدة بموجب نظام العمل السعودي"
                : "The 9 Saudi Labor Law Statutory Leave Types"}
            </h3>
            <p className="text-xs text-muted-foreground mt-0.5">
              {locale === "ar"
                ? "مدمجة بالكامل في النظام ومطابقة للأنظمة واللوائح التنفيذية الصادرة عن وزارة الموارد البشرية والتنمية الاجتماعية"
                : "Fully integrated and compliant with the Saudi Ministry of Human Resources and Social Development regulations"}
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {statutoryTypes.map((type) => (
              <div
                key={type.id}
                className="p-5 rounded-xl border border-border bg-card shadow-sm hover:shadow-md transition-all flex flex-col justify-between"
              >
                <div>
                  <div className="flex items-center justify-between mb-3">
                    <span className="px-2.5 py-1 rounded-full text-xs font-semibold bg-emerald-500/15 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20">
                      {type.saudiLawArticle}
                    </span>
                    <span
                      className={`px-2 py-0.5 rounded-full text-[11px] font-semibold ${
                        type.isPaid
                          ? "bg-blue-500/15 text-blue-600 dark:text-blue-400"
                          : "bg-amber-500/15 text-amber-600 dark:text-amber-400"
                      }`}
                    >
                      {type.isPaid ? (locale === "ar" ? "مدفوعة الأجر" : "Paid") : (locale === "ar" ? "بدون أجر" : "Unpaid")}
                    </span>
                  </div>

                  <h4 className="font-bold text-lg">{locale === "ar" ? type.nameAr : type.nameEn}</h4>
                  <div className="text-xs text-muted-foreground mt-0.5 font-mono">{type.code}</div>

                  <p className="mt-3 text-xs text-muted-foreground leading-relaxed">
                    {locale === "ar" ? type.descriptionAr : type.descriptionEn}
                  </p>
                </div>

                <div className="mt-4 pt-3 border-t border-border flex items-center justify-between text-xs">
                  <span className="text-muted-foreground">
                    {type.requiresAttachment
                      ? locale === "ar"
                        ? "تتطلب مستند إثبات (مرفق)"
                        : "Requires proof document"
                      : locale === "ar"
                      ? "لا تتطلب مرفق إثبات"
                      : "No attachment required"}
                  </span>
                  <span className="text-emerald-600 dark:text-emerald-400 font-semibold flex items-center gap-1">
                    <ShieldCheck className="h-3.5 w-3.5" />
                    {locale === "ar" ? "نظامي معتمد" : "Compliant"}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* TAB 4: TRANSACTION LEDGER */}
      {activeTab === "ledger" && (
        <div className="space-y-4">
          <div>
            <h3 className="text-base font-semibold">
              {locale === "ar" ? "سجل الحركات غير القابل للتعديل (Leave Ledger)" : "Immutable Leave Ledger"}
            </h3>
            <p className="text-xs text-muted-foreground mt-0.5">
              {locale === "ar"
                ? "سجل محاسبي رقمي يسجل كل إضافة أو خصم أو تسوية للأرصدة لضمان الشفافية والمطابقة النظامية"
                : "Append-only transaction ledger recording every accrual, deduction, and reconciliation"}
            </p>
          </div>

          <div className="overflow-x-auto rounded-xl border border-border bg-card shadow-sm">
            <table className="w-full text-sm text-right">
              <thead className="bg-muted/50 text-muted-foreground text-xs uppercase font-medium">
                <tr>
                  <th className="px-4 py-3">{(t as any).leave?.col_emp_name || "اسم الموظف"}</th>
                  <th className="px-4 py-3">{(t as any).leave?.col_leave_type || "نوع الإجازة"}</th>
                  <th className="px-4 py-3">{(t as any).leave?.col_tx_type || "نوع الحركة"}</th>
                  <th className="px-4 py-3">{(t as any).leave?.col_tx_amount || "الأيام"}</th>
                  <th className="px-4 py-3">{(t as any).leave?.col_tx_balance_after || "الرصيد بعد الحركة"}</th>
                  <th className="px-4 py-3">{(t as any).leave?.col_tx_ref || "المرجع النظامي"}</th>
                  <th className="px-4 py-3">{(t as any).leave?.col_tx_date || "تاريخ الحركة"}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {transactions.map((tx) => (
                  <tr key={tx.id} className="hover:bg-muted/30 transition-colors">
                    <td className="px-4 py-3.5 font-medium">
                      {locale === "ar" ? tx.employeeNameAr : tx.employeeNameEn}
                    </td>
                    <td className="px-4 py-3.5 text-muted-foreground font-mono text-xs">{tx.leaveTypeCode}</td>
                    <td className="px-4 py-3.5">
                      <span
                        className={`px-2.5 py-0.5 rounded-full text-xs font-semibold ${
                          tx.transactionType === "ACCRUAL"
                            ? "bg-emerald-500/15 text-emerald-600 dark:text-emerald-400"
                            : tx.transactionType === "USAGE"
                            ? "bg-rose-500/15 text-rose-600 dark:text-rose-400"
                            : "bg-blue-500/15 text-blue-600 dark:text-blue-400"
                        }`}
                      >
                        {tx.transactionType}
                      </span>
                    </td>
                    <td className="px-4 py-3.5 font-mono font-bold">
                      {tx.amountDays > 0 ? `+${tx.amountDays}` : tx.amountDays} {locale === "ar" ? "يوم" : "d"}
                    </td>
                    <td className="px-4 py-3.5 font-mono font-semibold text-primary">
                      {tx.balanceAfter} {locale === "ar" ? "يوم" : "d"}
                    </td>
                    <td className="px-4 py-3.5 text-xs text-muted-foreground">{tx.referenceType}</td>
                    <td className="px-4 py-3.5 font-mono text-xs text-muted-foreground">{tx.createdAt}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* MODAL 1: SUBMIT LEAVE REQUEST */}
      {showNewRequestModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
          <div className="bg-card w-full max-w-lg rounded-2xl border border-border p-6 shadow-2xl space-y-4">
            <h3 className="text-lg font-bold flex items-center gap-2">
              <Calendar className="h-5 w-5 text-emerald-600" />
              {locale === "ar" ? "تقديم طلب إجازة جديد" : "Submit Leave Request"}
            </h3>
            <form onSubmit={handleSubmitRequest} className="space-y-3">
              <div>
                <label className="block text-xs font-medium text-muted-foreground mb-1">
                  {locale === "ar" ? "الموظف" : "Employee"}
                </label>
                <select
                  value={newRequestData.employeeId}
                  onChange={(e) => setNewRequestData({ ...newRequestData, employeeId: e.target.value })}
                  className="w-full px-3 py-2 rounded-xl border border-input bg-background text-sm"
                >
                  <option value="EMP-2025-001">محمد العتيبي (EMP-2025-001) - الرصيد المتاح: 3.75 يوم</option>
                  <option value="EMP-2025-002">سارة الغامدي (EMP-2025-002) - الرصيد المتاح: 8.5 يوم</option>
                  <option value="EMP-2025-003">طارق منصور (EMP-2025-003) - الرصيد المتاح: 6.75 يوم</option>
                </select>
              </div>

              <div>
                <label className="block text-xs font-medium text-muted-foreground mb-1">
                  {locale === "ar" ? "نوع الإجازة" : "Leave Type"}
                </label>
                <select
                  value={newRequestData.leaveTypeCode}
                  onChange={(e) => setNewRequestData({ ...newRequestData, leaveTypeCode: e.target.value })}
                  className="w-full px-3 py-2 rounded-xl border border-input bg-background text-sm"
                >
                  {statutoryTypes.map((type) => (
                    <option key={type.code} value={type.code}>
                      {locale === "ar" ? `${type.nameAr} (${type.saudiLawArticle})` : `${type.nameEn} (${type.code})`}
                    </option>
                  ))}
                </select>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium text-muted-foreground mb-1">
                    {locale === "ar" ? "تاريخ البداية" : "Start Date"}
                  </label>
                  <input
                    type="date"
                    required
                    value={newRequestData.startDate}
                    onChange={(e) => setNewRequestData({ ...newRequestData, startDate: e.target.value })}
                    className="w-full px-3 py-2 rounded-xl border border-input bg-background text-sm font-mono"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-muted-foreground mb-1">
                    {locale === "ar" ? "تاريخ النهاية" : "End Date"}
                  </label>
                  <input
                    type="date"
                    required
                    value={newRequestData.endDate}
                    onChange={(e) => setNewRequestData({ ...newRequestData, endDate: e.target.value })}
                    className="w-full px-3 py-2 rounded-xl border border-input bg-background text-sm font-mono"
                  />
                </div>
              </div>

              <div className="p-3 rounded-xl bg-muted/40 text-xs flex justify-between items-center">
                <span className="text-muted-foreground">
                  {locale === "ar" ? "أيام العمل الفعلية (دون الجمعة والسبت):" : "Actual Working Days (Excl. Fri/Sat):"}
                </span>
                <span className="font-bold text-emerald-600 dark:text-emerald-400">
                  {calculateWorkingDays(newRequestData.startDate, newRequestData.endDate)}{" "}
                  {locale === "ar" ? "أيام عمل" : "working days"}
                </span>
              </div>

              <div>
                <label className="block text-xs font-medium text-muted-foreground mb-1">
                  {locale === "ar" ? "سبب الإجازة والملاحظات" : "Reason & Comments"}
                </label>
                <textarea
                  rows={2}
                  value={newRequestData.reason}
                  onChange={(e) => setNewRequestData({ ...newRequestData, reason: e.target.value })}
                  placeholder={locale === "ar" ? "إجازة دورية مجدولة..." : "Periodic vacation..."}
                  className="w-full px-3 py-2 rounded-xl border border-input bg-background text-sm"
                />
              </div>

              <div className="flex justify-end gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setShowNewRequestModal(false)}
                  className="px-4 py-2 rounded-xl border border-border text-sm hover:bg-muted"
                >
                  {locale === "ar" ? "إلغاء" : "Cancel"}
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white text-sm font-medium"
                >
                  {locale === "ar" ? "تأكيد تقديم الطلب" : "Submit Request"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL 2: ADJUST BALANCE */}
      {showAdjustBalanceModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
          <div className="bg-card w-full max-w-md rounded-2xl border border-border p-6 shadow-2xl space-y-4">
            <h3 className="text-lg font-bold flex items-center gap-2">
              <FileCheck2 className="h-5 w-5 text-primary" />
              {locale === "ar" ? "تسوية رصيد إجازة يدوياً" : "Manual Balance Adjustment"}
            </h3>
            <form onSubmit={handleAdjustBalance} className="space-y-3">
              <div>
                <label className="block text-xs font-medium text-muted-foreground mb-1">
                  {locale === "ar" ? "الموظف" : "Employee"}
                </label>
                <select
                  value={adjustData.employeeId}
                  onChange={(e) => setAdjustData({ ...adjustData, employeeId: e.target.value })}
                  className="w-full px-3 py-2 rounded-xl border border-input bg-background text-sm"
                >
                  <option value="EMP-2025-001">محمد العتيبي (EMP-2025-001)</option>
                  <option value="EMP-2025-002">سارة الغامدي (EMP-2025-002)</option>
                  <option value="EMP-2025-003">طارق منصور (EMP-2025-003)</option>
                </select>
              </div>

              <div>
                <label className="block text-xs font-medium text-muted-foreground mb-1">
                  {locale === "ar" ? "الأيام المراد إضافتها / تعديلها" : "Days to Adjust (+/-)"}
                </label>
                <input
                  type="number"
                  step="0.5"
                  required
                  value={adjustData.adjustmentDays}
                  onChange={(e) => setAdjustData({ ...adjustData, adjustmentDays: Number(e.target.value) })}
                  className="w-full px-3 py-2 rounded-xl border border-input bg-background text-sm font-mono"
                />
              </div>

              <div>
                <label className="block text-xs font-medium text-muted-foreground mb-1">
                  {locale === "ar" ? "سبب التسوية والمرجع" : "Adjustment Reason & Reference"}
                </label>
                <textarea
                  required
                  rows={2}
                  value={adjustData.reason}
                  onChange={(e) => setAdjustData({ ...adjustData, reason: e.target.value })}
                  placeholder={locale === "ar" ? "ترحيل رصيد من سنة سابقة أو تعويض..." : "Carry-over from prior year..."}
                  className="w-full px-3 py-2 rounded-xl border border-input bg-background text-sm"
                />
              </div>

              <div className="flex justify-end gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setShowAdjustBalanceModal(false)}
                  className="px-4 py-2 rounded-xl border border-border text-sm hover:bg-muted"
                >
                  {locale === "ar" ? "إلغاء" : "Cancel"}
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 rounded-xl bg-primary text-primary-foreground text-sm font-medium hover:bg-primary/90"
                >
                  {locale === "ar" ? "حفظ وترحيل للسجل" : "Post Adjustment"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
