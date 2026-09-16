"use client";

import React, { useState } from "react";
import {
  Banknote,
  DollarSign,
  ShieldCheck,
  FileSpreadsheet,
  Plus,
  Search,
  CheckCircle2,
  Clock,
  Send,
  FileCheck,
  Calendar,
  User,
  Building,
  Calculator,
  ArrowRight,
  TrendingUp,
  AlertCircle,
  FileText,
  CreditCard,
  Layers,
  ChevronRight,
  Printer,
  Download,
  Percent,
} from "lucide-react";
import { useLocale } from "@/components/layout/shell";

interface PayrollRunItem {
  id: string;
  runNumber: string;
  periodNameAr: string;
  periodNameEn: string;
  periodMonth: number;
  periodYear: number;
  runDate: string;
  employeeCount: number;
  totalGross: number;
  totalDeductions: number;
  totalNet: number;
  status: "DRAFT" | "CALCULATING" | "CALCULATED" | "APPROVED" | "POSTED" | "PAID";
  wpsFileGenerated: boolean;
  journalEntryNumber?: string;
}

interface PayslipItem {
  id: string;
  payslipNumber: string;
  employeeId: string;
  employeeNameAr: string;
  employeeNameEn: string;
  positionAr: string;
  positionEn: string;
  basicSalary: number;
  housingAllowance: number;
  transportAllowance: number;
  otherAllowances: number;
  overtimeAmount: number;
  grossSalary: number;
  gosiDeduction: number;
  loanDeduction: number;
  absenceDeduction: number;
  totalDeductions: number;
  netSalary: number;
  issueDate: string;
  iban: string;
  bankNameAr: string;
}

interface ComponentItem {
  code: string;
  nameAr: string;
  nameEn: string;
  type: "EARNING" | "DEDUCTION";
  calculationType: "FIXED" | "PERCENTAGE" | "FORMULA";
  isTaxable: boolean;
  isGosiApplicable: boolean;
  standardValue?: string;
}

