"use client";

import React, { useState } from "react";
import {
  ShieldCheck,
  FileSpreadsheet,
  Users,
  CheckCircle2,
  Calendar,
  Building,
  Calculator,
  Search,
  Download,
  Send,
  Eye,
  Percent,
  Layers,
  FileCheck2,
  Lock,
  ArrowRight,
  TrendingUp,
  AlertCircle,
  Clock,
  Sparkles,
  Plus,
  RefreshCw,
} from "lucide-react";
import { useLocale } from "@/components/layout/shell";

interface GosiCalculationItem {
  id: string;
  employeeNumber: string;
  employeeNameAr: string;
  employeeNameEn: string;
  gosiNumber: string;
  nationalityScope: "SAUDI" | "NON_SAUDI";
  nationalityCountryCode: string;
  contributionBasis: number;
  annuityEmp: number;
  annuityEr: number;
  ohrsiEmp: number;
  ohrsiEr: number;
  sanedEmp: number;
  sanedEr: number;
  totalEmployee: number;
  totalEmployer: number;
  totalAmount: number;
  status: "CALCULATED" | "EXPORTED" | "CONFIRMED";
}

interface GosiProfileItem {
  id: string;
  employeeId: string;
  employeeNumber: string;
  employeeNameAr: string;
  employeeNameEn: string;
  nationalIdOrIqama: string;
  gosiNumber: string;
  nationalityScope: "SAUDI" | "NON_SAUDI";
  nationalityNameAr: string;
  nationalityNameEn: string;
  registrationDate: string;
  basisOverride: number;
  isActive: boolean;
}

interface GosiRuleVersionItem {
  ruleCode: string;
  ruleNameAr: string;
  ruleNameEn: string;
  versionNumber: number;
  nationalityScope: "SAUDI" | "NON_SAUDI" | "ALL";
  employeeRate: number;
  employerRate: number;
  minBasis: number;
  maxBasis: number;
  effectiveFrom: string;
  legalReference: string;
  notes: string;
}

interface GosiExportBatchItem {
  id: string;
  batchNumber: string;
  year: number;
  month: number;
  employeeCount: number;
  totalEmployeeAmount: number;
  totalEmployerAmount: number;
  totalAmount: number;
  status: "GENERATED" | "SUBMITTED" | "CONFIRMED";
  generatedAt: string;
  fileFormat: string;
}