export function PayrollView() {
  const { locale } = useLocale();
  const isAr = locale === "ar";

  const [activeTab, setActiveTab] = useState<"runs" | "payslips" | "structures" | "eos">("runs");
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedPayslip, setSelectedPayslip] = useState<PayslipItem | null>(null);

  // Mock initial runs data
  const [runs, setRuns] = useState<PayrollRunItem[]>([
    {
      id: "run-001",
      runNumber: "PR-2026-09",
      periodNameAr: "مسير شهر سبتمبر 2026",
      periodNameEn: "September 2026 Payroll",
      periodMonth: 9,
      periodYear: 2026,
      runDate: "2026-09-27",
      employeeCount: 42,
      totalGross: 612500,
      totalDeductions: 62450,
      totalNet: 550050,
      status: "APPROVED",
      wpsFileGenerated: false,
      journalEntryNumber: "JE-PR-2026-09",
    },
    {
      id: "run-002",
      runNumber: "PR-2026-08",
      periodNameAr: "مسير شهر أغسطس 2026",
      periodNameEn: "August 2026 Payroll",
      periodMonth: 8,
      periodYear: 2026,
      runDate: "2026-08-27",
      employeeCount: 41,
      totalGross: 598000,
      totalDeductions: 61120,
      totalNet: 536880,
      status: "POSTED",
      wpsFileGenerated: true,
      journalEntryNumber: "JE-PR-2026-08",
    },
    {
      id: "run-003",
      runNumber: "PR-2026-07",
      periodNameAr: "مسير شهر يوليو 2026",
      periodNameEn: "July 2026 Payroll",
      periodMonth: 7,
      periodYear: 2026,
      runDate: "2026-07-27",
      employeeCount: 40,
      totalGross: 585000,
      totalDeductions: 59800,
      totalNet: 525200,
      status: "PAID",
      wpsFileGenerated: true,
      journalEntryNumber: "JE-PR-2026-07",
    },
  ]);

  // Mock Payslips
  const payslips: PayslipItem[] = [
    {
      id: "ps-01",
      payslipNumber: "PS-202609-0001",
      employeeId: "EMP-00104",
      employeeNameAr: "فهد الحربي",
      employeeNameEn: "Fahad Al-Harbi",
      positionAr: "كبير مهندسي البرمجيات",
      positionEn: "Senior Software Engineer",
      basicSalary: 16000,
      housingAllowance: 4000,
      transportAllowance: 1500,
      otherAllowances: 1000,
      overtimeAmount: 1250,
      grossSalary: 23750,
      gosiDeduction: 1950, // 9.75% on (16000+4000)
      loanDeduction: 0,
      absenceDeduction: 0,
      totalDeductions: 1950,
      netSalary: 21800,
      issueDate: "2026-09-27",
      iban: "SA4410000001234567890123",
      bankNameAr: "البنك الأهلي السعودي (SNB)",
    },
    {
      id: "ps-02",
      payslipNumber: "PS-202609-0002",
      employeeId: "EMP-00108",
      employeeNameAr: "سارة القحطاني",
      employeeNameEn: "Sara Al-Qahtani",
      positionAr: "مديرة الموارد البشرية",
      positionEn: "HR Manager",
      basicSalary: 18000,
      housingAllowance: 4500,
      transportAllowance: 1500,
      otherAllowances: 1500,
      overtimeAmount: 0,
      grossSalary: 25500,
      gosiDeduction: 2193.75, // 9.75% on (18000+4500)
      loanDeduction: 1000,
      absenceDeduction: 0,
      totalDeductions: 3193.75,
      netSalary: 22306.25,
      issueDate: "2026-09-27",
      iban: "SA8020000009876543210987",
      bankNameAr: "مصرف الراجحي",
    },
    {
      id: "ps-03",
      payslipNumber: "PS-202609-0003",
      employeeId: "EMP-00115",
      employeeNameAr: "طارق منصور",
      employeeNameEn: "Tariq Mansour",
      positionAr: "أخصائي مالي أول",
      positionEn: "Senior Financial Specialist",
      basicSalary: 12000,
      housingAllowance: 3000,
      transportAllowance: 1000,
      otherAllowances: 500,
      overtimeAmount: 850,
      grossSalary: 17350,
      gosiDeduction: 1462.5,
      loanDeduction: 500,
      absenceDeduction: 578.33,
      totalDeductions: 2540.83,
      netSalary: 14809.17,
      issueDate: "2026-09-27",
      iban: "SA1250000004561237890123",
      bankNameAr: "بنك الرياض",
    },
    {
      id: "ps-04",
      payslipNumber: "PS-202609-0004",
      employeeId: "EMP-00122",
      employeeNameAr: "عمر الشمري",
      employeeNameEn: "Omar Al-Shammari",
      positionAr: "مدير سلاسل الإمداد",
      positionEn: "Supply Chain Director",
      basicSalary: 22000,
      housingAllowance: 5500,
      transportAllowance: 2000,
      otherAllowances: 2000,
      overtimeAmount: 0,
      grossSalary: 31500,
      gosiDeduction: 2681.25,
      loanDeduction: 0,
      absenceDeduction: 0,
      totalDeductions: 2681.25,
      netSalary: 28818.75,
      issueDate: "2026-09-27",
      iban: "SA5510000008889991112223",
      bankNameAr: "البنك الأهلي السعودي (SNB)",
    },
  ];

  // Components list
  const components: ComponentItem[] = [
    {
      code: "BASIC",
      nameAr: "الراتب الأساسي",
      nameEn: "Basic Salary",
      type: "EARNING",
      calculationType: "FIXED",
      isTaxable: false,
      isGosiApplicable: true,
      standardValue: "أساس العقد (Qiwa)",
    },
    {
      code: "HOUSING",
      nameAr: "بدل السكن",
      nameEn: "Housing Allowance",
      type: "EARNING",
      calculationType: "PERCENTAGE",
      isTaxable: false,
      isGosiApplicable: true,
      standardValue: "25% من الأساسي",
    },
    {
      code: "TRANSPORT",
      nameAr: "بدل المواصلات والوقود",
      nameEn: "Transport Allowance",
      type: "EARNING",
      calculationType: "FIXED",
      isTaxable: false,
      isGosiApplicable: false,
      standardValue: "1,000 - 2,000 ر.س",
    },
    {
      code: "OVERTIME",
      nameAr: "العمل الإضافي (المادة 107)",
      nameEn: "Statutory Overtime",
      type: "EARNING",
      calculationType: "FORMULA",
      isTaxable: false,
      isGosiApplicable: false,
      standardValue: "1.50x معدل أجر الساعة",
    },
    {
      code: "GOSI_EMPLOYEE",
      nameAr: "استقطاع التأمينات (الموظف)",
      nameEn: "GOSI Social Insurance (Employee)",
      type: "DEDUCTION",
      calculationType: "PERCENTAGE",
      isTaxable: false,
      isGosiApplicable: true,
      standardValue: "9.75% (9% معاشات + 0.75% ساند)",
    },
    {
      code: "LOAN_INSTALLMENT",
      nameAr: "قسط سلفة / قرض شخصي",
      nameEn: "Loan / Advance Installment",
      type: "DEDUCTION",
      calculationType: "FIXED",
      isTaxable: false,
      isGosiApplicable: false,
      standardValue: "حسب جدول السداد المعتمد",
    },
    {
      code: "UNPAID_ABSENCE",
      nameAr: "استقطاع إجازة بدون راتب / غياب",
      nameEn: "Unpaid Leave / Absence Deduction",
      type: "DEDUCTION",
      calculationType: "FORMULA",
      isTaxable: false,
      isGosiApplicable: false,
      standardValue: "(الإجمالي ÷ 30) × الأيام",
    },
  ];

  // End of Service Calculator State
  const [eosGrossSalary, setEosGrossSalary] = useState(15000);
  const [eosYears, setEosYears] = useState(6);
  const [eosMonths, setEosMonths] = useState(4);
  const [eosReason, setEosReason] = useState<"TERMINATION_BY_EMPLOYER" | "RESIGNATION">("TERMINATION_BY_EMPLOYER");

  // Calculate EOS
  const calculateEOS = () => {
    const totalYears = eosYears + eosMonths / 12;
    let baseAward = 0;

    if (totalYears <= 5) {
      baseAward = totalYears * (eosGrossSalary * 0.5);
    } else {
      const firstFive = 5 * (eosGrossSalary * 0.5);
      const remaining = (totalYears - 5) * eosGrossSalary;
      baseAward = firstFive + remaining;
    }

    let multiplier = 1.0;
    let note = isAr
      ? "المادة 84 (إنهاء من صاحب العمل أو عقد محدد): استحقاق كامل المكافأة بنسبة 100%"
      : "Article 84 (Employer Termination): 100% full statutory entitlement";

    if (eosReason === "RESIGNATION") {
      if (totalYears < 2) {
        multiplier = 0.0;
        note = isAr
          ? "المادة 85 (استقالة أقل من سنتين): لا يستحق مكافأة نهاية خدمة (0%)"
          : "Article 85 (Resignation < 2 years): No entitlement (0%)";
      } else if (totalYears < 5) {
        multiplier = 1 / 3;
        note = isAr
          ? "المادة 85 (استقالة من 2 إلى 5 سنوات): يستحق ثلث المكافأة (33.33%)"
          : "Article 85 (Resignation 2-5 years): One-third entitlement (33.33%)";
      } else if (totalYears < 10) {
        multiplier = 2 / 3;
        note = isAr
          ? "المادة 85 (استقالة من 5 إلى 10 سنوات): يستحق ثلثي المكافأة (66.67%)"
          : "Article 85 (Resignation 5-10 years): Two-thirds entitlement (66.67%)";
      } else {
        multiplier = 1.0;
        note = isAr
          ? "المادة 85 (استقالة 10 سنوات فأكثر): يستحق المكافأة كاملة (100%)"
          : "Article 85 (Resignation 10+ years): Full entitlement (100%)";
      }
    }

    const finalAward = baseAward * multiplier;
    return {
      baseAward,
      multiplier,
      finalAward,
      note,
    };
  };

  const eosResult = calculateEOS();

  // Handlers for Payroll Run Actions
  const handleApproveRun = (id: string) => {
    setRuns((prev) =>
      prev.map((r) => (r.id === id ? { ...r, status: "APPROVED" as const } : r))
    );
  };

  const handlePostRun = (id: string) => {
    setRuns((prev) =>
      prev.map((r) =>
        r.id === id
          ? {
              ...r,
              status: "POSTED" as const,
              wpsFileGenerated: true,
              journalEntryNumber: `JE-${r.runNumber}`,
            }
          : r
      )
    );
  };

  const handleExportWps = (run: PayrollRunItem) => {
    const filename = `WPS_${run.runNumber}_${run.runDate.replace(/-/g, "")}.sif`;
    const dummyContent = `SCR,1010789456,NCBK,${run.runDate},${run.employeeCount},${run.totalNet.toFixed(2)},SAR\nEDR,1098765432,NCBK,SA4410000001234567890123,21800.00,16000.00,4000.00,1250.00,1950.00`;
    const blob = new Blob([dummyContent], { type: "text/plain;charset=utf-8" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  };

  const filteredPayslips = payslips.filter((ps) => {
    const term = searchQuery.toLowerCase();
    return (
      ps.employeeNameAr.includes(term) ||
      ps.employeeNameEn.toLowerCase().includes(term) ||
      ps.employeeId.toLowerCase().includes(term) ||
      ps.payslipNumber.toLowerCase().includes(term)
    );
  });

  return (
    <div className="space-y-8 pb-16">
      {/* Header Section */}
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 border-b border-border/60 pb-6">
        <div>
          <div className="flex items-center gap-3 mb-2">
            <div className="p-2 rounded-xl bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20">
              <Banknote className="w-6 h-6" />
            </div>
            <span className="text-xs font-semibold px-2.5 py-1 rounded-full bg-emerald-500/10 text-emerald-700 dark:text-emerald-300 border border-emerald-500/20">
              {isAr ? "محرك الرواتب وحماية الأجور WPS" : "Payroll Engine & Wage Protection (WPS)"}
            </span>
          </div>
          <h1 className="text-2xl font-bold tracking-tight text-foreground">
            {isAr ? "إدارة الرواتب والأجور ومسيرات الصرف" : "Payroll Management & Disbursement Runs"}
          </h1>
          <p className="text-sm text-muted-foreground mt-1 max-w-3xl">
            {isAr
              ? "حسابات دقيقة غير قابلة للتعديل بعد الترحيل، متوافقة 100% مع نظام حماية الأجور السعودي (WPS / MHRSD) وتأمينات GOSI ومكافأة نهاية الخدمة."
              : "Deterministic, immutable payroll calculation engine compliant with Saudi Wage Protection System (WPS / SIF), GOSI, and Labor Law EOS Award."}
          </p>
        </div>

        <div className="flex items-center gap-3">
          <button
            onClick={() => {
              const newRunNumber = `PR-2026-10-${Math.floor(100 + Math.random() * 900)}`;
              setRuns((prev) => [
                {
                  id: `run-${Date.now()}`,
                  runNumber: newRunNumber,
                  periodNameAr: "مسير شهر أكتوبر 2026",
                  periodNameEn: "October 2026 Payroll",
                  periodMonth: 10,
                  periodYear: 2026,
                  runDate: "2026-10-27",
                  employeeCount: 43,
                  totalGross: 625000,
                  totalDeductions: 63800,
                  totalNet: 561200,
                  status: "CALCULATED",
                  wpsFileGenerated: false,
                },
                ...prev,
              ]);
            }}
            className="flex items-center gap-2 px-4 py-2.5 rounded-xl bg-primary text-primary-foreground font-medium text-sm shadow-md shadow-primary/20 hover:opacity-95 transition-all"
          >
            <Plus className="w-4 h-4" />
            <span>{isAr ? "إنشاء مسير رواتب جديد" : "Create Payroll Run"}</span>
          </button>
        </div>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5">
        <div className="p-5 rounded-2xl bg-card border border-border/70 shadow-sm relative overflow-hidden group">
          <div className="absolute top-0 right-0 w-24 h-24 bg-emerald-500/5 rounded-bl-full pointer-events-none transition-transform group-hover:scale-110" />
          <div className="flex items-center justify-between text-muted-foreground mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider">
              {isAr ? "إجمالي كتلة الرواتب (شهرياً)" : "Total Monthly Payroll (Gross)"}
            </span>
            <DollarSign className="w-5 h-5 text-emerald-600 dark:text-emerald-400" />
          </div>
          <div className="text-2xl font-bold text-foreground">
            612,500 <span className="text-xs font-normal text-muted-foreground">SAR</span>
          </div>
          <div className="mt-2 flex items-center gap-1.5 text-xs text-emerald-600 dark:text-emerald-400 font-medium">
            <TrendingUp className="w-3.5 h-3.5" />
            <span>{isAr ? "+2.4% مقارنة بالشهر السابق" : "+2.4% vs last month"}</span>
          </div>
        </div>

        <div className="p-5 rounded-2xl bg-card border border-border/70 shadow-sm relative overflow-hidden group">
          <div className="absolute top-0 right-0 w-24 h-24 bg-blue-500/5 rounded-bl-full pointer-events-none transition-transform group-hover:scale-110" />
          <div className="flex items-center justify-between text-muted-foreground mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider">
              {isAr ? "صافي الرواتب المستحقة للصرف" : "Net Payable Salaries"}
            </span>
            <CreditCard className="w-5 h-5 text-blue-600 dark:text-blue-400" />
          </div>
          <div className="text-2xl font-bold text-foreground">
            550,050 <span className="text-xs font-normal text-muted-foreground">SAR</span>
          </div>
          <div className="mt-2 flex items-center gap-1.5 text-xs text-blue-600 dark:text-blue-400 font-medium">
            <Building className="w-3.5 h-3.5" />
            <span>{isAr ? "42 موظفاً مشمولين بالصرف" : "42 employees included"}</span>
          </div>
        </div>

        <div className="p-5 rounded-2xl bg-card border border-border/70 shadow-sm relative overflow-hidden group">
          <div className="absolute top-0 right-0 w-24 h-24 bg-purple-500/5 rounded-bl-full pointer-events-none transition-transform group-hover:scale-110" />
          <div className="flex items-center justify-between text-muted-foreground mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider">
              {isAr ? "استقطاعات التأمينات (GOSI)" : "GOSI Deductions"}
            </span>
            <ShieldCheck className="w-5 h-5 text-purple-600 dark:text-purple-400" />
          </div>
          <div className="text-2xl font-bold text-foreground">
            59,718 <span className="text-xs font-normal text-muted-foreground">SAR</span>
          </div>
          <div className="mt-2 flex items-center gap-1.5 text-xs text-purple-600 dark:text-purple-400 font-medium">
            <Percent className="w-3.5 h-3.5" />
            <span>{isAr ? "9.75% حصة الموظفين + ساند" : "9.75% Employee share"}</span>
          </div>
        </div>

        <div className="p-5 rounded-2xl bg-card border border-border/70 shadow-sm relative overflow-hidden group">
          <div className="absolute top-0 right-0 w-24 h-24 bg-teal-500/5 rounded-bl-full pointer-events-none transition-transform group-hover:scale-110" />
          <div className="flex items-center justify-between text-muted-foreground mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider">
              {isAr ? "نسبة الامتثال لحماية الأجور" : "WPS Compliance Rate"}
            </span>
            <CheckCircle2 className="w-5 h-5 text-teal-600 dark:text-teal-400" />
          </div>
          <div className="text-2xl font-bold text-foreground">100%</div>
          <div className="mt-2 flex items-center gap-1.5 text-xs text-teal-600 dark:text-teal-400 font-medium">
            <CheckCircle2 className="w-3.5 h-3.5" />
            <span>{isAr ? "متوافق كلياً مع منصة مدد ومنصة قوى" : "Fully compliant (Mudad & Qiwa)"}</span>
          </div>
        </div>
      </div>

      {/* Tabs Navigation */}
      <div className="flex border-b border-border">
        <button
          onClick={() => setActiveTab("runs")}
          className={`flex items-center gap-2 px-5 py-3 font-medium text-sm border-b-2 transition-all ${
            activeTab === "runs"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <Calendar className="w-4 h-4" />
          <span>{isAr ? "مسيرات الرواتب (Payroll Runs)" : "Payroll Runs"}</span>
        </button>

        <button
          onClick={() => setActiveTab("payslips")}
          className={`flex items-center gap-2 px-5 py-3 font-medium text-sm border-b-2 transition-all ${
            activeTab === "payslips"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <FileText className="w-4 h-4" />
          <span>{isAr ? "قسائم الرواتب الفردية (Payslips)" : "Individual Payslips"}</span>
        </button>

        <button
          onClick={() => setActiveTab("structures")}
          className={`flex items-center gap-2 px-5 py-3 font-medium text-sm border-b-2 transition-all ${
            activeTab === "structures"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <Layers className="w-4 h-4" />
          <span>{isAr ? "عناصر وهياكل الرواتب" : "Salary Components & Structures"}</span>
        </button>

        <button
          onClick={() => setActiveTab("eos")}
          className={`flex items-center gap-2 px-5 py-3 font-medium text-sm border-b-2 transition-all ${
            activeTab === "eos"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <Calculator className="w-4 h-4" />
          <span>{isAr ? "حاسبة مكافأة نهاية الخدمة (EOS)" : "End of Service Calculator"}</span>
        </button>
      </div>

      {/* Tab 1: Runs */}
      {activeTab === "runs" && (
        <div className="space-y-4">
          <div className="bg-card rounded-2xl border border-border/70 overflow-hidden shadow-sm">
            <div className="overflow-x-auto">
              <table className="w-full text-sm text-right">
                <thead className="bg-muted/40 text-muted-foreground text-xs font-semibold uppercase border-b border-border">
                  <tr>
                    <th className="px-5 py-4 text-start">{isAr ? "رقم المسير" : "Run #"}</th>
                    <th className="px-5 py-4 text-start">{isAr ? "فترة المسير" : "Period"}</th>
                    <th className="px-5 py-4 text-start">{isAr ? "تاريخ الاستحقاق" : "Due Date"}</th>
                    <th className="px-5 py-4 text-center">{isAr ? "الموظفون" : "Employees"}</th>
                    <th className="px-5 py-4 text-end">{isAr ? "إجمالي المستحقات" : "Total Gross"}</th>
                    <th className="px-5 py-4 text-end">{isAr ? "الاستقطاعات" : "Deductions"}</th>
                    <th className="px-5 py-4 text-end">{isAr ? "صافي الصرف" : "Total Net"}</th>
                    <th className="px-5 py-4 text-center">{isAr ? "الحالة" : "Status"}</th>
                    <th className="px-5 py-4 text-center">{isAr ? "حماية الأجور WPS" : "WPS SIF"}</th>
                    <th className="px-5 py-4 text-center">{isAr ? "الإجراءات" : "Actions"}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border/60">
                  {runs.map((run) => (
                    <tr key={run.id} className="hover:bg-muted/20 transition-colors">
                      <td className="px-5 py-4 font-mono font-medium text-foreground text-start">
                        {run.runNumber}
                      </td>
                      <td className="px-5 py-4 text-start font-medium text-foreground">
                        {isAr ? run.periodNameAr : run.periodNameEn}
                      </td>
                      <td className="px-5 py-4 text-start text-muted-foreground">
                        {run.runDate}
                      </td>
                      <td className="px-5 py-4 text-center">
                        <span className="inline-flex items-center gap-1 font-medium px-2 py-0.5 rounded-full bg-muted text-foreground text-xs">
                          <User className="w-3 h-3" />
                          {run.employeeCount}
                        </span>
                      </td>
                      <td className="px-5 py-4 text-end font-semibold text-foreground">
                        {run.totalGross.toLocaleString()} SAR
                      </td>
                      <td className="px-5 py-4 text-end text-rose-600 dark:text-rose-400 font-medium">
                        -{run.totalDeductions.toLocaleString()} SAR
                      </td>
                      <td className="px-5 py-4 text-end font-bold text-emerald-600 dark:text-emerald-400">
                        {run.totalNet.toLocaleString()} SAR
                      </td>
                      <td className="px-5 py-4 text-center">
                        <span
                          className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold ${
                            run.status === "POSTED"
                              ? "bg-purple-500/10 text-purple-600 dark:text-purple-400 border border-purple-500/20"
                              : run.status === "APPROVED"
                              ? "bg-blue-500/10 text-blue-600 dark:text-blue-400 border border-blue-500/20"
                              : run.status === "PAID"
                              ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20"
                              : "bg-amber-500/10 text-amber-600 dark:text-amber-400 border border-amber-500/20"
                          }`}
                        >
                          {run.status === "POSTED" && <CheckCircle2 className="w-3.5 h-3.5" />}
                          {run.status === "APPROVED" && <FileCheck className="w-3.5 h-3.5" />}
                          {run.status}
                        </span>
                      </td>
                      <td className="px-5 py-4 text-center">
                        {run.wpsFileGenerated ? (
                          <button
                            onClick={() => handleExportWps(run)}
                            className="inline-flex items-center gap-1 text-xs font-medium px-2.5 py-1 rounded-lg bg-emerald-500/10 text-emerald-700 dark:text-emerald-300 hover:bg-emerald-500/20 border border-emerald-500/20 transition-all"
                            title="تحميل ملف SIF المعتمد للبنك"
                          >
                            <Download className="w-3.5 h-3.5" />
                            <span>SIF Ready</span>
                          </button>
                        ) : (
                          <span className="text-xs text-muted-foreground">
                            {isAr ? "قيد التدقيق" : "Pending Post"}
                          </span>
                        )}
                      </td>
                      <td className="px-5 py-4 text-center">
                        <div className="flex items-center justify-center gap-2">
                          {run.status === "CALCULATED" && (
                            <button
                              onClick={() => handleApproveRun(run.id)}
                              className="px-2.5 py-1 rounded-lg bg-blue-600 hover:bg-blue-700 text-white text-xs font-medium transition-all"
                            >
                              {isAr ? "اعتماد" : "Approve"}
                            </button>
                          )}
                          {run.status === "APPROVED" && (
                            <button
                              onClick={() => handlePostRun(run.id)}
                              className="px-2.5 py-1 rounded-lg bg-purple-600 hover:bg-purple-700 text-white text-xs font-medium transition-all shadow-sm"
                            >
                              {isAr ? "ترحيل للدفتر" : "Post to GL"}
                            </button>
                          )}
                          {run.status === "POSTED" && (
                            <span className="text-xs font-mono text-muted-foreground flex items-center gap-1">
                              <CheckCircle2 className="w-3.5 h-3.5 text-emerald-500" />
                              {run.journalEntryNumber}
                            </span>
                          )}
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* Tab 2: Payslips */}
      {activeTab === "payslips" && (
        <div className="space-y-6">
          <div className="flex items-center justify-between gap-4">
            <div className="relative flex-1 max-w-md">
              <Search className="w-4 h-4 absolute top-1/2 -translate-y-1/2 start-3 text-muted-foreground" />
              <input
                type="text"
                placeholder={isAr ? "بحث بالاسم، الرقم الوظيفي، أو رقم القسيمة..." : "Search by name, ID, or payslip #..."}
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full ps-9 pe-4 py-2 rounded-xl bg-card border border-border text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
              />
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
            {filteredPayslips.map((slip) => (
              <div
                key={slip.id}
                onClick={() => setSelectedPayslip(slip)}
                className="p-5 rounded-2xl bg-card border border-border/70 hover:border-primary/50 shadow-sm hover:shadow-md transition-all cursor-pointer relative group flex flex-col justify-between"
              >
                <div>
                  <div className="flex items-center justify-between mb-3">
                    <span className="text-xs font-mono px-2 py-0.5 rounded bg-muted text-foreground">
                      {slip.payslipNumber}
                    </span>
                    <span className="text-xs text-muted-foreground">{slip.issueDate}</span>
                  </div>

                  <h3 className="font-semibold text-foreground text-base">
                    {isAr ? slip.employeeNameAr : slip.employeeNameEn}
                  </h3>
                  <p className="text-xs text-muted-foreground mt-0.5">
                    {isAr ? slip.positionAr : slip.positionEn} • {slip.employeeId}
                  </p>

                  <div className="mt-4 pt-4 border-t border-border/60 space-y-2 text-xs">
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">{isAr ? "الراتب الأساسي:" : "Basic Salary:"}</span>
                      <span className="font-medium text-foreground">{slip.basicSalary.toLocaleString()} SAR</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">{isAr ? "البدلات والإضافي:" : "Allowances & OT:"}</span>
                      <span className="font-medium text-foreground">
                        {(slip.housingAllowance + slip.transportAllowance + slip.otherAllowances + slip.overtimeAmount).toLocaleString()} SAR
                      </span>
                    </div>
                    <div className="flex justify-between text-rose-600 dark:text-rose-400">
                      <span>{isAr ? "إجمالي الاستقطاعات:" : "Deductions:"}</span>
                      <span className="font-medium">-{slip.totalDeductions.toLocaleString()} SAR</span>
                    </div>
                  </div>
                </div>

                <div className="mt-4 pt-4 border-t border-border/60 flex items-center justify-between">
                  <span className="text-xs font-medium text-muted-foreground">{isAr ? "صافي الراتب:" : "Net Salary:"}</span>
                  <div className="text-lg font-bold text-emerald-600 dark:text-emerald-400">
                    {slip.netSalary.toLocaleString()} SAR
                  </div>
                </div>
              </div>
            ))}
          </div>

          {/* Payslip Detailed Modal / Drawer */}
          {selectedPayslip && (
            <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
              <div className="bg-card w-full max-w-2xl rounded-3xl border border-border shadow-2xl p-6 sm:p-8 space-y-6 max-h-[90vh] overflow-y-auto">
                <div className="flex items-center justify-between border-b border-border pb-4">
                  <div className="flex items-center gap-3">
                    <div className="p-2.5 rounded-xl bg-primary/10 text-primary">
                      <FileSpreadsheet className="w-6 h-6" />
                    </div>
                    <div>
                      <h2 className="text-lg font-bold text-foreground">
                        {isAr ? "تفاصيل قسيمة الراتب الرسمية" : "Official Payslip Details"}
                      </h2>
                      <span className="text-xs font-mono text-muted-foreground">
                        {selectedPayslip.payslipNumber}
                      </span>
                    </div>
                  </div>
                  <button
                    onClick={() => setSelectedPayslip(null)}
                    className="p-2 rounded-xl text-muted-foreground hover:bg-muted transition-colors text-sm"
                  >
                    ✕
                  </button>
                </div>

                {/* Employee Info Header */}
                <div className="p-4 rounded-2xl bg-muted/40 border border-border/70 grid grid-cols-2 gap-4 text-xs">
                  <div>
                    <span className="text-muted-foreground">{isAr ? "اسم الموظف:" : "Employee:"}</span>
                    <p className="font-semibold text-sm text-foreground mt-0.5">
                      {isAr ? selectedPayslip.employeeNameAr : selectedPayslip.employeeNameEn}
                    </p>
                  </div>
                  <div>
                    <span className="text-muted-foreground">{isAr ? "الرقم الوظيفي:" : "Employee ID:"}</span>
                    <p className="font-mono text-foreground mt-0.5">{selectedPayslip.employeeId}</p>
                  </div>
                  <div>
                    <span className="text-muted-foreground">{isAr ? "الحساب البنكي (IBAN):" : "IBAN:"}</span>
                    <p className="font-mono text-foreground mt-0.5">{selectedPayslip.iban}</p>
                  </div>
                  <div>
                    <span className="text-muted-foreground">{isAr ? "البنك المعتمد:" : "Bank:"}</span>
                    <p className="font-semibold text-foreground mt-0.5">{selectedPayslip.bankNameAr}</p>
                  </div>
                </div>

                {/* Breakdown Tables */}
                <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
                  {/* Earnings */}
                  <div className="space-y-3">
                    <h4 className="text-xs font-bold uppercase text-emerald-600 dark:text-emerald-400 tracking-wider">
                      {isAr ? "الاستحقاقات (Earnings)" : "Earnings"}
                    </h4>
                    <div className="space-y-2 text-xs bg-card p-3 rounded-xl border border-border/70">
                      <div className="flex justify-between py-1 border-b border-border/40">
                        <span>{isAr ? "الراتب الأساسي" : "Basic Salary"}</span>
                        <span className="font-semibold">{selectedPayslip.basicSalary.toLocaleString()} SAR</span>
                      </div>
                      <div className="flex justify-between py-1 border-b border-border/40">
                        <span>{isAr ? "بدل السكن" : "Housing Allowance"}</span>
                        <span className="font-semibold">{selectedPayslip.housingAllowance.toLocaleString()} SAR</span>
                      </div>
                      <div className="flex justify-between py-1 border-b border-border/40">
                        <span>{isAr ? "بدل المواصلات" : "Transport Allowance"}</span>
                        <span className="font-semibold">{selectedPayslip.transportAllowance.toLocaleString()} SAR</span>
                      </div>
                      <div className="flex justify-between py-1 border-b border-border/40">
                        <span>{isAr ? "بدلات أخرى" : "Other Allowances"}</span>
                        <span className="font-semibold">{selectedPayslip.otherAllowances.toLocaleString()} SAR</span>
                      </div>
                      {selectedPayslip.overtimeAmount > 0 && (
                        <div className="flex justify-between py-1 border-b border-border/40 text-blue-600 dark:text-blue-400">
                          <span>{isAr ? "عمل إضافي (المادة 107)" : "Overtime"}</span>
                          <span className="font-semibold">{selectedPayslip.overtimeAmount.toLocaleString()} SAR</span>
                        </div>
                      )}
                      <div className="flex justify-between pt-2 font-bold text-foreground">
                        <span>{isAr ? "إجمالي المستحقات" : "Total Gross"}</span>
                        <span>{selectedPayslip.grossSalary.toLocaleString()} SAR</span>
                      </div>
                    </div>
                  </div>

                  {/* Deductions */}
                  <div className="space-y-3">
                    <h4 className="text-xs font-bold uppercase text-rose-600 dark:text-rose-400 tracking-wider">
                      {isAr ? "الاستقطاعات (Deductions)" : "Deductions"}
                    </h4>
                    <div className="space-y-2 text-xs bg-card p-3 rounded-xl border border-border/70">
                      <div className="flex justify-between py-1 border-b border-border/40">
                        <span>{isAr ? "التأمينات (9.75% معاشات وساند)" : "GOSI Employee (9.75%)"}</span>
                        <span className="font-semibold text-rose-600">
                          -{selectedPayslip.gosiDeduction.toLocaleString()} SAR
                        </span>
                      </div>
                      {selectedPayslip.loanDeduction > 0 && (
                        <div className="flex justify-between py-1 border-b border-border/40">
                          <span>{isAr ? "قسط سلفة موظف" : "Loan Repayment"}</span>
                          <span className="font-semibold text-rose-600">
                            -{selectedPayslip.loanDeduction.toLocaleString()} SAR
                          </span>
                        </div>
                      )}
                      {selectedPayslip.absenceDeduction > 0 && (
                        <div className="flex justify-between py-1 border-b border-border/40">
                          <span>{isAr ? "خصم غياب / بدون راتب" : "Unpaid Leave Deduction"}</span>
                          <span className="font-semibold text-rose-600">
                            -{selectedPayslip.absenceDeduction.toLocaleString()} SAR
                          </span>
                        </div>
                      )}
                      <div className="flex justify-between pt-2 font-bold text-rose-600">
                        <span>{isAr ? "إجمالي الاستقطاعات" : "Total Deductions"}</span>
                        <span>-{selectedPayslip.totalDeductions.toLocaleString()} SAR</span>
                      </div>
                    </div>
                  </div>
                </div>

                {/* Final Net Highlight */}
                <div className="p-4 rounded-2xl bg-emerald-500/10 border border-emerald-500/30 flex items-center justify-between">
                  <div>
                    <span className="text-xs font-medium text-emerald-800 dark:text-emerald-300">
                      {isAr ? "صافي الراتب المستحق للصرف في الحساب البنكي:" : "Net Payable Salary to Bank:"}
                    </span>
                    <p className="text-xs text-muted-foreground mt-0.5">
                      {isAr ? "تم إدراج هذا المبلغ في ملف حماية الأجور (WPS SIF)" : "Included in certified WPS SIF file"}
                    </p>
                  </div>
                  <div className="text-2xl font-black text-emerald-600 dark:text-emerald-400">
                    {selectedPayslip.netSalary.toLocaleString()} SAR
                  </div>
                </div>

                <div className="flex justify-end gap-3 pt-2">
                  <button
                    onClick={() => window.print()}
                    className="flex items-center gap-2 px-4 py-2 rounded-xl border border-border text-xs font-medium hover:bg-muted transition-all"
                  >
                    <Printer className="w-4 h-4" />
                    <span>{isAr ? "طباعة القسيمة الرسمية" : "Print Official Payslip"}</span>
                  </button>
                  <button
                    onClick={() => setSelectedPayslip(null)}
                    className="px-5 py-2 rounded-xl bg-primary text-primary-foreground text-xs font-medium transition-all"
                  >
                    {isAr ? "إغلاق" : "Close"}
                  </button>
                </div>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Tab 3: Salary Components & Structures */}
      {activeTab === "structures" && (
        <div className="space-y-6">
          <div className="bg-card rounded-2xl border border-border/70 overflow-hidden shadow-sm">
            <div className="p-5 border-b border-border/60">
              <h3 className="font-bold text-foreground text-base">
                {isAr ? "عناصر الأجور المعتمدة (Salary Components Catalog)" : "Standard Salary Components"}
              </h3>
              <p className="text-xs text-muted-foreground mt-1">
                {isAr
                  ? "تحديد الاستحقاقات والاستقطاعات وتوافقها مع احتساب التأمينات الاجتماعية (GOSI) والضرائب."
                  : "Earnings and deductions definitions configured according to Saudi Labor Law and GOSI rules."}
              </p>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full text-sm text-right">
                <thead className="bg-muted/40 text-muted-foreground text-xs font-semibold uppercase border-b border-border">
                  <tr>
                    <th className="px-5 py-3 text-start">{isAr ? "الرمز" : "Code"}</th>
                    <th className="px-5 py-3 text-start">{isAr ? "اسم العنصر" : "Component Name"}</th>
                    <th className="px-5 py-3 text-center">{isAr ? "النوع" : "Type"}</th>
                    <th className="px-5 py-3 text-center">{isAr ? "آلية الاحتساب" : "Calculation"}</th>
                    <th className="px-5 py-3 text-center">{isAr ? "خاضع للتأمينات (GOSI)" : "GOSI Subject"}</th>
                    <th className="px-5 py-3 text-start">{isAr ? "القيمة الافتراضية / المعادلة" : "Default / Formula"}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border/60">
                  {components.map((comp) => (
                    <tr key={comp.code} className="hover:bg-muted/20 transition-colors">
                      <td className="px-5 py-4 font-mono font-medium text-foreground text-start">
                        {comp.code}
                      </td>
                      <td className="px-5 py-4 text-start font-medium text-foreground">
                        {isAr ? comp.nameAr : comp.nameEn}
                      </td>
                      <td className="px-5 py-4 text-center">
                        <span
                          className={`text-xs px-2.5 py-1 rounded-full font-semibold ${
                            comp.type === "EARNING"
                              ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400"
                              : "bg-rose-500/10 text-rose-600 dark:text-rose-400"
                          }`}
                        >
                          {comp.type}
                        </span>
                      </td>
                      <td className="px-5 py-4 text-center font-mono text-xs text-muted-foreground">
                        {comp.calculationType}
                      </td>
                      <td className="px-5 py-4 text-center">
                        {comp.isGosiApplicable ? (
                          <span className="inline-flex items-center text-xs font-medium text-emerald-600 gap-1">
                            <CheckCircle2 className="w-3.5 h-3.5" />
                            {isAr ? "نعم" : "Yes"}
                          </span>
                        ) : (
                          <span className="text-xs text-muted-foreground">{isAr ? "لا" : "No"}</span>
                        )}
                      </td>
                      <td className="px-5 py-4 text-start text-xs font-medium text-muted-foreground">
                        {comp.standardValue}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* Tab 4: End of Service Calculator */}
      {activeTab === "eos" && (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="lg:col-span-2 space-y-6">
            <div className="bg-card rounded-2xl border border-border/70 p-6 shadow-sm space-y-5">
              <div>
                <h3 className="text-base font-bold text-foreground">
                  {isAr ? "حاسبة مكافأة نهاية الخدمة (نظام العمل السعودي)" : "Saudi Labor Law End of Service Award"}
                </h3>
                <p className="text-xs text-muted-foreground mt-1">
                  {isAr
                    ? "احتساب مكافأة نهاية الخدمة آلياً وتطبيق المواد 84 و 85 بدقة حسابية تامة."
                    : "Automated EOS calculation compliant with Articles 84 & 85 of Saudi Labor Law."}
                </p>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div className="space-y-1.5">
                  <label className="text-xs font-medium text-muted-foreground">
                    {isAr ? "آخر راتب شهري إجمالي (SAR):" : "Last Gross Salary (SAR):"}
                  </label>
                  <input
                    type="number"
                    value={eosGrossSalary}
                    onChange={(e) => setEosGrossSalary(Number(e.target.value) || 0)}
                    className="w-full px-3.5 py-2 rounded-xl bg-muted/40 border border-border text-sm font-semibold focus:outline-none focus:ring-2 focus:ring-primary/20"
                  />
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs font-medium text-muted-foreground">
                    {isAr ? "سبب انتهاء العلاقة العمالية:" : "Termination Reason:"}
                  </label>
                  <select
                    value={eosReason}
                    onChange={(e) => setEosReason(e.target.value as any)}
                    className="w-full px-3.5 py-2 rounded-xl bg-muted/40 border border-border text-sm font-medium focus:outline-none focus:ring-2 focus:ring-primary/20"
                  >
                    <option value="TERMINATION_BY_EMPLOYER">
                      {isAr ? "المادة 84 - إنهاء من صاحب العمل / انتهاء العقد" : "Article 84 - Employer Termination / End of Contract"}
                    </option>
                    <option value="RESIGNATION">
                      {isAr ? "المادة 85 - استقالة الموظف بمحض إرادته" : "Article 85 - Employee Resignation"}
                    </option>
                  </select>
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs font-medium text-muted-foreground">
                    {isAr ? "سنوات الخدمة:" : "Service Years:"}
                  </label>
                  <input
                    type="number"
                    min="0"
                    max="50"
                    value={eosYears}
                    onChange={(e) => setEosYears(Number(e.target.value) || 0)}
                    className="w-full px-3.5 py-2 rounded-xl bg-muted/40 border border-border text-sm font-semibold focus:outline-none focus:ring-2 focus:ring-primary/20"
                  />
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs font-medium text-muted-foreground">
                    {isAr ? "أشهر الخدمة الإضافية:" : "Additional Months:"}
                  </label>
                  <input
                    type="number"
                    min="0"
                    max="11"
                    value={eosMonths}
                    onChange={(e) => setEosMonths(Number(e.target.value) || 0)}
                    className="w-full px-3.5 py-2 rounded-xl bg-muted/40 border border-border text-sm font-semibold focus:outline-none focus:ring-2 focus:ring-primary/20"
                  />
                </div>
              </div>

              {/* Legal Notice */}
              <div className="p-4 rounded-xl bg-blue-500/10 border border-blue-500/20 text-xs text-blue-800 dark:text-blue-300 space-y-1">
                <div className="font-semibold flex items-center gap-1.5">
                  <AlertCircle className="w-4 h-4" />
                  <span>{isAr ? "القاعدة القانونية المطبقة:" : "Applied Statutory Rule:"}</span>
                </div>
                <p>{eosResult.note}</p>
              </div>
            </div>
          </div>

          {/* Result Card */}
          <div className="space-y-6">
            <div className="bg-card rounded-2xl border border-border/70 p-6 shadow-sm space-y-5">
              <h4 className="text-xs font-bold uppercase tracking-wider text-muted-foreground">
                {isAr ? "ملخص الاستحقاق النهائي" : "Statutory Settlement Summary"}
              </h4>

              <div className="space-y-3 text-xs">
                <div className="flex justify-between text-muted-foreground">
                  <span>{isAr ? "مدة الخدمة الإجمالية:" : "Total Service Duration:"}</span>
                  <span className="font-medium text-foreground">
                    {eosYears} {isAr ? "سنوات" : "years"} و {eosMonths} {isAr ? "أشهر" : "months"}
                  </span>
                </div>
                <div className="flex justify-between text-muted-foreground">
                  <span>{isAr ? "الاستحقاق الأولي (قبل نسبة المادة 85):" : "Base Award Amount:"}</span>
                  <span className="font-medium text-foreground">
                    {eosResult.baseAward.toLocaleString(undefined, { maximumFractionDigits: 2 })} SAR
                  </span>
                </div>
                <div className="flex justify-between text-muted-foreground">
                  <span>{isAr ? "معامل الاستحقاق القانوني:" : "Entitlement Ratio:"}</span>
                  <span className="font-bold text-foreground">
                    {(eosResult.multiplier * 100).toFixed(1)}%
                  </span>
                </div>
              </div>

              <div className="pt-4 border-t border-border">
                <span className="text-xs text-muted-foreground font-medium">
                  {isAr ? "صافي مكافأة نهاية الخدمة المستحقة:" : "Final End of Service Award:"}
                </span>
                <div className="text-2xl font-black text-emerald-600 dark:text-emerald-400 mt-1">
                  {eosResult.finalAward.toLocaleString(undefined, { maximumFractionDigits: 2 })} SAR
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