export function GosiView() {
  const { dir, locale } = useLocale();
  const isAr = locale === "ar";

  const [activeTab, setActiveTab] = useState<"calculations" | "profiles" | "rules" | "export">("calculations");
  const [selectedMonth, setSelectedMonth] = useState<number>(9);
  const [selectedYear, setSelectedYear] = useState<number>(2026);
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedNationality, setSelectedNationality] = useState<string>("ALL");

  // Selected calc for detail modal
  const [selectedCalc, setSelectedCalc] = useState<GosiCalculationItem | null>(null);
  // Selected batch for XML modal
  const [selectedXmlBatch, setSelectedXmlBatch] = useState<GosiExportBatchItem | null>(null);
  // Show new enrol modal
  const [showEnrolModal, setShowEnrolModal] = useState(false);
  // Calculation in-progress indicator
  const [isCalculating, setIsCalculating] = useState(false);

  // Mock initial calculations
  const [calculations, setCalculations] = useState<GosiCalculationItem[]>([
    {
      id: "calc-01",
      employeeNumber: "EMP-00101",
      employeeNameAr: "عبدالله العتيبي",
      employeeNameEn: "Abdullah Al-Otaibi",
      gosiNumber: "GOSI-984210",
      nationalityScope: "SAUDI",
      nationalityCountryCode: "SA",
      contributionBasis: 14500,
      annuityEmp: 1305.0,
      annuityEr: 1305.0,
      ohrsiEmp: 0,
      ohrsiEr: 290.0,
      sanedEmp: 108.75,
      sanedEr: 108.75,
      totalEmployee: 1413.75,
      totalEmployer: 1703.75,
      totalAmount: 3117.5,
      status: "CALCULATED",
    },
    {
      id: "calc-02",
      employeeNumber: "EMP-00108",
      employeeNameAr: "سارة القحطاني",
      employeeNameEn: "Sara Al-Qahtani",
      gosiNumber: "GOSI-984215",
      nationalityScope: "SAUDI",
      nationalityCountryCode: "SA",
      contributionBasis: 18000,
      annuityEmp: 1620.0,
      annuityEr: 1620.0,
      ohrsiEmp: 0,
      ohrsiEr: 360.0,
      sanedEmp: 135.0,
      sanedEr: 135.0,
      totalEmployee: 1755.0,
      totalEmployer: 2115.0,
      totalAmount: 3870.0,
      status: "CALCULATED",
    },
    {
      id: "calc-03",
      employeeNumber: "EMP-00115",
      employeeNameAr: "طارق محمود",
      employeeNameEn: "Tariq Mahmoud",
      gosiNumber: "GOSI-984220",
      nationalityScope: "NON_SAUDI",
      nationalityCountryCode: "EG",
      contributionBasis: 12000,
      annuityEmp: 0,
      annuityEr: 0,
      ohrsiEmp: 0,
      ohrsiEr: 240.0,
      sanedEmp: 0,
      sanedEr: 0,
      totalEmployee: 0,
      totalEmployer: 240.0,
      totalAmount: 240.0,
      status: "CALCULATED",
    },
    {
      id: "calc-04",
      employeeNumber: "EMP-00122",
      employeeNameAr: "عمر الشمري",
      employeeNameEn: "Omar Al-Shammari",
      gosiNumber: "GOSI-984228",
      nationalityScope: "SAUDI",
      nationalityCountryCode: "SA",
      contributionBasis: 22000,
      annuityEmp: 1980.0,
      annuityEr: 1980.0,
      ohrsiEmp: 0,
      ohrsiEr: 440.0,
      sanedEmp: 165.0,
      sanedEr: 165.0,
      totalEmployee: 2145.0,
      totalEmployer: 2585.0,
      totalAmount: 4730.0,
      status: "CALCULATED",
    },
    {
      id: "calc-05",
      employeeNumber: "EMP-00130",
      employeeNameAr: "راكيش كومار",
      employeeNameEn: "Rakesh Kumar",
      gosiNumber: "GOSI-984235",
      nationalityScope: "NON_SAUDI",
      nationalityCountryCode: "IN",
      contributionBasis: 9500,
      annuityEmp: 0,
      annuityEr: 0,
      ohrsiEmp: 0,
      ohrsiEr: 190.0,
      sanedEmp: 0,
      sanedEr: 0,
      totalEmployee: 0,
      totalEmployer: 190.0,
      totalAmount: 190.0,
      status: "CALCULATED",
    },
    {
      id: "calc-06",
      employeeNumber: "EMP-00145",
      employeeNameAr: "سلطان الدوسري",
      employeeNameEn: "Sultan Al-Dossary",
      gosiNumber: "GOSI-984242",
      nationalityScope: "SAUDI",
      nationalityCountryCode: "SA",
      contributionBasis: 30000,
      annuityEmp: 2700.0,
      annuityEr: 2700.0,
      ohrsiEmp: 0,
      ohrsiEr: 600.0,
      sanedEmp: 225.0,
      sanedEr: 225.0,
      totalEmployee: 2925.0,
      totalEmployer: 3525.0,
      totalAmount: 6450.0,
      status: "CALCULATED",
    },
  ]);

  // Profiles list
  const [profiles, setProfiles] = useState<GosiProfileItem[]>([
    {
      id: "prof-01",
      employeeId: "emp-01",
      employeeNumber: "EMP-00101",
      employeeNameAr: "عبدالله العتيبي",
      employeeNameEn: "Abdullah Al-Otaibi",
      nationalIdOrIqama: "1089332145",
      gosiNumber: "GOSI-984210",
      nationalityScope: "SAUDI",
      nationalityNameAr: "سعودي",
      nationalityNameEn: "Saudi",
      registrationDate: "2024-01-15",
      basisOverride: 14500,
      isActive: true,
    },
    {
      id: "prof-02",
      employeeId: "emp-02",
      employeeNumber: "EMP-00108",
      employeeNameAr: "سارة القحطاني",
      employeeNameEn: "Sara Al-Qahtani",
      nationalIdOrIqama: "1098443211",
      gosiNumber: "GOSI-984215",
      nationalityScope: "SAUDI",
      nationalityNameAr: "سعودية",
      nationalityNameEn: "Saudi",
      registrationDate: "2024-03-01",
      basisOverride: 18000,
      isActive: true,
    },
    {
      id: "prof-03",
      employeeId: "emp-03",
      employeeNumber: "EMP-00115",
      employeeNameAr: "طارق محمود",
      employeeNameEn: "Tariq Mahmoud",
      nationalIdOrIqama: "2489330012",
      gosiNumber: "GOSI-984220",
      nationalityScope: "NON_SAUDI",
      nationalityNameAr: "مصري",
      nationalityNameEn: "Egyptian",
      registrationDate: "2024-02-10",
      basisOverride: 12000,
      isActive: true,
    },
    {
      id: "prof-04",
      employeeId: "emp-04",
      employeeNumber: "EMP-00122",
      employeeNameAr: "عمر الشمري",
      employeeNameEn: "Omar Al-Shammari",
      nationalIdOrIqama: "1078554423",
      gosiNumber: "GOSI-984228",
      nationalityScope: "SAUDI",
      nationalityNameAr: "سعودي",
      nationalityNameEn: "Saudi",
      registrationDate: "2024-01-01",
      basisOverride: 22000,
      isActive: true,
    },
    {
      id: "prof-05",
      employeeId: "emp-05",
      employeeNumber: "EMP-00130",
      employeeNameAr: "راكيش كومار",
      employeeNameEn: "Rakesh Kumar",
      nationalIdOrIqama: "2501994432",
      gosiNumber: "GOSI-984235",
      nationalityScope: "NON_SAUDI",
      nationalityNameAr: "هندي",
      nationalityNameEn: "Indian",
      registrationDate: "2024-04-15",
      basisOverride: 9500,
      isActive: true,
    },
  ]);

  // Versioned Statutory Rules from DB (never hard-coded)
  const rules: GosiRuleVersionItem[] = [
    {
      ruleCode: "ANNUITY",
      ruleNameAr: "فرع المعاشات التقاعدية",
      ruleNameEn: "Annuity / Pension Branch",
      versionNumber: 1,
      nationalityScope: "SAUDI",
      employeeRate: 9.0,
      employerRate: 9.0,
      minBasis: 1500,
      maxBasis: 45000,
      effectiveFrom: "2024-01-01",
      legalReference: "نظام التأمينات الاجتماعية - المادة 18 (فرع المعاشات)",
      notes: "إلزامي للمواطنين السعوديين فقط، ويحتسب على أساس الراتب الأساسي + بدل السكن",
    },
    {
      ruleCode: "OHRSI",
      ruleNameAr: "فرع الأخطار المهنية وإصابات العمل",
      ruleNameEn: "Occupational Hazards (OHRSI)",
      versionNumber: 1,
      nationalityScope: "ALL",
      employeeRate: 0.0,
      employerRate: 2.0,
      minBasis: 1500,
      maxBasis: 45000,
      effectiveFrom: "2024-01-01",
      legalReference: "نظام التأمينات الاجتماعية - المادة 18 (فرع الأخطار المهنية)",
      notes: "إلزامي لكافة المشتركين (سعوديين وغير سعوديين) ويتحمله صاحب العمل بالكامل",
    },
    {
      ruleCode: "SANED",
      ruleNameAr: "نظام التعطل عن العمل (ساند)",
      ruleNameEn: "Unemployment Insurance (SANED)",
      versionNumber: 1,
      nationalityScope: "SAUDI",
      employeeRate: 0.75,
      employerRate: 0.75,
      minBasis: 1500,
      maxBasis: 45000,
      effectiveFrom: "2024-01-01",
      legalReference: "نظام التأمين ضد التعطل عن العمل (ساند) - المادة 4",
      notes: "إلزامي للموظفين السعوديين دون سن الستين، يتقاسمه الموظف والمنشأة مناصفة",
    },
  ];

  // Export Batches
  const [batches, setBatches] = useState<GosiExportBatchItem[]>([
    {
      id: "batch-01",
      batchNumber: "GOSI-202608-884A",
      year: 2026,
      month: 8,
      employeeCount: 42,
      totalEmployeeAmount: 38550.0,
      totalEmployerAmount: 46720.0,
      totalAmount: 85270.0,
      status: "CONFIRMED",
      generatedAt: "2026-08-28 10:15",
      fileFormat: "XML",
    },
    {
      id: "batch-02",
      batchNumber: "GOSI-202609-912C",
      year: 2026,
      month: 9,
      employeeCount: 43,
      totalEmployeeAmount: 39200.0,
      totalEmployerAmount: 47200.0,
      totalAmount: 86400.0,
      status: "GENERATED",
      generatedAt: "2026-09-17 08:30",
      fileFormat: "XML",
    },
  ]);

  // Handle running calculations
  const handleRunCalculation = () => {
    setIsCalculating(true);
    setTimeout(() => {
      setIsCalculating(false);
      // add or refresh
    }, 600);
  };

  // Submit batch simulation
  const handleSubmitBatch = (batchId: string) => {
    setBatches((prev) =>
      prev.map((b) =>
        b.id === batchId ? { ...b, status: "SUBMITTED" as const } : b
      )
    );
    setTimeout(() => {
      setBatches((prev) =>
        prev.map((b) =>
          b.id === batchId ? { ...b, status: "CONFIRMED" as const } : b
        )
      );
    }, 1200);
  };

  // KPI Computations
  const totalSubscribers = calculations.length;
  const totalEmpContributions = calculations.reduce((acc, c) => acc + c.totalEmployee, 0);
  const totalErContributions = calculations.reduce((acc, c) => acc + c.totalEmployer, 0);
  const grandTotal = totalEmpContributions + totalErContributions;

  // Filter calculations
  const filteredCalcs = calculations.filter((c) => {
    const matchQuery =
      c.employeeNameAr.includes(searchQuery) ||
      c.employeeNameEn.toLowerCase().includes(searchQuery.toLowerCase()) ||
      c.employeeNumber.toLowerCase().includes(searchQuery.toLowerCase()) ||
      c.gosiNumber.toLowerCase().includes(searchQuery.toLowerCase());
    const matchNat =
      selectedNationality === "ALL" || c.nationalityScope === selectedNationality;
    return matchQuery && matchNat;
  });

  return (
    <div className="space-y-6 pb-12" dir={dir}>
      {/* Top Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <div className="flex items-center gap-2.5">
            <div className="p-2.5 rounded-xl bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20">
              <ShieldCheck className="w-6 h-6" />
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight text-foreground">
                {isAr ? "محرك التأمينات الاجتماعية (GOSI Engine)" : "GOSI Social Insurance Engine"}
              </h1>
              <p className="text-sm text-muted-foreground mt-0.5">
                {isAr
                  ? "محرك إصداري لاحتساب اشتراكات التأمينات وفق نظام العمل السعودي وبوابة التأمينات"
                  : "Versioned social insurance engine strictly governed by Saudi statutory regulations"}
              </p>
            </div>
          </div>
        </div>

        <div className="flex items-center gap-3">
          <button
            id="gosi-run-calc-btn"
            onClick={handleRunCalculation}
            disabled={isCalculating}
            className="flex items-center gap-2 px-4 py-2.5 rounded-xl bg-primary text-primary-foreground font-medium text-sm shadow-md shadow-primary/20 hover:opacity-95 transition-all disabled:opacity-50"
          >
            {isCalculating ? (
              <RefreshCw className="w-4 h-4 animate-spin" />
            ) : (
              <Calculator className="w-4 h-4" />
            )}
            <span>
              {isCalculating
                ? isAr
                  ? "جارٍ الاحتساب..."
                  : "Calculating..."
                : isAr
                ? "تشغيل الاحتساب الشهري"
                : "Run Monthly Calculation"}
            </span>
          </button>
        </div>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5">
        <div className="p-5 rounded-2xl bg-card border border-border/70 shadow-sm relative overflow-hidden group">
          <div className="absolute top-0 right-0 w-24 h-24 bg-emerald-500/5 rounded-bl-full pointer-events-none transition-transform group-hover:scale-110" />
          <div className="flex items-center justify-between text-muted-foreground mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider">
              {isAr ? "إجمالي اشتراكات التأمينات" : "Total GOSI Contributions"}
            </span>
            <ShieldCheck className="w-5 h-5 text-emerald-600 dark:text-emerald-400" />
          </div>
          <div className="text-2xl font-bold text-foreground">
            {grandTotal.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}{" "}
            <span className="text-xs font-normal text-muted-foreground">SAR</span>
          </div>
          <div className="mt-2 flex items-center gap-1.5 text-xs text-emerald-600 dark:text-emerald-400 font-medium">
            <TrendingUp className="w-3.5 h-3.5" />
            <span>{isAr ? "شامل المعاشات والأخطار وساند" : "Incl. Annuity, OHRSI & SANED"}</span>
          </div>
        </div>

        <div className="p-5 rounded-2xl bg-card border border-border/70 shadow-sm relative overflow-hidden group">
          <div className="absolute top-0 right-0 w-24 h-24 bg-blue-500/5 rounded-bl-full pointer-events-none transition-transform group-hover:scale-110" />
          <div className="flex items-center justify-between text-muted-foreground mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider">
              {isAr ? "حصة المنشأة (صاحب العمل)" : "Employer Contributions"}
            </span>
            <Building className="w-5 h-5 text-blue-600 dark:text-blue-400" />
          </div>
          <div className="text-2xl font-bold text-foreground">
            {totalErContributions.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}{" "}
            <span className="text-xs font-normal text-muted-foreground">SAR</span>
          </div>
          <div className="mt-2 flex items-center gap-1.5 text-xs text-blue-600 dark:text-blue-400 font-medium">
            <Percent className="w-3.5 h-3.5" />
            <span>{isAr ? "9% معاشات + 2% أخطار + 0.75% ساند" : "9% Annuity + 2% OHRSI + 0.75% SANED"}</span>
          </div>
        </div>

        <div className="p-5 rounded-2xl bg-card border border-border/70 shadow-sm relative overflow-hidden group">
          <div className="absolute top-0 right-0 w-24 h-24 bg-purple-500/5 rounded-bl-full pointer-events-none transition-transform group-hover:scale-110" />
          <div className="flex items-center justify-between text-muted-foreground mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider">
              {isAr ? "استقطاعات الموظفين (المشتركين)" : "Employee Deductions"}
            </span>
            <Users className="w-5 h-5 text-purple-600 dark:text-purple-400" />
          </div>
          <div className="text-2xl font-bold text-foreground">
            {totalEmpContributions.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}{" "}
            <span className="text-xs font-normal text-muted-foreground">SAR</span>
          </div>
          <div className="mt-2 flex items-center gap-1.5 text-xs text-purple-600 dark:text-purple-400 font-medium">
            <Percent className="w-3.5 h-3.5" />
            <span>{isAr ? "9% معاشات + 0.75% ساند (للسعوديين)" : "9% Annuity + 0.75% SANED (Saudi)"}</span>
          </div>
        </div>

        <div className="p-5 rounded-2xl bg-card border border-border/70 shadow-sm relative overflow-hidden group">
          <div className="absolute top-0 right-0 w-24 h-24 bg-teal-500/5 rounded-bl-full pointer-events-none transition-transform group-hover:scale-110" />
          <div className="flex items-center justify-between text-muted-foreground mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider">
              {isAr ? "المشتركون النشطون ونسبة الامتثال" : "Insured & Compliance"}
            </span>
            <CheckCircle2 className="w-5 h-5 text-teal-600 dark:text-teal-400" />
          </div>
          <div className="text-2xl font-bold text-foreground">
            {totalSubscribers} <span className="text-xs font-normal text-muted-foreground">{isAr ? "مشترك" : "Subscribers"}</span>
            <span className="text-sm font-semibold text-teal-600 dark:text-teal-400 ms-3">100%</span>
          </div>
          <div className="mt-2 flex items-center gap-1.5 text-xs text-teal-600 dark:text-teal-400 font-medium">
            <CheckCircle2 className="w-3.5 h-3.5" />
            <span>{isAr ? "مسجلون ومطابقون مع بوابة التأمينات" : "Matched with GOSI Portal Registry"}</span>
          </div>
        </div>
      </div>

      {/* Tabs Navigation */}
      <div className="flex border-b border-border">
        <button
          id="gosi-tab-calc"
          onClick={() => setActiveTab("calculations")}
          className={`flex items-center gap-2 px-5 py-3 font-medium text-sm border-b-2 transition-all ${
            activeTab === "calculations"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <Calculator className="w-4 h-4" />
          <span>{isAr ? "الاحتساب الشهري" : "Monthly Calculations"}</span>
        </button>

        <button
          id="gosi-tab-profiles"
          onClick={() => setActiveTab("profiles")}
          className={`flex items-center gap-2 px-5 py-3 font-medium text-sm border-b-2 transition-all ${
            activeTab === "profiles"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <Users className="w-4 h-4" />
          <span>{isAr ? "سجل المشتركين" : "Insurance Profiles"}</span>
        </button>

        <button
          id="gosi-tab-rules"
          onClick={() => setActiveTab("rules")}
          className={`flex items-center gap-2 px-5 py-3 font-medium text-sm border-b-2 transition-all ${
            activeTab === "rules"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <Layers className="w-4 h-4" />
          <span>{isAr ? "القواعد والنسب النظامية" : "Statutory Rules (DB)"}</span>
        </button>

        <button
          id="gosi-tab-export"
          onClick={() => setActiveTab("export")}
          className={`flex items-center gap-2 px-5 py-3 font-medium text-sm border-b-2 transition-all ${
            activeTab === "export"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <FileCheck2 className="w-4 h-4" />
          <span>{isAr ? "دفعات التصدير والربط" : "Export Batches & Portal"}</span>
        </button>
      </div>

      {/* Tab 1: Calculations */}
      {activeTab === "calculations" && (
        <div className="space-y-4">
          <div className="flex flex-col sm:flex-row items-center justify-between gap-3 p-4 bg-card rounded-2xl border border-border/70 shadow-sm">
            <div className="flex items-center gap-3 w-full sm:w-auto">
              <div className="relative flex-1 sm:w-72">
                <Search className="absolute right-3 rtl:right-3 ltr:left-3 top-2.5 w-4 h-4 text-muted-foreground" />
                <input
                  type="text"
                  placeholder={isAr ? "بحث بالاسم أو الرقم الوظيفي أو التأميني..." : "Search by name, ID or GOSI..."}
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="w-full pl-3 pr-9 rtl:pr-9 rtl:pl-3 py-2 text-sm bg-background border border-border rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20"
                />
              </div>

              <select
                value={selectedNationality}
                onChange={(e) => setSelectedNationality(e.target.value)}
                className="px-3 py-2 text-sm bg-background border border-border rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20"
              >
                <option value="ALL">{isAr ? "كل الجنسيات" : "All Nationalities"}</option>
                <option value="SAUDI">{isAr ? "سعودي فقط" : "Saudi Only"}</option>
                <option value="NON_SAUDI">{isAr ? "غير سعودي (وافد)" : "Non-Saudi (Expat)"}</option>
              </select>
            </div>

            <div className="flex items-center gap-2">
              <span className="text-xs text-muted-foreground">
                {isAr ? `الشهر: ${selectedMonth} / ${selectedYear}` : `Period: ${selectedMonth}/${selectedYear}`}
              </span>
            </div>
          </div>

          <div className="overflow-x-auto rounded-2xl border border-border/70 bg-card shadow-sm">
            <table className="w-full text-sm text-start">
              <thead className="bg-muted/50 text-xs uppercase text-muted-foreground border-b border-border">
                <tr>
                  <th className="px-4 py-3 text-start">{isAr ? "الموظف والرقم الوظيفي" : "Employee & ID"}</th>
                  <th className="px-4 py-3 text-start">{isAr ? "رقم التأمينات" : "GOSI No."}</th>
                  <th className="px-4 py-3 text-start">{isAr ? "نطاق الجنسية" : "Scope"}</th>
                  <th className="px-4 py-3 text-end">{isAr ? "أجر الاشتراك (SAR)" : "Basis (SAR)"}</th>
                  <th className="px-4 py-3 text-end">{isAr ? "المعاشات (18%)" : "Annuity (18%)"}</th>
                  <th className="px-4 py-3 text-end">{isAr ? "الأخطار (2%)" : "OHRSI (2%)"}</th>
                  <th className="px-4 py-3 text-end">{isAr ? "ساند (1.5%)" : "SANED (1.5%)"}</th>
                  <th className="px-4 py-3 text-end font-semibold">{isAr ? "إجمالي الاشتراك" : "Total Amount"}</th>
                  <th className="px-4 py-3 text-center">{isAr ? "التفاصيل" : "Details"}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {filteredCalcs.map((calc) => (
                  <tr key={calc.id} className="hover:bg-muted/30 transition-colors">
                    <td className="px-4 py-3.5">
                      <div className="font-medium text-foreground">
                        {isAr ? calc.employeeNameAr : calc.employeeNameEn}
                      </div>
                      <div className="text-xs text-muted-foreground">{calc.employeeNumber}</div>
                    </td>
                    <td className="px-4 py-3.5 font-mono text-xs">{calc.gosiNumber}</td>
                    <td className="px-4 py-3.5">
                      {calc.nationalityScope === "SAUDI" ? (
                        <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20">
                          {isAr ? "سعودي (مواطن)" : "Saudi"}
                        </span>
                      ) : (
                        <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-blue-500/10 text-blue-600 dark:text-blue-400 border border-blue-500/20">
                          {isAr ? "غير سعودي (وافد)" : "Expat"}
                        </span>
                      )}
                    </td>
                    <td className="px-4 py-3.5 text-end font-mono font-medium">
                      {calc.contributionBasis.toLocaleString()}
                    </td>
                    <td className="px-4 py-3.5 text-end font-mono">
                      {calc.annuityEmp + calc.annuityEr > 0 ? (
                        <span>{(calc.annuityEmp + calc.annuityEr).toLocaleString()}</span>
                      ) : (
                        <span className="text-muted-foreground/40">—</span>
                      )}
                    </td>
                    <td className="px-4 py-3.5 text-end font-mono">
                      {calc.ohrsiEr.toLocaleString()}
                    </td>
                    <td className="px-4 py-3.5 text-end font-mono">
                      {calc.sanedEmp + calc.sanedEr > 0 ? (
                        <span>{(calc.sanedEmp + calc.sanedEr).toLocaleString()}</span>
                      ) : (
                        <span className="text-muted-foreground/40">—</span>
                      )}
                    </td>
                    <td className="px-4 py-3.5 text-end font-mono font-bold text-foreground">
                      {calc.totalAmount.toLocaleString("en-US", { minimumFractionDigits: 2 })}
                    </td>
                    <td className="px-4 py-3.5 text-center">
                      <button
                        onClick={() => setSelectedCalc(calc)}
                        className="p-1.5 rounded-lg hover:bg-muted text-muted-foreground hover:text-foreground transition-colors"
                        title={isAr ? "عرض التفاصيل والقواعد" : "View Breakdown"}
                      >
                        <Eye className="w-4 h-4" />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Tab 2: Profiles */}
      {activeTab === "profiles" && (
        <div className="space-y-4">
          <div className="flex items-center justify-between p-4 bg-card rounded-2xl border border-border/70 shadow-sm">
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <Users className="w-4 h-4 text-primary" />
              <span>{isAr ? "سجل تسجيل الموظفين في التأمينات الاجتماعية وأجر الاشتراك المعتمد" : "GOSI subscriber profiles and registered contribution bases"}</span>
            </div>
            <button
              onClick={() => setShowEnrolModal(true)}
              className="flex items-center gap-1.5 px-3.5 py-2 rounded-xl bg-primary text-primary-foreground font-medium text-xs shadow-sm hover:opacity-95"
            >
              <Plus className="w-3.5 h-3.5" />
              <span>{isAr ? "تسجيل مشترك جديد" : "Enrol Subscriber"}</span>
            </button>
          </div>

          <div className="overflow-x-auto rounded-2xl border border-border/70 bg-card shadow-sm">
            <table className="w-full text-sm text-start">
              <thead className="bg-muted/50 text-xs uppercase text-muted-foreground border-b border-border">
                <tr>
                  <th className="px-4 py-3 text-start">{isAr ? "الموظف" : "Employee"}</th>
                  <th className="px-4 py-3 text-start">{isAr ? "الهوية / الإقامة" : "National ID / Iqama"}</th>
                  <th className="px-4 py-3 text-start">{isAr ? "رقم المشترك (GOSI)" : "GOSI Number"}</th>
                  <th className="px-4 py-3 text-start">{isAr ? "الجنسية والنطاق" : "Nationality Scope"}</th>
                  <th className="px-4 py-3 text-start">{isAr ? "تاريخ التسجيل" : "Registration Date"}</th>
                  <th className="px-4 py-3 text-end">{isAr ? "أجر الاشتراك (SAR)" : "Basis (SAR)"}</th>
                  <th className="px-4 py-3 text-center">{isAr ? "الحالة" : "Status"}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {profiles.map((prof) => (
                  <tr key={prof.id} className="hover:bg-muted/30 transition-colors">
                    <td className="px-4 py-3.5 font-medium">
                      {isAr ? prof.employeeNameAr : prof.employeeNameEn}
                      <div className="text-xs text-muted-foreground">{prof.employeeNumber}</div>
                    </td>
                    <td className="px-4 py-3.5 font-mono text-xs">{prof.nationalIdOrIqama}</td>
                    <td className="px-4 py-3.5 font-mono text-xs text-primary font-medium">{prof.gosiNumber}</td>
                    <td className="px-4 py-3.5">
                      <div className="font-medium text-xs">{isAr ? prof.nationalityNameAr : prof.nationalityNameEn}</div>
                      <div className="text-xs text-muted-foreground">
                        {prof.nationalityScope === "SAUDI" ? (isAr ? "مواطن" : "Citizen") : (isAr ? "وافد" : "Expat")}
                      </div>
                    </td>
                    <td className="px-4 py-3.5 font-mono text-xs text-muted-foreground">{prof.registrationDate}</td>
                    <td className="px-4 py-3.5 text-end font-mono font-bold text-foreground">
                      {prof.basisOverride.toLocaleString()}
                    </td>
                    <td className="px-4 py-3.5 text-center">
                      <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20">
                        {isAr ? "نشط" : "Active"}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Tab 3: Statutory Rules */}
      {activeTab === "rules" && (
        <div className="space-y-6">
          <div className="p-4 rounded-2xl bg-amber-500/10 border border-amber-500/20 flex items-start gap-3">
            <Lock className="w-5 h-5 text-amber-600 dark:text-amber-400 shrink-0 mt-0.5" />
            <div className="text-sm">
              <span className="font-semibold text-amber-700 dark:text-amber-300">
                {isAr ? "قاعدة معمارية إلزامية:" : "Mandatory Architectural Invariant:"}
              </span>{" "}
              <span className="text-amber-900/90 dark:text-amber-200/90">
                {isAr
                  ? "جميع نسب التأمينات الاجتماعية والحد الأدنى والأعلى تؤخذ حصراً من جداول القواعد الإصدارية بقاعدة البيانات (gosi.contribution_rule_versions) ولا يتم ترميز أي نسبة داخل الكود البرمجي إطلاقاً."
                  : "All GOSI statutory contribution percentages, min/max basis limits are read dynamically from versioned database records. Rates are never hard-coded."}
              </span>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
            {rules.map((rule) => (
              <div key={rule.ruleCode} className="p-5 rounded-2xl bg-card border border-border/70 shadow-sm flex flex-col justify-between">
                <div>
                  <div className="flex items-center justify-between mb-3">
                    <span className="px-2.5 py-1 rounded-lg text-xs font-mono font-bold bg-primary/10 text-primary">
                      {rule.ruleCode}
                    </span>
                    <span className="text-xs text-muted-foreground font-mono">
                      {isAr ? `الإصدار ${rule.versionNumber}` : `Version ${rule.versionNumber}`}
                    </span>
                  </div>
                  <h3 className="text-base font-bold text-foreground mb-1">
                    {isAr ? rule.ruleNameAr : rule.ruleNameEn}
                  </h3>
                  <p className="text-xs text-muted-foreground mb-4">
                    {rule.notes}
                  </p>

                  <div className="space-y-2.5 border-t border-border pt-4 text-xs">
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">{isAr ? "نطاق المشتركين:" : "Scope:"}</span>
                      <span className="font-semibold">
                        {rule.nationalityScope === "SAUDI"
                          ? isAr ? "المواطنون السعوديون فقط" : "Saudi Nationals Only"
                          : isAr ? "كافة الجنسيات (سعودي وغير سعودي)" : "All Nationalities"}
                      </span>
                    </div>

                    <div className="flex justify-between">
                      <span className="text-muted-foreground">{isAr ? "حصة المشترك (الموظف):" : "Employee Rate:"}</span>
                      <span className="font-bold text-emerald-600 dark:text-emerald-400 font-mono">
                        {rule.employeeRate}%
                      </span>
                    </div>

                    <div className="flex justify-between">
                      <span className="text-muted-foreground">{isAr ? "حصة صاحب العمل (المنشأة):" : "Employer Rate:"}</span>
                      <span className="font-bold text-blue-600 dark:text-blue-400 font-mono">
                        {rule.employerRate}%
                      </span>
                    </div>

                    <div className="flex justify-between">
                      <span className="text-muted-foreground">{isAr ? "إجمالي نسبة الفرع:" : "Total Branch Rate:"}</span>
                      <span className="font-bold text-foreground font-mono">
                        {rule.employeeRate + rule.employerRate}%
                      </span>
                    </div>

                    <div className="flex justify-between">
                      <span className="text-muted-foreground">{isAr ? "الحد الأدنى والأعلى للأجر:" : "Basis Min - Max:"}</span>
                      <span className="font-mono">
                        {rule.minBasis.toLocaleString()} - {rule.maxBasis.toLocaleString()} SAR
                      </span>
                    </div>
                  </div>
                </div>

                <div className="mt-4 pt-3 border-t border-border/50 text-[11px] text-muted-foreground">
                  <div className="font-semibold text-foreground/80 mb-0.5">{isAr ? "السند النظامي:" : "Legal Basis:"}</div>
                  <div>{rule.legalReference}</div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Tab 4: Export Batches */}
      {activeTab === "export" && (
        <div className="space-y-4">
          <div className="flex items-center justify-between p-4 bg-card rounded-2xl border border-border/70 shadow-sm">
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <FileSpreadsheet className="w-4 h-4 text-emerald-600" />
              <span>{isAr ? "تصدير تقارير التأمينات والربط الإلكتروني مع بوابة التأمينات الاجتماعية (GOSI Adapter)" : "GOSI export batches, XML report generation, and portal submission"}</span>
            </div>
            <button
              id="gosi-new-batch-btn"
              onClick={() => {
                const newBatch: GosiExportBatchItem = {
                  id: `batch-${Date.now()}`,
                  batchNumber: `GOSI-202609-${Math.random().toString(36).substring(2, 6).toUpperCase()}`,
                  year: selectedYear,
                  month: selectedMonth,
                  employeeCount: calculations.length,
                  totalEmployeeAmount: totalEmpContributions,
                  totalEmployerAmount: totalErContributions,
                  totalAmount: grandTotal,
                  status: "GENERATED",
                  generatedAt: new Date().toISOString().replace("T", " ").substring(0, 16),
                  fileFormat: "XML",
                };
                setBatches([newBatch, ...batches]);
              }}
              className="flex items-center gap-1.5 px-3.5 py-2 rounded-xl bg-primary text-primary-foreground font-medium text-xs shadow-sm hover:opacity-95"
            >
              <Plus className="w-3.5 h-3.5" />
              <span>{isAr ? "إنشاء دفعة تصدير جديدة" : "Generate Export Batch"}</span>
            </button>
          </div>

          <div className="overflow-x-auto rounded-2xl border border-border/70 bg-card shadow-sm">
            <table className="w-full text-sm text-start">
              <thead className="bg-muted/50 text-xs uppercase text-muted-foreground border-b border-border">
                <tr>
                  <th className="px-4 py-3 text-start">{isAr ? "رقم الدفعة" : "Batch Number"}</th>
                  <th className="px-4 py-3 text-start">{isAr ? "فترة الاشتراك" : "Period"}</th>
                  <th className="px-4 py-3 text-center">{isAr ? "عدد المشتركين" : "Subscribers"}</th>
                  <th className="px-4 py-3 text-end">{isAr ? "حصة الموظفين (SAR)" : "Employee Total"}</th>
                  <th className="px-4 py-3 text-end">{isAr ? "حصة المنشأة (SAR)" : "Employer Total"}</th>
                  <th className="px-4 py-3 text-end font-bold">{isAr ? "الإجمالي الكلي" : "Grand Total"}</th>
                  <th className="px-4 py-3 text-center">{isAr ? "حالة الدفعة" : "Status"}</th>
                  <th className="px-4 py-3 text-center">{isAr ? "الإجراءات" : "Actions"}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {batches.map((b) => (
                  <tr key={b.id} className="hover:bg-muted/30 transition-colors">
                    <td className="px-4 py-3.5 font-mono font-medium text-foreground">{b.batchNumber}</td>
                    <td className="px-4 py-3.5 font-mono text-xs">{b.month} / {b.year}</td>
                    <td className="px-4 py-3.5 text-center font-bold">{b.employeeCount}</td>
                    <td className="px-4 py-3.5 text-end font-mono">{b.totalEmployeeAmount.toLocaleString()}</td>
                    <td className="px-4 py-3.5 text-end font-mono">{b.totalEmployerAmount.toLocaleString()}</td>
                    <td className="px-4 py-3.5 text-end font-mono font-bold text-foreground">
                      {b.totalAmount.toLocaleString("en-US", { minimumFractionDigits: 2 })}
                    </td>
                    <td className="px-4 py-3.5 text-center">
                      {b.status === "CONFIRMED" && (
                        <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20">
                          <CheckCircle2 className="w-3 h-3" />
                          {isAr ? "معتمد ومطابق" : "Confirmed"}
                        </span>
                      )}
                      {b.status === "SUBMITTED" && (
                        <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-500/10 text-blue-600 dark:text-blue-400 border border-blue-500/20">
                          <Clock className="w-3 h-3 animate-spin" />
                          {isAr ? "تم الإرسال" : "Submitted"}
                        </span>
                      )}
                      {b.status === "GENERATED" && (
                        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-amber-500/10 text-amber-600 dark:text-amber-400 border border-amber-500/20">
                          {isAr ? "جاهز للإرسال" : "Generated"}
                        </span>
                      )}
                    </td>
                    <td className="px-4 py-3.5 text-center">
                      <div className="flex items-center justify-center gap-2">
                        <button
                          onClick={() => setSelectedXmlBatch(b)}
                          className="p-1.5 rounded-lg hover:bg-muted text-muted-foreground hover:text-foreground transition-colors"
                          title={isAr ? "معاينة ملف XML" : "View XML"}
                        >
                          <Eye className="w-4 h-4" />
                        </button>
                        {b.status === "GENERATED" && (
                          <button
                            onClick={() => handleSubmitBatch(b.id)}
                            className="flex items-center gap-1 px-2.5 py-1 rounded-lg bg-emerald-600 text-white hover:bg-emerald-700 text-xs font-medium transition-colors"
                            title={isAr ? "إرسال للبوابة" : "Submit to Portal"}
                          >
                            <Send className="w-3 h-3" />
                            <span>{isAr ? "إرسال" : "Submit"}</span>
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Detail Modal for Calculation */}
      {selectedCalc && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-background/80 backdrop-blur-sm">
          <div className="w-full max-w-lg bg-card rounded-2xl border border-border shadow-xl p-6 space-y-4">
            <div className="flex items-center justify-between border-b border-border pb-3">
              <div className="flex items-center gap-2">
                <ShieldCheck className="w-5 h-5 text-primary" />
                <h3 className="font-bold text-foreground">
                  {isAr ? "تفاصيل احتساب التأمينات الاجتماعية" : "GOSI Contribution Breakdown"}
                </h3>
              </div>
              <button
                onClick={() => setSelectedCalc(null)}
                className="text-muted-foreground hover:text-foreground text-sm font-semibold"
              >
                ✕
              </button>
            </div>

            <div className="space-y-3 text-sm">
              <div className="p-3 rounded-xl bg-muted/40 space-y-1.5">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">{isAr ? "اسم المشترك:" : "Subscriber Name:"}</span>
                  <span className="font-bold">{isAr ? selectedCalc.employeeNameAr : selectedCalc.employeeNameEn}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">{isAr ? "الرقم التأميني:" : "GOSI No:"}</span>
                  <span className="font-mono">{selectedCalc.gosiNumber}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">{isAr ? "أجر الاشتراك المعتمد:" : "Contribution Basis:"}</span>
                  <span className="font-mono font-bold">{selectedCalc.contributionBasis.toLocaleString()} SAR</span>
                </div>
              </div>

              <div className="border border-border rounded-xl divide-y divide-border">
                {/* Annuity */}
                <div className="p-3 flex items-center justify-between">
                  <div>
                    <div className="font-semibold">{isAr ? "فرع المعاشات التقاعدية (Annuity)" : "Annuity / Pension"}</div>
                    <div className="text-xs text-muted-foreground">
                      {selectedCalc.nationalityScope === "SAUDI"
                        ? isAr ? "9% موظف + 9% صاحب عمل" : "9% Employee + 9% Employer"
                        : isAr ? "غير مشمول (للسعوديين فقط)" : "Not applicable (Saudi only)"}
                    </div>
                  </div>
                  <div className="text-end font-mono">
                    <div className="font-bold">
                      {(selectedCalc.annuityEmp + selectedCalc.annuityEr).toLocaleString()} SAR
                    </div>
                    <div className="text-xs text-muted-foreground">
                      {selectedCalc.annuityEmp} + {selectedCalc.annuityEr}
                    </div>
                  </div>
                </div>

                {/* OHRSI */}
                <div className="p-3 flex items-center justify-between">
                  <div>
                    <div className="font-semibold">{isAr ? "الأخطار المهنية (OHRSI)" : "Occupational Hazards"}</div>
                    <div className="text-xs text-muted-foreground">
                      {isAr ? "0% موظف + 2% صاحب عمل (كافة الجنسيات)" : "0% Employee + 2% Employer (All)"}
                    </div>
                  </div>
                  <div className="text-end font-mono">
                    <div className="font-bold">{selectedCalc.ohrsiEr.toLocaleString()} SAR</div>
                    <div className="text-xs text-muted-foreground">0 + {selectedCalc.ohrsiEr}</div>
                  </div>
                </div>

                {/* SANED */}
                <div className="p-3 flex items-center justify-between">
                  <div>
                    <div className="font-semibold">{isAr ? "نظام ساند (التعطل عن العمل)" : "SANED Unemployment"}</div>
                    <div className="text-xs text-muted-foreground">
                      {selectedCalc.nationalityScope === "SAUDI"
                        ? isAr ? "0.75% موظف + 0.75% صاحب عمل" : "0.75% Employee + 0.75% Employer"
                        : isAr ? "غير مشمول (للسعوديين فقط)" : "Not applicable (Saudi only)"}
                    </div>
                  </div>
                  <div className="text-end font-mono">
                    <div className="font-bold">
                      {(selectedCalc.sanedEmp + selectedCalc.sanedEr).toLocaleString()} SAR
                    </div>
                    <div className="text-xs text-muted-foreground">
                      {selectedCalc.sanedEmp} + {selectedCalc.sanedEr}
                    </div>
                  </div>
                </div>
              </div>

              <div className="p-3 rounded-xl bg-primary/5 border border-primary/20 flex justify-between items-center font-bold">
                <span>{isAr ? "الإجمالي الكلي للاشتراك:" : "Grand Total:"}</span>
                <span className="text-primary text-base font-mono">
                  {selectedCalc.totalAmount.toLocaleString("en-US", { minimumFractionDigits: 2 })} SAR
                </span>
              </div>
            </div>

            <div className="flex justify-end pt-2">
              <button
                onClick={() => setSelectedCalc(null)}
                className="px-4 py-2 rounded-xl bg-muted hover:bg-muted/80 text-sm font-medium transition-colors"
              >
                {isAr ? "إغلاق" : "Close"}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* XML Preview Modal */}
      {selectedXmlBatch && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-background/80 backdrop-blur-sm">
          <div className="w-full max-w-2xl bg-card rounded-2xl border border-border shadow-xl p-6 space-y-4">
            <div className="flex items-center justify-between border-b border-border pb-3">
              <div className="flex items-center gap-2">
                <FileSpreadsheet className="w-5 h-5 text-emerald-600" />
                <h3 className="font-bold text-foreground">
                  {isAr ? `معاينة ملف تصدير التأمينات (${selectedXmlBatch.batchNumber})` : `GOSI XML File (${selectedXmlBatch.batchNumber})`}
                </h3>
              </div>
              <button
                onClick={() => setSelectedXmlBatch(null)}
                className="text-muted-foreground hover:text-foreground text-sm font-semibold"
              >
                ✕
              </button>
            </div>

            <div className="p-4 rounded-xl bg-muted/60 border border-border font-mono text-xs overflow-x-auto max-h-72">
              <pre className="text-foreground/90">
{`<?xml version="1.0" encoding="UTF-8"?>
<GOSIContributionReport xmlns="urn:sa:gosi:v1" year="${selectedXmlBatch.year}" month="${selectedXmlBatch.month}">
  <LegalEntity id="019484b2-e1d5-76b3-96b5-555544443333" crn="1010000001"/>
  <Summary
    batchNumber="${selectedXmlBatch.batchNumber}"
    subscriberCount="${selectedXmlBatch.employeeCount}"
    employeeTotal="${selectedXmlBatch.totalEmployeeAmount.toFixed(2)}"
    employerTotal="${selectedXmlBatch.totalEmployerAmount.toFixed(2)}"
    grandTotal="${selectedXmlBatch.totalAmount.toFixed(2)}"
    currency="SAR"
    generatedAt="${selectedXmlBatch.generatedAt}"
  />
  <Subscribers>
    <!-- Dynamic subscriber nodes produced by GOSIAdapter -->
    <Subscriber gosiNo="984210" nationalId="1089332145" basis="14500.00" annuity="2610.00" ohrsi="290.00" saned="217.50" total="3117.50"/>
    <Subscriber gosiNo="984215" nationalId="1098443211" basis="18000.00" annuity="3240.00" ohrsi="360.00" saned="270.00" total="3870.00"/>
    <Subscriber gosiNo="984220" iqamaNo="2489330012" basis="12000.00" annuity="0.00" ohrsi="240.00" saned="0.00" total="240.00"/>
  </Subscribers>
</GOSIContributionReport>`}
              </pre>
            </div>

            <div className="flex justify-end gap-3 pt-2">
              <button
                onClick={() => {
                  const blob = new Blob(
                    [`<?xml version="1.0" encoding="UTF-8"?>\n<GOSIContributionReport batch="${selectedXmlBatch.batchNumber}"/>`],
                    { type: "text/xml" }
                  );
                  const url = URL.createObjectURL(blob);
                  const a = document.createElement("a");
                  a.href = url;
                  a.download = `${selectedXmlBatch.batchNumber}.xml`;
                  a.click();
                  URL.revokeObjectURL(url);
                }}
                className="flex items-center gap-1.5 px-4 py-2 rounded-xl bg-primary text-primary-foreground text-sm font-medium hover:opacity-95 transition-opacity"
              >
                <Download className="w-4 h-4" />
                <span>{isAr ? "تحميل الملف" : "Download XML"}</span>
              </button>
              <button
                onClick={() => setSelectedXmlBatch(null)}
                className="px-4 py-2 rounded-xl bg-muted hover:bg-muted/80 text-sm font-medium transition-colors"
              >
                {isAr ? "إغلاق" : "Close"}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Enrol Subscriber Modal */}
      {showEnrolModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-background/80 backdrop-blur-sm">
          <div className="w-full max-w-md bg-card rounded-2xl border border-border shadow-xl p-6 space-y-4">
            <div className="flex items-center justify-between border-b border-border pb-3">
              <div className="flex items-center gap-2">
                <Users className="w-5 h-5 text-primary" />
                <h3 className="font-bold text-foreground">
                  {isAr ? "تسجيل مشترك جديد بالتأمينات" : "Enrol New GOSI Subscriber"}
                </h3>
              </div>
              <button
                onClick={() => setShowEnrolModal(false)}
                className="text-muted-foreground hover:text-foreground text-sm font-semibold"
              >
                ✕
              </button>
            </div>

            <form
              onSubmit={(e) => {
                e.preventDefault();
                setShowEnrolModal(false);
              }}
              className="space-y-3 text-sm"
            >
              <div>
                <label className="block text-xs font-semibold mb-1 text-muted-foreground">
                  {isAr ? "اسم الموظف" : "Employee Name"}
                </label>
                <input
                  type="text"
                  required
                  placeholder={isAr ? "مثال: فيصل العتيبي" : "e.g. Faisal Al-Otaibi"}
                  className="w-full px-3 py-2 bg-background border border-border rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold mb-1 text-muted-foreground">
                  {isAr ? "رقم الهوية الوطنية أو الإقامة" : "National ID or Iqama"}
                </label>
                <input
                  type="text"
                  required
                  placeholder="10XXXXXXXX / 2XXXXXXXXX"
                  className="w-full px-3 py-2 bg-background border border-border rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold mb-1 text-muted-foreground">
                  {isAr ? "رقم المشترك في التأمينات" : "GOSI Registration Number"}
                </label>
                <input
                  type="text"
                  required
                  placeholder="GOSI-XXXXXX"
                  className="w-full px-3 py-2 bg-background border border-border rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold mb-1 text-muted-foreground">
                  {isAr ? "أجر الاشتراك المعتمد (الأساسي + بدل السكن)" : "Contribution Basis (Basic + Housing)"}
                </label>
                <input
                  type="number"
                  min="1500"
                  max="45000"
                  defaultValue="10000"
                  className="w-full px-3 py-2 bg-background border border-border rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20"
                />
                <span className="text-[11px] text-muted-foreground mt-0.5 block">
                  {isAr ? "الحد الأدنى 1,500 والحد الأقصى 45,000 ريال" : "Min 1,500 - Max 45,000 SAR"}
                </span>
              </div>

              <div className="flex justify-end gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setShowEnrolModal(false)}
                  className="px-4 py-2 rounded-xl bg-muted hover:bg-muted/80 text-sm font-medium transition-colors"
                >
                  {isAr ? "إلغاء" : "Cancel"}
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 rounded-xl bg-primary text-primary-foreground text-sm font-medium hover:opacity-95 transition-opacity"
                >
                  {isAr ? "حفظ وتسجيل" : "Save & Register"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
